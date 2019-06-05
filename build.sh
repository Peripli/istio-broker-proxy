#!/bin/bash -x

set -euo pipefail

SCRIPT_DIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" >/dev/null && pwd )"

export GOPATH=${SCRIPT_DIR}/../../../..
export CGO_ENABLED=0

cd ${SCRIPT_DIR}

golint  -set_exit_status ./pkg/...
golint  -set_exit_status main.go

#go get github.com/kubernetes-incubator/service-catalog/pkg/apis/servicecatalog/v1beta1
go build ./pkg/config/client
./client --help
BROKER_VERSION=$(git rev-parse HEAD)
go build --ldflags="-X main.commitHash=$BROKER_VERSION"
