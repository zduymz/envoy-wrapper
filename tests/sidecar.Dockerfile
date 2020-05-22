FROM envoyproxy/envoy-dev:latest

RUN apt-get update && apt-get -q install -y curl
COPY service-envoy.yaml /etc/service-envoy.yaml
COPY envoy-wrapper /bin/envoy-wrapper

ENTRYPOINT ["/bin/envoy-wrapper", "/usr/local/bin/envoy", "-c",  "/etc/service-envoy.yaml"] 
