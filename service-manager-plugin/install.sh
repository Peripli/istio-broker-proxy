#!/bin/bash

set -euox pipefail

SCRIPT_DIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" >/dev/null && pwd )"

if [ -z $HUB ] || [ -z $TAG ]; then
  echo "error: Environment variables HUB and TAG must be defined."
  exit 1
fi

export GOPATH=${SCRIPT_DIR}/../../../../..

cd ${SCRIPT_DIR}
cd $GOPATH/src/github.com/peripli/service-broker-proxy-k8s

helm del --purge service-broker-proxy || true
helm install \
    --name service-broker-proxy \
    --namespace service-broker-proxy \
    --set config.sm.url=https://service-manager-nocis.cfapps.dev01.aws.istio.sapcloud.io \
    --set sm.user=o1msSsCXOYA5WQ9MjGd+oYxD03CqhZrrsUQMt0IfTzI= \
    --set sm.password=wcRV6lrPNyb/apAXsSk2i1kxZ0fMfteMJ6GoI8VheT4= \
    --set image.repository=$HUB/sb-istio-proxy-k8s \
    --set image.tag=$TAG \
    charts/service-broker-proxy-k8s
