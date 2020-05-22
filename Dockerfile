ARG VERSION
FROM envoyproxy/envoy-alpine:${VERSION}

COPY ./build/linux/envoy-wrapper /bin/envoy-wrapper