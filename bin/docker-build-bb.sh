#!/usr/bin/env bash

set -e
target_dir=target

rm -rf ${target_dir}
mkdir ${target_dir}
GOOS=linux go build -o target/bb .
docker build  -f ./Dockerfile . -t buoyantio/bb:v1
