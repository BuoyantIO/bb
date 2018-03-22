#!/bin/bash

set -eu

bindir="$( cd "$( dirname "${BASH_SOURCE[0]}" )" && pwd )"
rootdir="$( cd "$bindir/.." && pwd )"
cd "$rootdir"

echo "====================== running go vet  ======================"
go vet -v -tests -all ./...

echo ""
echo "====================== running go test ======================"
go test -race -v ./...
