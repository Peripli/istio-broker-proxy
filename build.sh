#!/bin/bash

set -euo pipefail

SCRIPT_DIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" >/dev/null && pwd )"

export GOPATH=${SCRIPT_DIR}/../../../..

cd ${SCRIPT_DIR}

go build ./pkg/config/client
./client --help
go build -v
