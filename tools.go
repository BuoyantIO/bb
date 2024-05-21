//go:build tools
// +build tools

package tools

import (
	// for bin/protoc-gen
	_ "google.golang.org/grpc/cmd/protoc-gen-go-grpc"
)
