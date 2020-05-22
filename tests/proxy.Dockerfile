FROM envoyproxy/envoy-dev:latest

RUN apt-get update && apt-get -q install -y curl
COPY proxy-envoy.yaml /etc/proxy-envoy.yaml

ENTRYPOINT ["/usr/local/bin/envoy", "-c", "/etc/proxy-envoy.yaml"] 
