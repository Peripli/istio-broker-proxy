#!/bin/bash

set -euo pipefail

export GOPATH=$GOPATH:$(pwd)/go

go get -d -v github.infra.hana.ondemand.com/istio/istio-broker/...
go build github.infra.hana.ondemand.com/istio/istio-broker/pkg/config/client
./client --help
go build -v github.infra.hana.ondemand.com/istio/istio-broker
( cd go/src/github.infra.hana.ondemand.com/istio/istio-broker/test && ./test_update_credentials.sh )
