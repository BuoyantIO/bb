#!/usr/bin/env bash

set -e
target_dir=target

rm -rf ${target_dir}
mkdir ${target_dir}
GOOS=linux go build -o target/bb ./building_blocks
docker build  -f building_blocks/Dockerfile . -t buoyantio/bb:v1