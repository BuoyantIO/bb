package strategies

import (
	"context"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"net/url"
	"regexp"
	"strings"
	"time"

	pb "github.com/buoyantio/bb/gen"
	"github.com/buoyantio/bb/service"
	log "github.com/sirupsen/logrus"
)

const (
	// HTTPEgressStrategyName is the user-friendly name of this strategy
	HTTPEgressStrategyName = "http-egress"

	// HTTPEgressURLToInvokeArgName is the parameter used to supply the URL to fetch from
	HTTPEgressURLToInvokeArgName = "url"

	// HTTPEgressHTTPMethodToUseArgName is the parameter used to supply the HTTP 1.1 method used when fetching the URL
	HTTPEgressHTTPMethodToUseArgName = "method"

	// HTTPEgressHTTPTimeoutArgName is the timeout used to configure the HTTP client when fetching the URL
	HTTPEgressHTTPTimeoutArgName = "http-client-timeout"
)

var validHTTPMethods = map[string]bool{"GET": true, "POST": true, "PUT": true, "DELETE": true, "PATCH": true}

// HTTPEgressStrategy a strategy that makes a HTTP 1.1 call to a pre-configured URL
type HTTPEgressStrategy struct {
	httpClientToUse *http.Client
	urlToInvoke     string
	methodToUse     string
}

// Do executes the request
func (s *HTTPEgressStrategy) Do(_ context.Context, req *pb.TheRequest) (*pb.TheResponse, error) {
	var body io.Reader
	if s.methodToUse == http.MethodPost || s.methodToUse == http.MethodPut || s.methodToUse == http.MethodPatch {
		// only POST, PUT and PATCH methods can have a body
		body = strings.NewReader(req.RequestUID)
	}

	httpRequest, err := http.NewRequest(s.methodToUse, s.urlToInvoke, body)
	if err != nil {
		return nil, err
	}

	log.Infof("Making [%s] request to [%s] for requestUID [%s]", s.methodToUse, s.urlToInvoke, req.GetRequestUID())
	httpResp, err := s.httpClientToUse.Do(httpRequest)
	if err != nil {
		return nil, err
	}

	log.Infof("Response from [%s] for requestUID [%s] was: %+v", s.urlToInvoke, req.GetRequestUID(), httpResp)
	statusCode := httpResp.StatusCode
	if statusCode < 200 || statusCode > 299 {
		return nil, fmt.Errorf("unexpected status returned by [%s]for requestUID [%s]: %d", s.urlToInvoke, req.GetRequestUID(), statusCode)
	}

	bytes, err := ioutil.ReadAll(httpResp.Body)
	if err != nil {
		return nil, err
	}

	resp := &pb.TheResponse{
		Payload: string(bytes),
	}
	return resp, err
}

// NewHTTPEgress creates a new HTTPEgressStrategy
func NewHTTPEgress(config *service.Config, servers []service.Server, clients []service.Client) (service.Strategy, error) {
	if len(clients) != 0 || len(servers) == 0 {
		return nil, fmt.Errorf("strategy [%s] requires at least one server port and exactly zero downstream services, but was configured as: %+v", HTTPEgressStrategyName, config)
	}

	urlToInvoke := config.ExtraArguments[HTTPEgressURLToInvokeArgName]
	if urlToInvoke == "" {
		return nil, fmt.Errorf("URL to invoke is nil")
	}

	isHTTP, err := regexp.MatchString("https?://", urlToInvoke)
	if err != nil {
		return nil, fmt.Errorf("error while validating URL [%s]: %v", urlToInvoke, err)
	}
	if !isHTTP {
		return nil, fmt.Errorf("url must be HTTP or HTTPS, was [%s]", urlToInvoke)
	}

	_, err = url.Parse(urlToInvoke)
	if err != nil {
		return nil, fmt.Errorf("error while parsing URL [%s]: %v", urlToInvoke, err)
	}

	httpMethodToUse := config.ExtraArguments[HTTPEgressHTTPMethodToUseArgName]
	if !validHTTPMethods[httpMethodToUse] {
		return nil, fmt.Errorf("HTTP method [%s] isn't supported [%v]", httpMethodToUse, validHTTPMethods)
	}

	timeout, err := time.ParseDuration(config.ExtraArguments[HTTPEgressHTTPTimeoutArgName])
	if err != nil {
		return nil, fmt.Errorf("error while parsing timeout [%s]: %v", config.ExtraArguments[HTTPEgressHTTPTimeoutArgName], err)
	}

	httpClient := &http.Client{
		Timeout: timeout,
	}
	log.Infof("HTTP client being used is: %+v", httpClient)

	return &HTTPEgressStrategy{
		urlToInvoke:     urlToInvoke,
		methodToUse:     httpMethodToUse,
		httpClientToUse: httpClient,
	}, nil
}
