package strategies

import (
	"context"
	"fmt"
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
	// HttpEgressStrategyName is the user-friendly name of this strategy
	HttpEgressStrategyName = "http-egress"

	// HttpEgressUrlToInvokeArgName is the parameter used to supply the URL to fetch from
	HttpEgressUrlToInvokeArgName = "url"

	// HttpEgressHttpMethodToUseArgName is the parameter used to supply the HTTP 1.1 method used when fetching the URL
	HttpEgressHttpMethodToUseArgName = "method"

	// HttpEgressHttpTimeoutArgName is the timeout used to configure the HTTP client when fetching the URL
	HttpEgressHttpTimeoutArgName = "http-client-timeout"
)

var validHttpMethods = map[string]bool{"GET": true, "POST": true, "PUT": true, "DELETE": true, "PATCH": true}

// HttpEgressStrategy a strategy that makes a HTTP 1.1 call to a pre-configured URL
type HttpEgressStrategy struct {
	httpClientToUse *http.Client
	urlToInvoke     string
	methodToUse     string
}

// Do executes the request
func (s *HttpEgressStrategy) Do(_ context.Context, req *pb.TheRequest) (*pb.TheResponse, error) {

	httpRequest, err := http.NewRequest(s.methodToUse, s.urlToInvoke, strings.NewReader(req.RequestUid))
	if err != nil {
		return nil, err
	}

	log.Infof("Making [%s] request to [%s] for requestUid [%s]", s.methodToUse, s.urlToInvoke, req.GetRequestUid())
	httpResp, err := s.httpClientToUse.Do(httpRequest)
	if err != nil {
		return nil, err
	}

	log.Infof("Response from [%s] for requestUid [%s] was: %+v", s.urlToInvoke, req.GetRequestUid(), httpResp)
	statusCode := httpResp.StatusCode
	if statusCode < 200 || statusCode > 299 {
		return nil, fmt.Errorf("unexpected status returned by [%s]for requestUid [%s]: %d", s.urlToInvoke, req.GetRequestUid(), statusCode)
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

// NewHttpEgress creates a new HttpEgressStrategy
func NewHttpEgress(config *service.Config, servers []service.Server, clients []service.Client) (service.Strategy, error) {
	if len(clients) != 0 || len(servers) == 0 {
		return nil, fmt.Errorf("strategy [%s] requires at least one server port and exactly zero downstream services, but was configured as: %+v", HttpEgressStrategyName, config)
	}

	urlToInvoke := config.ExtraArguments[HttpEgressUrlToInvokeArgName]
	if urlToInvoke == "" {
		return nil, fmt.Errorf("URL to invoke is nil")
	}

	isHttp, err := regexp.MatchString("https?://", urlToInvoke)
	if err != nil {
		return nil, fmt.Errorf("error while validating URL [%s]: %v", urlToInvoke, err)
	}
	if !isHttp {
		return nil, fmt.Errorf("url must be HTTP or HTTPS, was [%s]", urlToInvoke)
	}

	_, err = url.Parse(urlToInvoke)
	if err != nil {
		return nil, fmt.Errorf("error while parsing URL [%s]: %v", urlToInvoke, err)
	}

	httpMethodToUse := config.ExtraArguments[HttpEgressHttpMethodToUseArgName]
	if !validHttpMethods[httpMethodToUse] {
		return nil, fmt.Errorf("HTTP method [%s] isn't supported [%v]", httpMethodToUse, validHttpMethods)
	}

	timeout, err := time.ParseDuration(config.ExtraArguments[HttpEgressHttpTimeoutArgName])
	if err != nil {
		return nil, fmt.Errorf("error while parsing timeout [%s]: %v", config.ExtraArguments[HttpEgressHttpTimeoutArgName], err)
	}

	httpClient := &http.Client{
		Timeout: timeout,
	}
	log.Infof("HTTP client being used is: %+v", httpClient)

	return &HttpEgressStrategy{
		urlToInvoke:     urlToInvoke,
		methodToUse:     httpMethodToUse,
		httpClientToUse: httpClient,
	}, nil
}
