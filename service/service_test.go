package service

import (
	"context"
	"errors"
	"strings"
	"testing"

	pb "github.com/buoyantio/bb/gen"
)

func TestRequestHandler(t *testing.T) {
	t.Run("delegates to underlying strategy", func(t *testing.T) {
		expectedRequest := &pb.TheRequest{RequestUID: "expected req"}
		expectedResponse := &pb.TheResponse{Payload: "expected resp"}
		strategy := &MockStrategy{
			ResponseToReturn: expectedResponse,
		}

		handler := RequestHandler{
			config:   &Config{},
			Strategy: strategy,
		}

		actualResponse, err := handler.Handle(context.TODO(), expectedRequest)
		if err != nil {
			t.Fatalf("Unexpected error: %v", err)
		}

		if actualResponse.Payload != expectedResponse.Payload {
			t.Fatalf("Expected returned response to have payload [%s], but got [%s]", expectedResponse.Payload, actualResponse.Payload)
		}

		if actualResponse.RequestUID != expectedRequest.RequestUID {
			t.Fatalf("Expected response to have UID [%s], but got [%s]", expectedRequest.RequestUID, actualResponse.RequestUID)
		}
	})

	t.Run("returns error when underlying strategy has error", func(t *testing.T) {
		expectedRequest := &pb.TheRequest{RequestUID: "expected req"}
		expectedError := errors.New("expected")
		strategy := &MockStrategy{
			ErrorToReturn: expectedError,
		}

		handler := RequestHandler{
			config:   &Config{},
			Strategy: strategy,
		}

		_, err := handler.Handle(context.TODO(), expectedRequest)
		if err != expectedError {
			t.Fatalf("Expected returned error to be [%v], but got [%v]", expectedError, err)
		}
	})

	t.Run("will fail requests as per failure percentage", func(t *testing.T) {
		expectedRequest := &pb.TheRequest{RequestUID: "expected req"}
		expectedResponse := &pb.TheResponse{Payload: "expected resp"}
		strategy := &MockStrategy{
			ResponseToReturn: expectedResponse,
		}

		neverFailHandler := RequestHandler{
			config: &Config{
				PercentageFailedRequests: 0,
			},
			Strategy: strategy,
		}

		alwaysFailHandler := RequestHandler{
			config: &Config{
				PercentageFailedRequests: 100,
			},
			Strategy: strategy,
		}

		sometimesFailHandler := RequestHandler{
			config: &Config{
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

	t.Run("will exit after a specified number of requests", func(t *testing.T) {
		expectedRequest := &pb.TheRequest{RequestUID: "expected req"}
		expectedResponse := &pb.TheResponse{Payload: "expected resp"}
		strategy := &MockStrategy{
			ResponseToReturn: expectedResponse,
		}

		terminateLimit := 2
		terminateHandler := NewRequestHandler(&Config{TerminateAfter: terminateLimit})
		terminateHandler.Strategy = strategy

		for i := 0; i < terminateLimit; i++ {
			_, err := terminateHandler.Handle(context.TODO(), expectedRequest)
			if err != nil {
				t.Fatalf("Unexpected error: %v", err)
			}

			if i+1 < terminateLimit {
				select {
				case <-terminateHandler.Stopping():
					t.Fatalf("RequestHandler terminated when it should not have: %d != %d", i+1, terminateLimit)
				default:
				}
			} else {
				// this will timeout the test if it fails
				<-terminateHandler.Stopping()
			}
		}
	})
}

func TestFireAndForgetClient(t *testing.T) {
	t.Run("calls underlying client and returns stub response", func(t *testing.T) {
		barrier := make(chan bool)
		expectedResponseToClose := errors.New("error to return on close")
		expectedResponseToGetID := "some id goes here"

		underlyingClient := &MockClient{
			IDToReturn:       expectedResponseToGetID,
			ErrorToReturn:    expectedResponseToClose,
			ResponseToReturn: &pb.TheResponse{Payload: "this will be ignored anyway"},
			RequestInterceptor: func(req *pb.TheRequest) {
				barrier <- true
			},
		}

		fnfClient := fireAndForgetClient{
			underlyingClient: underlyingClient,
		}

		actualResponseToGetID := fnfClient.GetID()
		if actualResponseToGetID != expectedResponseToGetID {
			t.Fatalf("Expected call to getID() to be delegated and return [%s], but got [%s]", expectedResponseToGetID, actualResponseToGetID)
		}

		actualResponseToClose := fnfClient.Close()
		if actualResponseToClose != expectedResponseToClose {
			t.Fatalf("Expecting call to Close() to return [%v], but got [%v]", expectedResponseToClose, actualResponseToClose)
		}

		request := &pb.TheRequest{
			RequestUID: "some request UID goes here",
		}

		response, err := fnfClient.Send(request)
		if err != nil {
			t.Fatalf("Unexpected error: %v", expectedResponseToClose)
		}

		if !strings.Contains(response.Payload, "fire-and-forget") && !strings.Contains(response.Payload, request.RequestUID) {
			t.Fatalf("Expected response's payload to contain the fire-and-forget stub message, got [%v]", response)
		}

		<-barrier
		actualRequestReceived := underlyingClient.RequestReceived
		if actualRequestReceived != request {
			t.Fatalf("Expected fire and forget client to eventually delegate request to client, but got [%v]", actualRequestReceived)
		}
	})

	t.Run("ignores errors from calling underlying client and returns stub response", func(t *testing.T) {

		underlyingClient := &MockClient{
			ErrorToReturn: errors.New("this error will be ignored"),
		}

		fnfClient := fireAndForgetClient{
			underlyingClient: underlyingClient,
		}

		request := &pb.TheRequest{
			RequestUID: "some request UID goes here",
		}

		response, err := fnfClient.Send(request)
		if err != nil {
			t.Fatalf("Unexpected error: %v", errors.New("error to return on close"))
		}

		if !strings.Contains(response.Payload, "fire-and-forget") && !strings.Contains(response.Payload, request.RequestUID) {
			t.Fatalf("Expected response's payload to contain the fire-and-forget stub message, got [%v]", response)
		}

	})
}

func TestService(t *testing.T) {
	t.Run("it closes all underlying clients", func(t *testing.T) {
		client1 := &MockClient{}
		client2 := &MockClient{}
		expectedClients := []Client{client1, client2}

		svc := Service{
			Clients: expectedClients,
		}

		err := svc.Close()
		if err != nil {
			t.Fatalf("Unexpected error: %v", err)
		}

		if !client1.CloseWasCalled || !client2.CloseWasCalled {
			t.Fatalf("expecting close to be called for both [%v] and [%v]", client1, client2)
		}
	})

	t.Run("close returns error if any client returns error when closing", func(t *testing.T) {
		someError := errors.New("expected")
		client1 := &MockClient{
			ErrorToReturn: someError,
		}
		client2 := &MockClient{}
		expectedClients := []Client{client1, client2}

		svc := Service{
			Clients: expectedClients,
		}

		err := svc.Close()
		if err == nil {
			t.Fatalf("Expecting error, got nothing")
		}

		if !client1.CloseWasCalled || !client2.CloseWasCalled {
			t.Fatalf("expecting close to be called for both [%v] and [%v]", client1, client2)
		}
	})
}
