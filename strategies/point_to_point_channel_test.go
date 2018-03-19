package strategies

import (
	"context"
	"errors"
	"testing"

	pb "github.com/buoyantio/conduit-test/gen"
	"github.com/buoyantio/conduit-test/service"
)

func TestPointToPointChannelStrategy(t *testing.T) {
	t.Run("forwards all requests to strategy and returns its response", func(t *testing.T) {
		expectedResponse := &pb.TheResponse{Payload: "1"}
		mockClient := &service.MockClient{IdToReturn: "1", ResponseToReturn: expectedResponse}

		expectedRequest := &pb.TheRequest{
			RequestUid: "expected-req",
		}

		strategy, err := NewPointToPointChannel(&service.Config{}, []service.Server{service.MockServer{}}, []service.Client{mockClient})
		if err != nil {
			t.Fatalf("Unexpected error: %v", err)
		}

		actualResponse, err := strategy.Do(context.TODO(), expectedRequest)
		if err != nil {
			t.Fatalf("Unexpected error: %v", err)
		}

		actualRequest := mockClient.RequestReceived
		if actualRequest != expectedRequest {
			t.Fatalf("Expected client [%s] to receive request [%v], but got [%v]", mockClient.GetId(), expectedRequest, actualRequest)
		}

		if actualResponse != expectedResponse {
			t.Fatalf("Expected to return response [%v], but got [%v]", expectedRequest, actualRequest)
		}
	})

	t.Run("forwards errors returned by clients", func(t *testing.T) {
		mockClient := &service.MockClient{IdToReturn: "1", ErrorToReturn: errors.New("expected")}

		expectedRequest := &pb.TheRequest{
			RequestUid: "expected-req",
		}

		strategy, err := NewPointToPointChannel(&service.Config{}, []service.Server{service.MockServer{}}, []service.Client{mockClient})
		if err != nil {
			t.Fatalf("Unexpected error: %v", err)
		}

		_, err = strategy.Do(context.TODO(), expectedRequest)
		if err == nil {
			t.Fatalf("Expecting error, got nothing")
		}

		actualRequest := mockClient.RequestReceived
		if actualRequest != expectedRequest {
			t.Fatalf("Expected client [%s] to receive request [%v], but got [%v]", mockClient.GetId(), expectedRequest, actualRequest)
		}
	})
}
