package protocols

import (
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	pb "github.com/buoyantio/conduit-test/building_blocks/gen"
	"github.com/buoyantio/conduit-test/building_blocks/service"
	"github.com/gogo/protobuf/jsonpb"
)

func TestTheHttpServer(t *testing.T) {
	t.Run("treats an empty request as the first request in the service call chain", func(t *testing.T) {
		expectedProtoResponse := &pb.TheResponse{}

		strategy := &stubStrategy{
			theResponseToReturn: expectedProtoResponse,
		}

		handler := newHttpHandler(&service.RequestHandler{Strategy: strategy, Config: &service.Config{}})
		theServer := httptest.NewServer(handler)
		defer theServer.Close()

		resp, err := http.Post(theServer.URL, "application/json", strings.NewReader(""))
		if err != nil {
			t.Fatalf("Unexpected error: %v", err)
		}

		expectedHttpStatus := http.StatusOK
		if resp.StatusCode != expectedHttpStatus {
			t.Fatalf("Expecting response to have status [%d] but was: %v", expectedHttpStatus, resp)
		}

		defer resp.Body.Close()
		bytesResp, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			t.Fatalf("Unexpected error: %v", err)
		}

		fmt.Println(string(bytesResp))
		var actualProtoResponse pb.TheResponse
		jsonpb.UnmarshalString(string(bytesResp), &actualProtoResponse)

		if expectedProtoResponse.Payload != actualProtoResponse.Payload {
			t.Fatalf("Expected HTTP response to contain protobuf [%v] but it was [%v]", expectedProtoResponse, actualProtoResponse)
		}
	})

	t.Run("returns whatever the strategy returned", func(t *testing.T) {
		expectedProtoResponse := &pb.TheResponse{
			Payload: "something",
		}

		expectedProtoRequest := &pb.TheRequest{
			RequestUid: "123",
		}

		strategy := &stubStrategy{
			theResponseToReturn: expectedProtoResponse,
		}

		handler := newHttpHandler(&service.RequestHandler{Config: &service.Config{}, Strategy: strategy})
		theServer := httptest.NewServer(handler)
		defer theServer.Close()

		req, err := marshaller.MarshalToString(expectedProtoRequest)
		if err != nil {
			t.Fatalf("Unexpected error: %v", err)
		}

		resp, err := http.Post(theServer.URL, "application/json", strings.NewReader(req))
		if err != nil {
			t.Fatalf("Unexpected error: %v", err)
		}

		actualProtoRequest := strategy.theRequestReceived
		if expectedProtoRequest.RequestUid != actualProtoRequest.RequestUid {
			t.Fatalf("Expected HTTP request to contain protobuf [%v] but it was [%v]", expectedProtoRequest, actualProtoRequest)
		}

		expectedHttpStatus := http.StatusOK
		if resp.StatusCode != expectedHttpStatus {
			t.Fatalf("Expecting response to have status [%d] but was: %v", expectedHttpStatus, resp)
		}

		defer resp.Body.Close()
		bytesResp, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			t.Fatalf("Unexpected error: %v", err)
		}

		fmt.Println(string(bytesResp))
		var actualProtoResponse pb.TheResponse
		jsonpb.UnmarshalString(string(bytesResp), &actualProtoResponse)

		if expectedProtoResponse.Payload != actualProtoResponse.Payload {
			t.Fatalf("Expected HTTP response to contain protobuf [%v] but it was [%v]", expectedProtoResponse, actualProtoResponse)
		}
	})

	t.Run("returns a 500 if payload is not the expected protobuf as json", func(t *testing.T) {
		strategy := &stubStrategy{}

		handler := newHttpHandler(&service.RequestHandler{Config: &service.Config{}, Strategy: strategy})
		theServer := httptest.NewServer(handler)
		defer theServer.Close()

		req := "something that isnt valid"

		resp, err := http.Post(theServer.URL, "application/json", strings.NewReader(req))
		if err != nil {
			t.Fatalf("Unexpected error: %v", err)
		}

		if strategy.theRequestReceived != nil {
			t.Fatalf("Expected HTTP server not to delegate error request to strategy, but got [%v]", strategy.theRequestReceived)
		}

		expectedHttpStatus := http.StatusInternalServerError
		if resp.StatusCode != expectedHttpStatus {
			t.Fatalf("Expecting response to have status [%d] but was: %v", expectedHttpStatus, resp)
		}
	})

	t.Run("returns a 500 if strategy returned error", func(t *testing.T) {
		expectedError := errors.New("expected ")

		strategy := &stubStrategy{
			theErrorToReturn: expectedError,
		}

		expectedProtoRequest := &pb.TheRequest{
			RequestUid: "123",
		}

		req, err := marshaller.MarshalToString(expectedProtoRequest)
		if err != nil {
			t.Fatalf("Unexpected error: %v", err)
		}

		handler := newHttpHandler(&service.RequestHandler{Config: &service.Config{}, Strategy: strategy})
		theServer := httptest.NewServer(handler)
		defer theServer.Close()

		resp, err := http.Post(theServer.URL, "application/json", strings.NewReader(req))
		if err != nil {
			t.Fatalf("Unexpected error: %v", err)
		}

		actualProtoRequest := strategy.theRequestReceived
		if expectedProtoRequest.RequestUid != actualProtoRequest.RequestUid {
			t.Fatalf("Expected HTTP request to contain protobuf [%v] but it was [%v]", expectedProtoRequest, actualProtoRequest)
		}

		expectedHttpStatus := http.StatusInternalServerError
		if resp.StatusCode != expectedHttpStatus {
			t.Fatalf("Expecting response to have status [%d] but was: %v", expectedHttpStatus, resp)
		}
	})
}

func TestHttpClient(t *testing.T) {
	t.Run("returns expected response when everything went well", func(t *testing.T) {
		expectedProtoResponse := &pb.TheResponse{
			Payload: "something",
		}

		expectedProtoRequest := &pb.TheRequest{
			RequestUid: "123",
		}

		strategy := &stubStrategy{
			theResponseToReturn: expectedProtoResponse,
		}

		handler := newHttpHandler(&service.RequestHandler{Config: &service.Config{}, Strategy: strategy})
		theServer := httptest.NewServer(handler)
		defer theServer.Close()

		client := httpClient{
			id:        t.Name(),
			serverUrl: theServer.URL,
		}

		actualProtoResponse, err := client.Send(expectedProtoRequest)
		if err != nil {
			t.Fatalf("Unexpected error: %v", err)
		}

		actualProtoRequest := strategy.theRequestReceived
		if expectedProtoRequest.RequestUid != actualProtoRequest.RequestUid {
			t.Fatalf("Expected HTTP request to contain protobuf [%v] but it was [%v]", expectedProtoRequest, actualProtoRequest)
		}

		if expectedProtoResponse.Payload != actualProtoResponse.Payload {
			t.Fatalf("Expected HTTP response to contain protobuf [%v] but it was [%v]", expectedProtoResponse, actualProtoResponse)
		}
	})

	t.Run("returns error when server returned any 5xx error", func(t *testing.T) {
		expectedProtoRequest := &pb.TheRequest{
			RequestUid: "123",
		}

		strategy := &stubStrategy{
			theErrorToReturn: errors.New("expected"),
		}

		handler := newHttpHandler(&service.RequestHandler{Config: &service.Config{}, Strategy: strategy})
		theServer := httptest.NewServer(handler)
		defer theServer.Close()

		client := httpClient{
			id:        t.Name(),
			serverUrl: theServer.URL,
		}

		_, err := client.Send(expectedProtoRequest)
		if err == nil {
			t.Fatalf("Expecting error, got nothing")
		}

		actualProtoRequest := strategy.theRequestReceived
		if expectedProtoRequest.RequestUid != actualProtoRequest.RequestUid {
			t.Fatalf("Expected HTTP request to contain protobuf [%v] but it was [%v]", expectedProtoRequest, actualProtoRequest)
		}
	})
}
