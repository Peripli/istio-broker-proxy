#!/bin/bash

set -euo pipefail

SCRIPT_DIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" >/dev/null && pwd )"

export GOPATH=${SCRIPT_DIR}/../../../../..
export CGO_ENABLED=0
export CLEANUP_ORPHANED_OBJECTS=true

cd ${SCRIPT_DIR}

go test -mod=vendor -v -run TestCleanup