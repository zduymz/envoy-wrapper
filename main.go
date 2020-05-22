package main

import (
	"os"
	"os/exec"
	"os/signal"
	"syscall"

	log "github.com/sirupsen/logrus"
	prefixed "github.com/x-cray/logrus-prefixed-formatter"
)

func init() {
	log.SetFormatter(&prefixed.TextFormatter{
		DisableColors:   true,
		TimestampFormat: "2006-01-02 15:04:05",
		FullTimestamp:   true,
		ForceFormatting: true,
	})
}

func main() {
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM, syscall.SIGHUP, syscall.SIGQUIT)

	if len(os.Args) < 2 {
		log.Errorf("Usage: %s  [envoy] [args]\n", os.Args[0])
		os.Exit(1)
	}

	sm := newSMContext()
	go doShutdownManager(sm)

	cmd := exec.Command(os.Args[1], os.Args[2:]...)

	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.SysProcAttr = &syscall.SysProcAttr{}
	cmd.SysProcAttr.Setpgid = true

	if err := cmd.Start(); err != nil {
		log.Error(err)
	}

	waitCh := make(chan error, 1)
	go func() {
		waitCh <- cmd.Wait()
		close(waitCh)
	}()

	for {
		select {
		case sig := <-sigChan:
			log.Info("Envoy wrapper received signal: ", sig)
			sm.shutdownHandler(nil, nil)
			log.Info("Envoy wrapper forward signal: ", sig)
			if err := cmd.Process.Signal(sig); err != nil {
				log.Error("Unable to forward signal: ", err)
			}
			log.Info("Envoy shutdown completely")
			return
		case err := <-waitCh:
			if exitError, ok := err.(*exec.ExitError); ok {
				waitStatus := exitError.Sys().(syscall.WaitStatus)
				os.Exit(waitStatus.ExitStatus())
			}
			if err != nil {
				log.Fatal(err)
			}
			return
		}
	}

}
