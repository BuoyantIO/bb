// Code generated by protoc-gen-go-grpc. DO NOT EDIT.
// versions:
// - protoc-gen-go-grpc v1.3.0
// - protoc             v3.20.1
// source: api.proto

package gen

import (
	context "context"
	grpc "google.golang.org/grpc"
	codes "google.golang.org/grpc/codes"
	status "google.golang.org/grpc/status"
)

// This is a compile-time assertion to ensure that this generated file
// is compatible with the grpc package it is being compiled against.
// Requires gRPC-Go v1.32.0 or later.
const _ = grpc.SupportPackageIsVersion7

const (
	TheService_TheFunction_FullMethodName = "/buoyantio.bb.TheService/theFunction"
)

// TheServiceClient is the client API for TheService service.
//
// For semantics around ctx use and closing/ending streaming RPCs, please refer to https://pkg.go.dev/google.golang.org/grpc/?tab=doc#ClientConn.NewStream.
type TheServiceClient interface {
	TheFunction(ctx context.Context, in *TheRequest, opts ...grpc.CallOption) (*TheResponse, error)
}

type theServiceClient struct {
	cc grpc.ClientConnInterface
}

func NewTheServiceClient(cc grpc.ClientConnInterface) TheServiceClient {
	return &theServiceClient{cc}
}

func (c *theServiceClient) TheFunction(ctx context.Context, in *TheRequest, opts ...grpc.CallOption) (*TheResponse, error) {
	out := new(TheResponse)
	err := c.cc.Invoke(ctx, TheService_TheFunction_FullMethodName, in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

// TheServiceServer is the server API for TheService service.
// All implementations must embed UnimplementedTheServiceServer
// for forward compatibility
type TheServiceServer interface {
	TheFunction(context.Context, *TheRequest) (*TheResponse, error)
	mustEmbedUnimplementedTheServiceServer()
}

// UnimplementedTheServiceServer must be embedded to have forward compatible implementations.
type UnimplementedTheServiceServer struct {
}

func (UnimplementedTheServiceServer) TheFunction(context.Context, *TheRequest) (*TheResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method TheFunction not implemented")
}
func (UnimplementedTheServiceServer) mustEmbedUnimplementedTheServiceServer() {}

// UnsafeTheServiceServer may be embedded to opt out of forward compatibility for this service.
// Use of this interface is not recommended, as added methods to TheServiceServer will
// result in compilation errors.
type UnsafeTheServiceServer interface {
	mustEmbedUnimplementedTheServiceServer()
}

func RegisterTheServiceServer(s grpc.ServiceRegistrar, srv TheServiceServer) {
	s.RegisterService(&TheService_ServiceDesc, srv)
}

func _TheService_TheFunction_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(TheRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(TheServiceServer).TheFunction(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: TheService_TheFunction_FullMethodName,
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(TheServiceServer).TheFunction(ctx, req.(*TheRequest))
	}
	return interceptor(ctx, in, info, handler)
}

// TheService_ServiceDesc is the grpc.ServiceDesc for TheService service.
// It's only intended for direct use with grpc.RegisterService,
// and not to be introspected or modified (even as a copy)
var TheService_ServiceDesc = grpc.ServiceDesc{
	ServiceName: "buoyantio.bb.TheService",
	HandlerType: (*TheServiceServer)(nil),
	Methods: []grpc.MethodDesc{
		{
			MethodName: "theFunction",
			Handler:    _TheService_TheFunction_Handler,
		},
	},
	Streams:  []grpc.StreamDesc{},
	Metadata: "api.proto",
}
