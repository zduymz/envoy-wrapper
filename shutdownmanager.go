package main

import (
	"fmt"
	"io"
	"net/http"
	"sync"
	"time"
	"github.com/prometheus/common/expfmt"
	log "github.com/sirupsen/logrus"
)

const (
	prometheusURL       = "http://localhost:19000/stats/prometheus"
	healthcheckFailURL  = "http://localhost:19000/healthcheck/fail"
	drainingListenerURL = "http://localhost:19000//drain_listeners"
	prometheusStat      = "envoy_http_downstream_cx_active"
)

func prometheusLabels() []string {
	return []string{"public_listener_http"}
}

type SMContext struct {
	// checkInterval defines time delay between polling Envoy for open connections
	checkInterval time.Duration

	// checkDelay defines time to wait before polling Envoy for open connections
	checkDelay time.Duration

	// minOpenConnections defines the minimum amount of connections
	// that can be open when polling for active connections in Envoy
	minOpenConnections int

	// httpServePort defines what port the shutdown-manager listens on
	httpServePort int

	// preStop and SIGINT, whatever call first, go first
	lock sync.Mutex
}

func newSMContext() *SMContext {
	return &SMContext{
		checkInterval:      5 * time.Second,
		checkDelay:         30 * time.Second,
		minOpenConnections: 0,
		httpServePort:      8090,
	}
}

func (s *SMContext) shutdownHandler(w http.ResponseWriter, r *http.Request) {
	// Send shutdown signal to Envoy to start draining connections
	s.lock.Lock()
	defer s.lock.Unlock()
	log.Infof("shutdown envoy gracefully")
	if err := shutdownEnvoy(); err != nil {
		log.Error(err)
	}

	if err := drainListenersEnvoy(); err != nil {
		log.Error(err)
	}

	log.Infof("waiting %s before polling for draining connections", s.checkDelay)
	time.Sleep(s.checkDelay)

	for {
		openConnections, err := getOpenConnections()
		if err != nil {
			log.Error(err)
		} else {
			if openConnections <= s.minOpenConnections {
				log.WithField("open_connections", openConnections).
					WithField("min_connections", s.minOpenConnections).
					Info("min number of open connections found, shutting down")
				return
			}
			log.WithField("open_connections", openConnections).
				WithField("min_connections", s.minOpenConnections).
				Info("polled open connections")
		}
		time.Sleep(s.checkInterval)
	}
}

// shutdownEnvoy sends a POST request to /healthcheck/fail to tell Envoy to start draining connections
func shutdownEnvoy() error {
	resp, err := http.Post(healthcheckFailURL, "", nil)
	if err != nil {
		return fmt.Errorf("creating healthcheck/fail POST request failed: %s", err)
	}

	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("POST for %q returned HTTP status %s", healthcheckFailURL, resp.Status)
	}
	return nil
}

func drainListenersEnvoy() error {
	resp, err := http.Post(drainingListenerURL, "", nil)
	if err != nil {
		return fmt.Errorf("creating healthcheck/fail POST request failed: %s", err)
	}

	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("POST for %q returned HTTP status %s", drainingListenerURL, resp.Status)
	}
	return nil
}

// getOpenConnections parses a http request to a prometheus endpoint returning the sum of values found
func getOpenConnections() (int, error) {
	// Make request to Envoy Prometheus endpoint
	resp, err := http.Get(prometheusURL)
	if err != nil {
		return -1, fmt.Errorf("creating metrics GET request failed: %s", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return -1, fmt.Errorf("GET for %q returned HTTP status %s", prometheusURL, resp.Status)
	}

	// Parse Prometheus listener stats for open connections
	return parseOpenConnections(resp.Body)
}

// parseOpenConnections returns the sum of open connections from a Prometheus HTTP request
func parseOpenConnections(stats io.Reader) (int, error) {
	var parser expfmt.TextParser
	openConnections := 0

	if stats == nil {
		return -1, fmt.Errorf("stats input was nil")
	}

	// Parse Prometheus http response
	metricFamilies, err := parser.TextToMetricFamilies(stats)
	if err != nil {
		return -1, fmt.Errorf("parsing Prometheus text format failed: %v", err)
	}

	// Validate stat exists in output
	if _, ok := metricFamilies[prometheusStat]; !ok {
		return -1, fmt.Errorf("error finding Prometheus stat [%q] in the request result", prometheusStat)
	}

	// Look up open connections value
	for _, metrics := range metricFamilies[prometheusStat].Metric {
		for _, labels := range metrics.Label {
			for _, item := range prometheusLabels() {
				if item == labels.GetValue() {
					openConnections += int(metrics.Gauge.GetValue())
				}
			}
		}
	}
	return openConnections, nil
}

func doShutdownManager(config *SMContext) {
	log.Info("started envoy shutdown manager")
	defer log.Info("stopped")

	http.HandleFunc("/shutdown", config.shutdownHandler)
	log.Fatal(http.ListenAndServe(fmt.Sprintf(":%d", config.httpServePort), nil))
}
