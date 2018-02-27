package strategies

import (
	"context"
	"fmt"
	"io/ioutil"
	"net/http"

	pb "github.com/buoyantio/conduit-test/building_blocks/gen"
	"github.com/buoyantio/conduit-test/building_blocks/service"
	log "github.com/sirupsen/logrus"
)

const HttpEgressStrategyName = "http-egress"
const HttpEgressUrlToInvokeArgName = "url"

type HttpEgressStrategy struct {
	config *service.Config
}

func (s *HttpEgressStrategy) Do(_ context.Context, req *pb.TheRequest) (*pb.TheResponse, error) {

	urlToInvoke := s.config.ExtraArguments[HttpEgressUrlToInvokeArgName]

	log.Infof("Making GET request to [%s] for requestUid [%s]", urlToInvoke, req.GetRequestUid())
	httpResp, err := http.Get(urlToInvoke)
	if err != nil {
		return nil, err
	}

	log.Infof("Response from [%s] for requestUid [%s] was: %v", urlToInvoke, req.GetRequestUid(), httpResp)
	statusCode := httpResp.StatusCode
	if statusCode < 200 || statusCode > 299 {
		return nil, fmt.Errorf("unexpected status returned by [%s]for requestUid [%s]: %d", urlToInvoke, req.GetRequestUid(), statusCode)
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

func NewHttpEgress(config *service.Config, servers []service.Server, clients []service.Client) (service.Strategy, error) {
	if len(clients) != 0 || len(servers) == 0 {
		return nil, fmt.Errorf("strategy [%s] requires at least one server port and exactly zero downstream services, but was configured as: %+v", HttpEgressStrategyName, config)
	}

	return &HttpEgressStrategy{config: config}, nil
}
