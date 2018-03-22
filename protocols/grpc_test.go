package protocols

import (
	"context"
	"errors"
	"testing"

	pb "github.com/buoyantio/bb/gen"
	"github.com/buoyantio/bb/service"
)

func TestTheGrpcServer(t *testing.T) {
	t.Run("returns as the response whaterver the strategy returned", func(t *testing.T) {
		expectedProtoResponse := &pb.TheResponse{
			Payload: "something",
		}

		expectedProtoRequest := &pb.TheRequest{
			RequestUID: "123",
		}

		strategy := &stubStrategy{
			theResponseToReturn: expectedProtoResponse,
		}

		grpcServer := theGrpcServer{serviceHandler: &service.RequestHandler{Config: &service.Config{}, Strategy: strategy}}

		actualProtoResponse, err := grpcServer.TheFunction(context.TODO(), expectedProtoRequest)
		if err != nil {
			t.Fatalf("Unexpected error: %v", err)
		}

		actualProtoRequest := strategy.theRequestReceived
		if expectedProtoRequest != actualProtoRequest {
			t.Fatalf("Expected request to be [%v] but it was [%v]", expectedProtoRequest, actualProtoRequest)
		}

		if actualProtoResponse != expectedProtoResponse {
			t.Fatalf("Expected response to be [%v] but it was [%v]", expectedProtoResponse, actualProtoResponse)
		}
	})

	t.Run("returns error if strategy returned error", func(t *testing.T) {
		expectedError := errors.New("expected")

		expectedProtoRequest := &pb.TheRequest{
			RequestUID: "123",
		}

		strategy := &stubStrategy{
			theErrorToReturn: expectedError,
		}

		grpcServer := theGrpcServer{serviceHandler: &service.RequestHandler{Config: &service.Config{}, Strategy: strategy}}

		_, actualError := grpcServer.TheFunction(context.TODO(), expectedProtoRequest)
		if actualError == nil {
			t.Fatalf("Expecting error, got nothing")
		}

		if actualError != expectedError {
			t.Fatalf("Expecting returned error to be [%v] but was [%v]", expectedError, actualError)
		}
	})
}
