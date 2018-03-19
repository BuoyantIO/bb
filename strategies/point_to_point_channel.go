package strategies

import (
	"context"
	"fmt"

	pb "github.com/buoyantio/conduit-test/gen"
	"github.com/buoyantio/conduit-test/service"
)

const PointToPointStrategyName = "point-to-point-channel"

type PointToPointChannelStrategy struct {
	clients []service.Client
}

func (s *PointToPointChannelStrategy) Do(_ context.Context, req *pb.TheRequest) (*pb.TheResponse, error) {
	client := s.clients[0]
	resp, err := client.Send(req)
	return resp, err
}

func NewPointToPointChannel(config *service.Config, servers []service.Server, clients []service.Client) (service.Strategy, error) {
	if len(clients) != 1 || len(servers) != 1 {
		return nil, fmt.Errorf("strategy [%s] requires exactly one server and one downstream service, but had clients [%v] servers [%v] and configured as: %+v", PointToPointStrategyName, clients, servers, config)
	}

	return &PointToPointChannelStrategy{
		clients: clients,
	}, nil
}
