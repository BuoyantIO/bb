#!/usr/bin/env bash

set -eu

# keep in sync with google.golang.org/protobuf in go.mod
protoc_gen_go_version=v1.33.0
# keep in sync with google.golang.org/grpc/cmd/protoc-gen-go-grpc in go.mod
protoc_gen_go_grpc_version=v1.3.0

# fetch tools and dependencies
go install google.golang.org/protobuf/cmd/protoc-gen-go@$protoc_gen_go_version
go install google.golang.org/grpc/cmd/protoc-gen-go-grpc@$protoc_gen_go_grpc_version

basedir=$(cd "$(dirname "$0")"/..; pwd)
outdir="$basedir"/gen
rm -rf "$outdir"
mkdir "$outdir"

"$basedir"/bin/protoc \
    --proto_path="$basedir" \
    --go_out="$outdir" \
    --go_opt=paths=source_relative \
    --go-grpc_out="$outdir" \
    --go-grpc_opt=paths=source_relative \
    "$basedir"/api.proto
