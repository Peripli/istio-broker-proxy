#!/bin/bash

set -euo pipefail

SCRIPT_DIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" >/dev/null && pwd )"

export GOPATH=${SCRIPT_DIR}/../../../..

cd ${SCRIPT_DIR}

#go get github.com/kubernetes-incubator/service-catalog/pkg/apis/servicecatalog/v1beta1
go build ./pkg/config/client
./client --help
go build -v
