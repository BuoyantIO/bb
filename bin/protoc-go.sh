#!/bin/sh

set -eu

generated_src_dir=building_blocks/gen

go install ./vendor/github.com/golang/protobuf/protoc-gen-go

rm -rf $generated_src_dir
mkdir $generated_src_dir
bin/protoc -I building_blocks --go_out=plugins=grpc:$generated_src_dir building_blocks/api.proto

