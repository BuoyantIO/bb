package strategies

import (
	"context"
	"fmt"

	pb "github.com/buoyantio/bb/gen"
	"github.com/buoyantio/bb/service"
)

// PointToPointStrategyName is the user-friendly name of this strategy
const PointToPointStrategyName = "point-to-point-channel"

// PointToPointChannelStrategy is a strategy that takes a request and forwards it to a single downstream service.
type PointToPointChannelStrategy struct {
	clients []service.Client
}

func (s *PointToPointChannelStrategy) Do(_ context.Context, req *pb.TheRequest) (*pb.TheResponse, error) {
	client := s.clients[0]
	resp, err := client.Send(req)
	return resp, err
}

// NewPointToPointChannel creates a new PointToPointChannelStrategy
func NewPointToPointChannel(config *service.Config, servers []service.Server, clients []service.Client) (service.Strategy, error) {
	if len(clients) != 1 || len(servers) != 1 {
		return nil, fmt.Errorf("strategy [%s] requires exactly one server and one downstream service, but had clients [%v] servers [%v] and configured as: %+v", PointToPointStrategyName, clients, servers, config)
	}

	return &PointToPointChannelStrategy{
		clients: clients,
	}, nil
}
