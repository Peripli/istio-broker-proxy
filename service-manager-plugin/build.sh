#!/bin/bash

set -euox pipefail

SCRIPT_DIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" >/dev/null && pwd )"

export GOPATH=${SCRIPT_DIR}/../../../../..

cd ${SCRIPT_DIR}

CGO_ENABLED=1 GOOS=linux GOARCH=amd64 go build -v -a -buildmode=plugin  -o service-manager-plugin.so  ./main.go
docker build .. -f ./Dockerfile -t gcr.io/sap-se-gcp-istio-dev/sb-istio-proxy-k8s
docker push gcr.io/sap-se-gcp-istio-dev/sb-istio-proxy-k8s

