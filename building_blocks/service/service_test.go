package service

import (
	"context"
	"errors"
	"testing"

	pb "github.com/buoyantio/conduit-test/building_blocks/gen"
)

type clientMock struct {
	errorToReturn    error
	idToReturn       string
	responseToReturn *pb.TheResponse
	closeWasCalled   bool
}

func (c *clientMock) Close() error {
	c.closeWasCalled = true
	return c.errorToReturn
}

func (c *clientMock) GetId() string {
	return c.idToReturn
}
func (c *clientMock) Send(*pb.TheRequest) (*pb.TheResponse, error) {
	return c.responseToReturn, c.errorToReturn
}

func TestRequestHandler(t *testing.T) {
	t.Run("delegates to underlying strategy", func(t *testing.T) {
		expectedRequest := &pb.TheRequest{RequestUid: "expected req"}
		expectedResponse := &pb.TheResponse{Payload: "expected resp"}
		strategy := &MockStrategy{
			ResponseToReturn: expectedResponse,
		}

		handler := RequestHandler{
			Config:   &Config{},
			Strategy: strategy,
		}

		actualResponse, err := handler.Handle(context.TODO(), expectedRequest)
		if err != nil {
			t.Fatalf("Unexpected error: %v", err)
		}

		if actualResponse.Payload != expectedResponse.Payload {
			t.Fatalf("Expected returned response to have payload [%s], but got [%s]", expectedResponse.Payload, actualResponse.Payload)
		}

		if actualResponse.RequestUid != expectedRequest.RequestUid {
			t.Fatalf("Expected response to have uid [%s], but got [%s]", expectedRequest.RequestUid, actualResponse.RequestUid)
		}
	})

	t.Run("returns error when underlying strategy has error", func(t *testing.T) {
		expectedRequest := &pb.TheRequest{RequestUid: "expected req"}
		expectedError := errors.New("expected")
		strategy := &MockStrategy{
			ErrorToReturn: expectedError,
		}

		handler := RequestHandler{
			Config:   &Config{},
			Strategy: strategy,
		}

		_, err := handler.Handle(context.TODO(), expectedRequest)
		if err != expectedError {
			t.Fatalf("Expected returned error to be [%v], but got [%v]", expectedError, err)
		}
	})

	t.Run("will fail requests as per failure percentage", func(t *testing.T) {
		expectedRequest := &pb.TheRequest{RequestUid: "expected req"}
		expectedResponse := &pb.TheResponse{Payload: "expected resp"}
		strategy := &MockStrategy{
			ResponseToReturn: expectedResponse,
		}

		neverFailHandler := RequestHandler{
			Config: &Config{
				PercentageFailedRequests: 0,
			},
			Strategy: strategy,
		}

		alwaysFailHandler := RequestHandler{
			Config: &Config{
				PercentageFailedRequests: 100,
			},
			Strategy: strategy,
		}

		sometimesFailHandler := RequestHandler{
			Config: &Config{
				PercentageFailedRequests: 70,
			},
			Strategy: strategy,
		}

		var resultsForSometimesError []error
		for i := 0; i < 10000; i++ {
			_, err := alwaysFailHandler.Handle(context.TODO(), expectedRequest)
			if err == nil {
				t.Fatalf("Expecting error, got nothing")
			}

			_, err = neverFailHandler.Handle(context.TODO(), expectedRequest)
			if err != nil {
				t.Fatalf("Unexpected error: %v", err)
			}

			_, err = sometimesFailHandler.Handle(context.TODO(), expectedRequest)
			if err != nil {
				resultsForSometimesError = append(resultsForSometimesError, err)
			}
		}

		if len(resultsForSometimesError) == 0 {
			t.Fatalf("Expected sometimes error to fail at least once, but it didnt fail")
		}
	})
}

func TestServiceClose(t *testing.T) {
	t.Run("it closes all underlying clients", func(t *testing.T) {
		client1 := &clientMock{}
		client2 := &clientMock{}
		expectedClients := []Client{client1, client2}

		svc := Service{
			Clients: expectedClients,
		}

		err := svc.Close()
		if err != nil {
			t.Fatalf("Unexpected error: %v", err)
		}

		if !client1.closeWasCalled || !client2.closeWasCalled {
			t.Fatalf("expecting close to be called for both [%v] and [%v]", client1, client2)
		}
	})

	t.Run("close returns error if any client returns error when closing", func(t *testing.T) {
		someError := errors.New("expected")
		client1 := &clientMock{
			errorToReturn: someError,
		}
		client2 := &clientMock{}
		expectedClients := []Client{client1, client2}

		svc := Service{
			Clients: expectedClients,
		}

		err := svc.Close()
		if err == nil {
			t.Fatalf("Expecting error, got nothing")
		}

		if !client1.closeWasCalled || !client2.closeWasCalled {
			t.Fatalf("expecting close to be called for both [%v] and [%v]", client1, client2)
		}
	})
}
