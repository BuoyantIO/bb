package strategies

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"testing"

	pb "github.com/buoyantio/bb/gen"
	"github.com/buoyantio/bb/service"
)

func TestHttpEgressStrategy(t *testing.T) {
	t.Run("Rejects malformed or incomplete URLs", func(t *testing.T) {
		malformedURLs := []string{"ftp://httpbin.org", "httpbin.org", "httpbin.org:80", "", "123"}

		for _, urlToInvoke := range malformedURLs {
			httpConfig := &service.Config{
				ExtraArguments: map[string]string{
					HTTPEgressURLToInvokeArgName: urlToInvoke,
				},
			}
			_, err := NewHTTPEgress(httpConfig, []service.Server{service.MockServer{}}, []service.Client{})
			if err == nil {
				t.Fatalf("Expecting error, got nothing when configuring url [%s]", urlToInvoke)
			}

		}
	})

	t.Run("Calls external service using both HTTP and HTTPS", func(t *testing.T) {
		protocols := []string{"http", "https"}

		for _, protocolToTest := range protocols {
			expectedHost := "httpbin.org"
			urlToInvoke := fmt.Sprintf("%s://%s/anything", protocolToTest, expectedHost)
			httpConfig := &service.Config{
				ExtraArguments: map[string]string{
					HTTPEgressHTTPMethodToUseArgName: "GET",
					HTTPEgressURLToInvokeArgName:     urlToInvoke,
					HTTPEgressHTTPTimeoutArgName:     "10s",
				},
			}
			egress, err := NewHTTPEgress(httpConfig, []service.Server{service.MockServer{}}, []service.Client{})
			if err != nil {
				t.Fatalf("Unexpected error: %v", err)
			}

			request := &pb.TheRequest{}
			response, err := egress.Do(context.Background(), request)
			if err != nil {
				t.Fatalf("Unexpected error: %v", err)
			}

			var jsonPayload map[string]interface{}
			json.Unmarshal([]byte(response.Payload), &jsonPayload)

			// hard-code HTTPS due to https://github.com/postmanlabs/httpbin/issues/536
			expectedURL := fmt.Sprintf("https://%s/anything", expectedHost)
			actualURL := jsonPayload["url"]
			if actualURL != expectedURL {
				t.Fatalf("Expected HTTP URL to be [%s], but got [%s]", expectedURL, actualURL)
			}

			expectedMethod := "GET"
			actualMethod := jsonPayload["method"]
			if actualMethod != expectedMethod {
				t.Fatalf("Expected HTTP method to be [%s], but got [%s]", expectedMethod, actualMethod)
			}

			headers := jsonPayload["headers"].(map[string]interface{})

			actualHostHeader := headers["Host"]
			if actualHostHeader != expectedHost {
				t.Fatalf("Expected Host header to be [%s], but got [%s]", expectedHost, actualHostHeader)
			}
		}
	})

	t.Run("Can call external service using any HTTP method", func(t *testing.T) {
		methods := []string{"GET", "POST", "PATCH", "PUT", "DELETE"}

		for _, methodToTest := range methods {
			expectedHost := "httpbin.org"
			urlToInvoke := fmt.Sprintf("https://httpbin.org/%s", strings.ToLower(methodToTest))

			httpConfig := &service.Config{
				ExtraArguments: map[string]string{
					HTTPEgressURLToInvokeArgName:     urlToInvoke,
					HTTPEgressHTTPMethodToUseArgName: methodToTest,
					HTTPEgressHTTPTimeoutArgName:     "10s",
				},
			}
			egress, err := NewHTTPEgress(httpConfig, []service.Server{service.MockServer{}}, []service.Client{})
			if err != nil {
				t.Fatalf("Unexpected error: %v", err)
			}

			request := &pb.TheRequest{}
			response, err := egress.Do(context.Background(), request)
			if err != nil {
				t.Fatalf("Unexpected error: %v", err)
			}

			var jsonPayload map[string]interface{}
			json.Unmarshal([]byte(response.Payload), &jsonPayload)

			expectedURL := urlToInvoke
			actualURL := jsonPayload["url"]
			if actualURL != expectedURL {
				t.Fatalf("Expected HTTP method to be [%s], but got [%s]", expectedURL, actualURL)
			}

			headers := jsonPayload["headers"].(map[string]interface{})

			actualHostHeader := headers["Host"]
			if actualHostHeader != expectedHost {
				t.Fatalf("Expected Host header to be [%s], but got [%s]", expectedHost, actualHostHeader)
			}
		}
	})

	t.Run("Returns error when response is anything but status 2xx or 3xx", func(t *testing.T) {
		unpexpectedStatus := []int{100, 101, 400, 401, 403, 404, 405, 406, 407, 408, 409, 410, 411, 412, 413, 414,
			415, 416, 417, 418, 426, 428, 429, 431, 451, 500, 501, 502, 503, 504, 505, 511,
		}

		for _, statusToReturn := range unpexpectedStatus {
			urlToInvoke := fmt.Sprintf("https://httpbin.org/status/%d", statusToReturn)
			httpConfig := &service.Config{
				ExtraArguments: map[string]string{
					HTTPEgressHTTPMethodToUseArgName: "GET",
					HTTPEgressURLToInvokeArgName:     urlToInvoke,
					HTTPEgressHTTPTimeoutArgName:     "10s",
				},
			}
			egress, err := NewHTTPEgress(httpConfig, []service.Server{service.MockServer{}}, []service.Client{})
			if err != nil {
				t.Fatalf("Unexpected error: %v", err)
			}

			request := &pb.TheRequest{}
			response, err := egress.Do(context.Background(), request)
			if err == nil {
				t.Fatalf("Expecting error, got nothing but respo0nse: %v", response)
			}

		}
	})
}
