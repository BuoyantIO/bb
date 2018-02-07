package service

import (
	"context"

	pb "github.com/buoyantio/conduit-test/building_blocks/gen"
)

type MockClient struct {
	IdToReturn       string
	ResponseToReturn *pb.TheResponse
	ErrorToReturn    error
	RequestReceived  *pb.TheRequest
}

func (m *MockClient) Close() error { return m.ErrorToReturn }

func (m *MockClient) GetId() string { return m.IdToReturn }

func (m *MockClient) Send(req *pb.TheRequest) (*pb.TheResponse, error) {
	m.RequestReceived = req
	return m.ResponseToReturn, m.ErrorToReturn
}

type MockServer struct {
	IdToReturn string
}

func (m MockServer) GetId() string { return m.IdToReturn }

type MockStrategy struct {
	ContextReceived  context.Context
	RequestReceived  *pb.TheRequest
	ResponseToReturn *pb.TheResponse
	ErrorToReturn    error
}

func (m *MockStrategy) Do(ctx context.Context, req *pb.TheRequest) (*pb.TheResponse, error) {
	m.ContextReceived = ctx
	m.RequestReceived = req

	return m.ResponseToReturn, m.ErrorToReturn
}
