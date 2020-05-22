#!/bin/bash

echo "make 100 request in service"
for i in {1..100}; do
    curl -s http://localhost:8080/$i > /dev/null &
done

echo "call envoy-wrapper shutdown"
curl -s http://localhost:8090/shutdown &
# docker stop -t 300 testing_web-server-sidecar_1s

echo "continue sending requesets, should thrown error"
for i in {1..10}; do
   curl -s http://localhost:8080/120
done
