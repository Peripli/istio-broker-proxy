#!/bin/bash
set -ueo pipefail

go build github.infra.hana.ondemand.com/istio/istio-broker

./istio-broker --providerId="x" &
ISTIOPID=$!
trap "kill $ISTIOPID" 0
sleep 1

CMD="curl -s -i -d @invalidRequest.json -X PUT http://localhost:8080/v2/service_instances/1/service_bindings/2/adapt_credentials"
echo $CMD
if $CMD | grep "HTTP/1.1 400"
then
        echo "*** OK ***"
else
        echo "*** FAIL ***"
        exit 1
fi  

CMD="curl -s -i -d @exampleRequest.json -X PUT http://localhost:8080/v2/service_instances/1/service_bindings/2/adapt_credentials"
echo $CMD
if $CMD  | grep "HTTP/1.1 200"
then
        echo "*** OK ***"
else
        echo "*** FAIL ***"
        exit 1
fi  
