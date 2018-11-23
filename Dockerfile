FROM gcr.io/sap-se-gcp-istio-dev/client

RUN mkdir -p /app
ADD istio-broker-proxy /app/istio-broker-proxy

ENTRYPOINT [ "/app/istio-broker-proxy" ]
