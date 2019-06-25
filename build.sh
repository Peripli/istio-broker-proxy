#!/bin/bash -x

export GO111MODULE=on
set -euo pipefail

SCRIPT_DIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" >/dev/null && pwd )"

export CGO_ENABLED=0

cd ${SCRIPT_DIR}

golint  -set_exit_status ./pkg/...
golint  -set_exit_status main.go

#go get github.com/kubernetes-incubator/service-catalog/pkg/apis/servicecatalog/v1beta1
go build -mod=vendor ./pkg/config/client
./client --help
VERSION=$(git rev-parse HEAD)
go build -mod=vendor --ldflags="-X main.version=$VERSION"
