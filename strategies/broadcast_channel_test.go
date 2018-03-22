package strategies

import (
	"context"
	"errors"
	"strings"
	"testing"

	pb "github.com/buoyantio/bb/gen"
	"github.com/buoyantio/bb/service"
)

func TestScatherGatherChannelStrategy(t *testing.T) {
	allServers := []service.Server{service.MockServer{}}

	t.Run("sends message to all clients and aggregates their response into a single message", func(t *testing.T) {
		client1 := &service.MockClient{IDToReturn: "1", ResponseToReturn: &pb.TheResponse{Payload: "1"}}
		client2 := &service.MockClient{IDToReturn: "2", ResponseToReturn: &pb.TheResponse{Payload: "2"}}
		allClients := []service.Client{client1, client2}

		expectedRequest := &pb.TheRequest{
			RequestUID: "expected-req",
		}

		strategy, err := NewBroadcastChannel(&service.Config{}, allServers, allClients)
		if err != nil {
			t.Fatalf("Unexpected error: %v", err)
		}

		response, err := strategy.Do(context.TODO(), expectedRequest)
		if err != nil {
			t.Fatalf("Unexpected error: %v", err)
		}

		for _, client := range allClients {
			mockClient := client.(*service.MockClient)
			actualRequest := mockClient.RequestReceived
			if actualRequest != expectedRequest {
				t.Fatalf("Expected client [%s] top receive request [%v], but got [%v]", client.GetID(), expectedRequest, actualRequest)
			}

			aggregatedResponse := response.Payload
			expectedResponse := mockClient.ResponseToReturn.Payload
			if !strings.Contains(aggregatedResponse, expectedResponse) {
				t.Fatalf("Expected aggregated response to contain response [%s] from client [%s], but got [%s]", expectedResponse, mockClient.GetID(),
					aggregatedResponse)
			}
		}
	})

	t.Run("sends message to all clients and aggregates any errors into a single error", func(t *testing.T) {
		client1 := &service.MockClient{IDToReturn: "1", ResponseToReturn: &pb.TheResponse{Payload: "1"}}
		client2 := &service.MockClient{IDToReturn: "2", ErrorToReturn: errors.New("2")}
		client3 := &service.MockClient{IDToReturn: "3", ResponseToReturn: &pb.TheResponse{Payload: "3"}}
		client4 := &service.MockClient{IDToReturn: "4", ErrorToReturn: errors.New("4")}
		allClients := []service.Client{client1, client2, client3, client4}

		expectedRequest := &pb.TheRequest{
			RequestUID: "expected-req",
		}

		strategy, err := NewBroadcastChannel(&service.Config{}, allServers, allClients)
		if err != nil {
			t.Fatalf("Unexpected error: %v", err)
		}

		response, err := strategy.Do(context.TODO(), expectedRequest)
		if err == nil {
			t.Fatalf("Expecting error, got nothing")
		}
		aggregatedError := err.Error()

		if response != nil {
			t.Fatalf("Expecting response to be nil when an error happens, got: %v", response)
		}

		for _, client := range allClients {
			mockClient := client.(*service.MockClient)
			actualRequest := mockClient.RequestReceived
			if actualRequest != expectedRequest {
				t.Fatalf("Expected client [%s] top receive request [%v], but got [%v]", client.GetID(), expectedRequest, actualRequest)
			}

			if mockClient.ErrorToReturn != nil {
				expectedError := mockClient.ErrorToReturn.Error()
				if !strings.Contains(aggregatedError, expectedError) {
					t.Fatalf("Expected aggregated error to contain [%s] from client [%s], but got [%s]", expectedError, mockClient.GetID(),
						aggregatedError)
				}
			}
		}
	})
}
