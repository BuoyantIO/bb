package service

import (
	"context"
	"errors"
	"strings"
	"testing"

	pb "github.com/buoyantio/conduit-test/building_blocks/gen"
)

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

func TestFireAndForgetClient(t *testing.T) {
	t.Run("calls underlying client and returns stub response", func(t *testing.T) {
		barrier := make(chan bool)
		expectedResponseToClose := errors.New("error to return on close")
		expectedResponseToGetId := "some id goes here"

		underlyingClient := &MockClient{
			IdToReturn:       expectedResponseToGetId,
			ErrorToReturn:    expectedResponseToClose,
			ResponseToReturn: &pb.TheResponse{Payload: "this will be ignored anyway"},
			RequestInterceptor: func(req *pb.TheRequest) {
				barrier <- true
			},
		}

		fnfClient := fireAndForgetClient{
			underlyingClient: underlyingClient,
		}

		actualResponseToGetId := fnfClient.GetId()
		if actualResponseToGetId != expectedResponseToGetId {
			t.Fatalf("Expected call to getId() to be delegated and return [%s], but got [%s]", expectedResponseToGetId, actualResponseToGetId)
		}

		actualResponseToClose := fnfClient.Close()
		if actualResponseToClose != expectedResponseToClose {
			t.Fatalf("Expecting call to Close() to return [%v], but got [%v]", expectedResponseToClose, actualResponseToClose)
		}

		request := &pb.TheRequest{
			RequestUid: "some request uid goes here",
		}

		response, err := fnfClient.Send(request)
		if err != nil {
			t.Fatalf("Unexpected error: %v", expectedResponseToClose)
		}

		if !strings.Contains(response.Payload, "fire-and-forget") && !strings.Contains(response.Payload, request.RequestUid) {
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
			RequestUid: "some request uid goes here",
		}

		response, err := fnfClient.Send(request)
		if err != nil {
			t.Fatalf("Unexpected error: %v", errors.New("error to return on close"))
		}

		if !strings.Contains(response.Payload, "fire-and-forget") && !strings.Contains(response.Payload, request.RequestUid) {
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
