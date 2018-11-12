#!/bin/bash

set -euox pipefail

SCRIPT_DIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" >/dev/null && pwd )"

export GOPATH=${SCRIPT_DIR}/../../../../..

cd ${SCRIPT_DIR}

helm del --purge service-broker-proxy
cd $GOPATH/src/github.infra.hana.ondemand.com/istio/service-broker-proxy-k8s
helm install charts/service-broker-proxy-k8s \
    --name service-broker-proxy \
    --namespace service-broker-proxy \
    --set config.sm.url=https://service-manager-nocis.cfapps.dev01.aws.istio.sapcloud.io \
    --set sm.user=o1msSsCXOYA5WQ9MjGd+oYxD03CqhZrrsUQMt0IfTzI= \
    --set sm.password=wcRV6lrPNyb/apAXsSk2i1kxZ0fMfteMJ6GoI8VheT4=
    --set image.repository: gcr.io/sap-se-gcp-istio-dev/sb-istio-proxy-k8s
