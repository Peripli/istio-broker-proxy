#!/bin/bash

set -euo pipefail

SCRIPT_DIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" >/dev/null && pwd )"

PGBENCH_OUTPUT=${1:-/tmp/pgbench.out}

export GOPATH=${SCRIPT_DIR}/../../../../..
export CGO_ENABLED=0

cd ${SCRIPT_DIR}

BROKER_VERSION=`git rev-parse HEAD`
go test -v -mod=vendor -timeout 20m -run . --pgbench-output $PGBENCH_OUTPUT --pgbench-time 60 -ldflags="-X router.version=$BROKER_VERSION"
