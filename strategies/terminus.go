package strategies

import (
	"context"
	"fmt"
	"time"

	pb "github.com/buoyantio/bb/gen"
	"github.com/buoyantio/bb/service"
)

// TerminusStrategyName Name is the user-friendly name of this strategy
const TerminusStrategyName = "terminus"

// TerminusResponseTextArgName is the parameter used to supply the text to be returned by the TerminusStrategy
const TerminusResponseTextArgName = "response-text"

// TerminusStrategy is a strategy that always returns a pre-configured text message as the response to any requests.
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

// NewTerminusStrategy creates a new TerminusStrategy
func NewTerminusStrategy(config *service.Config, servers []service.Server, clients []service.Client) (service.Strategy, error) {
	if len(clients) != 0 || len(servers) == 0 {
		return nil, fmt.Errorf("strategy [%s] requires at least one server port and exactly zero downstream services, but was configured as: %+v", TerminusStrategyName, config)
	}

	return &TerminusStrategy{config: config}, nil
}
