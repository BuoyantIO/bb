package service

import (
	"context"

	pb "github.com/buoyantio/bb/gen"
)

type MockClient struct {
	IDToReturn         string
	ResponseToReturn   *pb.TheResponse
	ErrorToReturn      error
	RequestReceived    *pb.TheRequest
	RequestInterceptor func(req *pb.TheRequest)
	CloseWasCalled     bool
}

func (m *MockClient) Close() error {
	m.CloseWasCalled = true
	return m.ErrorToReturn
}

func (m *MockClient) GetID() string { return m.IDToReturn }

func (m *MockClient) Send(req *pb.TheRequest) (*pb.TheResponse, error) {
	m.RequestReceived = req
	if m.RequestInterceptor != nil {
		m.RequestInterceptor(req)
	}
	return m.ResponseToReturn, m.ErrorToReturn
}

type MockServer struct {
	IDToReturn string
}

func (m MockServer) GetID() string { return m.IDToReturn }

func (m MockServer) Shutdown() error { return nil }

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
