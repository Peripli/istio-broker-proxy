#!/bin/bash
go build github.infra.hana.ondemand.com/istio/istio-broker
./istio-broker &
curl -i -d @invalidRequest.json -X POST http://localhost:8080/update_credentials
curl -i -d @exampleRequest.json -X POST http://localhost:8080/update_credentials
killall istio-broker 2> /dev/null
