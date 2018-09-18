#!/usr/bin/env bash

set -euo pipefail

godoc2md github.com/andy2046/pool \
    > $GOPATH/src/github.com/andy2046/pool/docs.md
