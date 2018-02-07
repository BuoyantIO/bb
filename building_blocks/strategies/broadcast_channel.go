package strategies

import (
	"context"
	"errors"
	"fmt"
	"strings"

	pb "github.com/buoyantio/conduit-test/building_blocks/gen"
	"github.com/buoyantio/conduit-test/building_blocks/service"
)

const BroadcastChannelStrategyName = "broadcast-channel"

type BroadcastChannelStrategy struct {
	clients []service.Client
}

func (s *BroadcastChannelStrategy) Do(_ context.Context, req *pb.TheRequest) (*pb.TheResponse, error) {
	allResponsePayloads := []string{}
	allErrors := []string{}

	for _, client := range s.clients {
		clientResp, err := client.Send(req)
		if err != nil {
			allErrors = append(allErrors, err.Error())
		}

		if clientResp != nil {
			allResponsePayloads = append(allResponsePayloads, clientResp.Payload)
		}
	}

	var aggregatedResp *pb.TheResponse
	var aggregatedErrors error

	if len(allErrors) > 0 {
		aggregatedErrors = errors.New(strings.Join(allErrors, ","))
	} else {
		aggregatedResp = &pb.TheResponse{
			Payload: strings.Join(allResponsePayloads, ","),
		}
	}

	return aggregatedResp, aggregatedErrors
}

func NewBroadcastChannel(config *service.Config, servers []service.Server, clients []service.Client) (service.Strategy, error) {
	if len(clients) < 2 || len(servers) != 1 {
		var clientNames []string
		for _, client := range clients {
			clientNames = append(clientNames, client.GetId())
		}

		var serverNames []string
		for _, server := range servers {
			serverNames = append(clientNames, server.GetId())
		}

		return nil, fmt.Errorf("strategy [%s] requires exactly one server and more than one downstream services, but had clients [%s] servers [%s] and configured as: %+v", BroadcastChannelStrategyName, clientNames, serverNames, config)
	}

	return &BroadcastChannelStrategy{
		clients: clients,
	}, nil
}
