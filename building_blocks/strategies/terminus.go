package strategies

import (
	"context"
	"fmt"
	"time"

	pb "github.com/buoyantio/conduit-test/building_blocks/gen"
	"github.com/buoyantio/conduit-test/building_blocks/service"
)

const TerminusStrategyName = "terminus"
const TerminusResponseTextArgName = "response-text"

type TerminusStrategy struct {
	config *service.Config
}

func (s *TerminusStrategy) Do(_ context.Context, req *pb.TheRequest) (*pb.TheResponse, error) {
	messageToReturn := fmt.Sprintf("terminus at [%d]", time.Now().Nanosecond())
	if s.config.ExtraArguments[TerminusResponseTextArgName] != "" {
		messageToReturn = s.config.ExtraArguments[TerminusResponseTextArgName]
	}

	resp := pb.TheResponse{
		Payload: messageToReturn,
	}
	return &resp, nil
}

func NewTerminusStrategy(config *service.Config, servers []service.Server, clients []service.Client) (service.Strategy, error) {
	if len(clients) != 0 || len(servers) == 0 {
		return nil, fmt.Errorf("strategy [%s] requires at least one server port and exactly zero downstream services, but was configured as: %+v", TerminusStrategyName, config)
	}

	return &TerminusStrategy{config: config}, nil
}
