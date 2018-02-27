package strategies

import (
	"context"
	"encoding/json"
	"fmt"
	"testing"

	pb "github.com/buoyantio/conduit-test/building_blocks/gen"
	"github.com/buoyantio/conduit-test/building_blocks/service"
)

func TestHttpEgressStrategy(t *testing.T) {

	t.Run("Calls external service using both HTTP and HTTPS", func(t *testing.T) {
		protocols := []string{"http", "https"}

		for _, protocolToTest := range protocols {
			expectedHost := "httpbin.org"
			urlToInvoke := fmt.Sprintf("%s://%s/anything", protocolToTest, expectedHost)
			httpConfig := &service.Config{
				ExtraArguments: map[string]string{
					HttpEgressUrlToInvokeArgName: urlToInvoke,
				},
			}
			egress, err := NewHttpEgress(httpConfig, []service.Server{service.MockServer{}}, []service.Client{})
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

			expectedUrl := urlToInvoke
			actualUrl := jsonPayload["url"]
			if actualUrl != expectedUrl {
				t.Fatalf("Expected HTTP method to be [%s], but got [%s]", expectedUrl, actualUrl)
			}

			expectedMethod := "GET"
			actualMethod := jsonPayload["method"]
			if actualMethod != expectedMethod {
				t.Fatalf("Expected HTTP method to be [%s], but got [%s]", expectedMethod, actualMethod)
			}

			headers := jsonPayload["headers"].(map[string]interface{})

			actualHostHeader := headers["Host"]
			if actualHostHeader != expectedHost {
				t.Fatalf("Expected Hist header to be [%s], but got [%s]", expectedMethod, actualHostHeader)
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
					HttpEgressUrlToInvokeArgName: urlToInvoke,
				},
			}
			egress, err := NewHttpEgress(httpConfig, []service.Server{service.MockServer{}}, []service.Client{})
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
