FROM ubuntu:18.04

RUN mkdir -p /app
ADD istio-broker-proxy /app/istio-broker-proxy

ENTRYPOINT [ "/app/istio-broker-proxy" ]
