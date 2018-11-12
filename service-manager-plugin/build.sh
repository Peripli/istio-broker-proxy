#!/bin/bash

set -euo pipefail

SCRIPT_DIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" >/dev/null && pwd )"
cd ${SCRIPT_DIR}

docker build .. -f ./Dockerfile -t gcr.io/sap-se-gcp-istio-dev/sb-istio-proxy-k8s
docker push gcr.io/sap-se-gcp-istio-dev/sb-istio-proxy-k8s

