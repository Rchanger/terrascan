#!/bin/bash

set -o errexit
set -o nounset
set -o pipefail

export GO111MODULE=on
export PATH=$PATH:$(go env GOPATH)/bin
go install honnef.co/go/tools/cmd/staticcheck
staticcheck -f stylish ./...
