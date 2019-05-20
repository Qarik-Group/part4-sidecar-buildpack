#!/usr/bin/env bash
set -exuo pipefail

cd "$( dirname "${BASH_SOURCE[0]}" )/.."
source .envrc

GOOS=linux go build -ldflags="-s -w" -o bin/supply sample3-sidecar/supply/cli
GOOS=linux go build -ldflags="-s -w" -o bin/finalize sample3-sidecar/finalize/cli
