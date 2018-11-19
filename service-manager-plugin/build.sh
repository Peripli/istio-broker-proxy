#!/bin/bash

set -euo pipefail

SCRIPT_DIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" >/dev/null && pwd )"
cd ${SCRIPT_DIR}

if [ -z $HUB ] || [ -z $TAG ]; then
  echo "error: Environment variables HUB and TAG must be defined."
  exit 1
fi 

docker build .. -f ./Dockerfile -t $HUB/sb-istio-proxy-k8s:$TAG
docker push $HUB/sb-istio-proxy-k8s:$TAG

