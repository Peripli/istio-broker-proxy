#!/bin/bash

set -euo pipefail

SCRIPT_DIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" >/dev/null && pwd )"

export GOPATH=${SCRIPT_DIR}/../../../../..
export CGO_ENABLED=0

cd ${SCRIPT_DIR}

go get github.com/kubernetes-incubator/service-catalog/pkg/apis/servicecatalog/v1beta1
#go get -tags integration -v -d ../...
go test -tags integration -v
