#!/bin/bash

docker run -d --name envoy -p 51051:51051 -p 9901:9901 \
-v `pwd`/hotels/envoy-proxy.yaml:/etc/hotels-proxy.yaml \
-v `pwd`/hotels/hotels.pb:/tmp/envoy/hotels.pb \
envoyproxy/envoy /usr/local/bin/envoy -c /etc/hotels-proxy.yaml --service-cluster hotels-proxy
