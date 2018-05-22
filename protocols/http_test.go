package protocols

import (
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	pb "github.com/buoyantio/bb/gen"
	"github.com/buoyantio/bb/service"
	"github.com/gogo/protobuf/jsonpb"
)

func TestTheHTTPServer(t *testing.T) {
	t.Run("adds request UID when request doesnt have one", func(t *testing.T) {
		expectedProtoResponse := &pb.TheResponse{}

		strategy := &stubStrategy{
			theResponseToReturn: expectedProtoResponse,
		}

		requestHandler := service.NewRequestHandler(&service.Config{})
		requestHandler.Strategy = strategy
		handler := newHTTPHandler(requestHandler)
		theServer := httptest.NewServer(handler)
		defer theServer.Close()

		resp, err := http.Post(theServer.URL, "application/json", strings.NewReader(""))
		if err != nil {
			t.Fatalf("Unexpected error: %v", err)
		}

		expectedHTTPStatus := http.StatusOK
		if resp.StatusCode != expectedHTTPStatus {
			t.Fatalf("Expecting response to have status [%d] but was: %v", expectedHTTPStatus, resp)
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

		if actualProtoResponse.RequestUID == "" {
			t.Fatalf("Expected HTTP response to contain a new request UID assigned to protobuf, but was nil")
		}
	})

	t.Run("serializes the response returned by the strategy", func(t *testing.T) {
		expectedProtoResponse := &pb.TheResponse{
			Payload: "something",
		}

		expectedProtoRequest := &pb.TheRequest{
			RequestUID: "123",
		}

		strategy := &stubStrategy{
			theResponseToReturn: expectedProtoResponse,
		}

		requestHandler := service.NewRequestHandler(&service.Config{})
		requestHandler.Strategy = strategy
		handler := newHTTPHandler(requestHandler)
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
		if expectedProtoRequest.RequestUID != actualProtoRequest.RequestUID {
			t.Fatalf("Expected HTTP request to contain protobuf [%v] but it was [%v]", expectedProtoRequest, actualProtoRequest)
		}

		expectedHTTPStatus := http.StatusOK
		if resp.StatusCode != expectedHTTPStatus {
			t.Fatalf("Expecting response to have status [%d] but was: %v", expectedHTTPStatus, resp)
		}

		defer resp.Body.Close()
		bytesResp, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			t.Fatalf("Unexpected error: %v", err)
		}

		var actualProtoResponse pb.TheResponse
		jsonpb.UnmarshalString(string(bytesResp), &actualProtoResponse)

		if expectedProtoResponse.Payload != actualProtoResponse.Payload {
			t.Fatalf("Expected HTTP response to contain protobuf [%v] but it was [%v]", expectedProtoResponse, actualProtoResponse)
		}
	})

	t.Run("returns a 500 if payload is not the expected protobuf as json", func(t *testing.T) {
		strategy := &stubStrategy{}

		requestHandler := service.NewRequestHandler(&service.Config{})
		requestHandler.Strategy = strategy
		handler := newHTTPHandler(requestHandler)
		theServer := httptest.NewServer(handler)
		defer theServer.Close()

		req := "this error was injected by [terminus-grpc:-1-h1:9090]"

		resp, err := http.Post(theServer.URL, "application/json", strings.NewReader(req))
		if err != nil {
			t.Fatalf("Unexpected error: %v", err)
		}

		if strategy.theRequestReceived != nil {
			t.Fatalf("Expected HTTP server not to delegate error request to strategy, but got [%v]", strategy.theRequestReceived)
		}

		expectedHTTPStatus := http.StatusInternalServerError
		if resp.StatusCode != expectedHTTPStatus {
			t.Fatalf("Expecting response to have status [%d] but was: %v", expectedHTTPStatus, resp)
		}
	})

	t.Run("returns a 500 if strategy returned error", func(t *testing.T) {
		expectedError := errors.New("expected")

		strategy := &stubStrategy{
			theErrorToReturn: expectedError,
		}

		expectedProtoRequest := &pb.TheRequest{
			RequestUID: "123",
		}

		req, err := marshaller.MarshalToString(expectedProtoRequest)
		if err != nil {
			t.Fatalf("Unexpected error: %v", err)
		}

		requestHandler := service.NewRequestHandler(&service.Config{})
		requestHandler.Strategy = strategy
		handler := newHTTPHandler(requestHandler)
		theServer := httptest.NewServer(handler)
		defer theServer.Close()

		resp, err := http.Post(theServer.URL, "application/json", strings.NewReader(req))
		if err != nil {
			t.Fatalf("Unexpected error: %v", err)
		}

		actualProtoRequest := strategy.theRequestReceived
		if expectedProtoRequest.RequestUID != actualProtoRequest.RequestUID {
			t.Fatalf("Expected HTTP request to contain protobuf [%v] but it was [%v]", expectedProtoRequest, actualProtoRequest)
		}

		expectedHTTPStatus := http.StatusInternalServerError
		if resp.StatusCode != expectedHTTPStatus {
			t.Fatalf("Expecting response to have status [%d] but was: %v", expectedHTTPStatus, resp)
		}

		defer resp.Body.Close()
		bytesResp, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			t.Fatalf("Unexpected error: %v", err)
		}

		stringResp := string(bytesResp)

		expectedInBody := expectedError.Error()
		if !strings.Contains(stringResp, expectedInBody) {
			t.Fatalf("Expecting response body to contain the error message [%s], but got [%s]", expectedInBody, stringResp)
		}
	})
}

func TestHTTPClient(t *testing.T) {
	t.Run("returns expected response when everything went well", func(t *testing.T) {
		expectedProtoResponse := &pb.TheResponse{
			Payload: "something",
		}

		expectedProtoRequest := &pb.TheRequest{
			RequestUID: "123",
		}

		strategy := &stubStrategy{
			theResponseToReturn: expectedProtoResponse,
		}

		requestHandler := service.NewRequestHandler(&service.Config{})
		requestHandler.Strategy = strategy
		handler := newHTTPHandler(requestHandler)
		theServer := httptest.NewServer(handler)
		defer theServer.Close()

		client := httpClient{
			id:                        t.Name(),
			serverURL:                 theServer.URL,
			clientForDownsteamServers: http.DefaultClient,
		}

		actualProtoResponse, err := client.Send(expectedProtoRequest)
		if err != nil {
			t.Fatalf("Unexpected error: %v", err)
		}

		actualProtoRequest := strategy.theRequestReceived
		if expectedProtoRequest.RequestUID != actualProtoRequest.RequestUID {
			t.Fatalf("Expected HTTP request to contain protobuf [%v] but it was [%v]", expectedProtoRequest, actualProtoRequest)
		}

		if expectedProtoResponse.Payload != actualProtoResponse.Payload {
			t.Fatalf("Expected HTTP response to contain protobuf [%v] but it was [%v]", expectedProtoResponse, actualProtoResponse)
		}
	})

	t.Run("returns error when server returned any 5xx error", func(t *testing.T) {
		expectedProtoRequest := &pb.TheRequest{
			RequestUID: "123",
		}

		strategy := &stubStrategy{
			theErrorToReturn: errors.New("this error was injected by [terminus-grpc:-1-h1:9090]"),
		}

		requestHandler := service.NewRequestHandler(&service.Config{})
		requestHandler.Strategy = strategy
		handler := newHTTPHandler(requestHandler)
		theServer := httptest.NewServer(handler)
		defer theServer.Close()

		client := httpClient{
			id:                        t.Name(),
			serverURL:                 theServer.URL,
			clientForDownsteamServers: http.DefaultClient,
		}

		_, err := client.Send(expectedProtoRequest)
		if err == nil {
			t.Fatalf("Expecting error, got nothing")
		}

		actualProtoRequest := strategy.theRequestReceived
		if expectedProtoRequest.RequestUID != actualProtoRequest.RequestUID {
			t.Fatalf("Expected HTTP request to contain protobuf [%v] but it was [%v]", expectedProtoRequest, actualProtoRequest)
		}
	})

	t.Run("returns error when server returned something that isn't the expected protobuf in json", func(t *testing.T) {
		expectedProtoRequest := &pb.TheRequest{
			RequestUID: "123",
		}

		expectedPayload := "this error was injected by [terminus-grpc:-1-h1:9090]"

		theServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			fmt.Fprint(w, expectedPayload)
		}))
		defer theServer.Close()

		client := httpClient{
			id:                        t.Name(),
			serverURL:                 theServer.URL,
			clientForDownsteamServers: http.DefaultClient,
		}

		_, err := client.Send(expectedProtoRequest)
		if err == nil {
			t.Fatalf("Expecting error, got nothing")
		}

		if expectedPayload != err.Error() {
			t.Fatalf("Expecting error text to e [%s], but received [%s]", expectedPayload, err.Error())
		}
	})
}
