#!/bin/bash
set -ueo pipefail
go build github.infra.hana.ondemand.com/istio/istio-broker
./istio-broker &

RC=0

CMD="curl -s -i -d @invalidRequest.json -X POST http://localhost:8080/update_credentials" 
echo $CMD
if $CMD | grep "HTTP/1.1 400"
then
        echo "*** OK ***"
else
        echo "*** FAIL ***"
        RC=1
fi  

CMD="curl -s -i -d @exampleRequest.json -X POST http://localhost:8080/update_credentials"
echo $CMD
if $CMD  | grep "HTTP/1.1 200"
then
        echo "*** OK ***"
else
        echo "*** FAIL ***"
        RC=1
fi  

killall istio-broker 2> /dev/null

exit $RC
