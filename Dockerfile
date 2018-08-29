FROM ubuntu:16.04 as build-env

RUN apt-get update
RUN apt-get -y install golang

WORKDIR /go/src/github.infra.hana.ondemand.com/istio/istio-broker
ADD . .
ENV GOPATH /go
RUN GOOS=linux go build -v  -o istio-broker
RUN ls -la istio-broker
RUN mkdir -p /app && cp istio-broker /app

ENTRYPOINT [ "/app/istio-broker" ]
