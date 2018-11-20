#!/bin/bash

set -euox pipefail

SCRIPT_DIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" >/dev/null && pwd )"
GOPATH=${SCRIPT_DIR}/../../../../..

cd ${SCRIPT_DIR}
cd $GOPATH/src/github.com/peripli/service-broker-proxy-k8s

helm del --purge service-broker-proxy || true
helm install \
    --name service-broker-proxy \
    --namespace service-broker-proxy \
    --set config.sm.url=https://service-manager-nocis.cfapps.dev01.aws.istio.sapcloud.io \
    --set sm.user=$SM_USER \
    --set sm.password=$SM_PASSWORD \
    --set image.repository=$HUB/sb-istio-proxy-k8s \
    --set image.tag=$TAG \
    charts/service-broker-proxy-k8s
