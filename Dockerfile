FROM gcr.io/sap-se-gcp-istio-dev/client

RUN mkdir -p /app
ADD istio-broker /app/istio-broker

ENTRYPOINT [ "/app/istio-broker" ]
