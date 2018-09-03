FROM gcr.io/sap-se-gcp-istio-dev/client

RUN apt-get update && apt-get -y install golang

WORKDIR /go/src/github.infra.hana.ondemand.com/istio/istio-broker
ADD . .
ENV GOPATH /go
RUN go version
RUN GOOS=linux go build -v  -o istio-broker
RUN mkdir -p /app && cp istio-broker /app

ENTRYPOINT [ "/app/istio-broker" ]
