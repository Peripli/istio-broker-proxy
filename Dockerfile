FROM scratch

ADD istio-broker-proxy /app/istio-broker-proxy

ENTRYPOINT [ "/app/istio-broker-proxy" ]
