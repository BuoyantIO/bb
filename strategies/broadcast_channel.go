package strategies

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"sync"

	pb "github.com/buoyantio/bb/gen"
	"github.com/buoyantio/bb/service"
	log "github.com/sirupsen/logrus"
)

// BroadcastChannelStrategyName is the user-friendly name of this strategy
const BroadcastChannelStrategyName = "broadcast-channel"

// BroadcastChannelStrategy is a strategy that will take in a request and broadact it to all downstream services.
type BroadcastChannelStrategy struct {
	clients []service.Client
}

// Do executes the request
func (s *BroadcastChannelStrategy) Do(_ context.Context, req *pb.TheRequest) (*pb.TheResponse, error) {
	numberOfRequestsToMake := len(s.clients)
	log.Infof("Starting broadcast to [%d] downstream services", numberOfRequestsToMake)
	var wg sync.WaitGroup
	wg.Add(numberOfRequestsToMake)

	allResults := make(chan interface{}, numberOfRequestsToMake)
	for _, client := range s.clients {
		go func(c service.Client) {
			log.Infof("Making request to [%s]", c.GetID())
			defer wg.Done()
			clientResp, err := c.Send(req)
			if err != nil {
				log.Errorf("Error when broadcasting request [%v] to client [%s]: %v", req, c.GetID(), err)
				allResults <- fmt.Errorf("downstream server [%s] returned error: %v", c.GetID(), err)
			} else {
				allResults <- clientResp
			}
		}(client)
	}
	wg.Wait()
	close(allResults)
	log.Info("Finished broadcast")

	allErrorMessages := make([]string, 0)
	allResponsePayloads := make([]string, 0)
	for result := range allResults {
		if err, isErr := result.(error); isErr {
			allErrorMessages = append(allErrorMessages, err.Error())
		} else {
			resp := result.(*pb.TheResponse)
			allResponsePayloads = append(allResponsePayloads, resp.Payload)
		}
	}

	var aggregatedResp *pb.TheResponse
	var aggregatedErrors error
	if len(allErrorMessages) > 0 {
		aggregatedErrors = errors.New(strings.Join(allErrorMessages, ","))
	} else {
		aggregatedResp = &pb.TheResponse{
			Payload: strings.Join(allResponsePayloads, ","),
		}
	}

	return aggregatedResp, aggregatedErrors
}

// NewBroadcastChannel creates a new BroadcastChannelStrategy
func NewBroadcastChannel(config *service.Config, servers []service.Server, clients []service.Client) (service.Strategy, error) {
	if len(clients) < 2 || len(servers) != 1 {
		var clientNames []string
		for _, client := range clients {
			clientNames = append(clientNames, client.GetID())
		}

		var serverNames []string
		for _, server := range servers {
			serverNames = append(clientNames, server.GetID())
		}

		return nil, fmt.Errorf("strategy [%s] requires exactly one server and more than one downstream services, but had clients [%s] servers [%s] and configured as: %+v", BroadcastChannelStrategyName, clientNames, serverNames, config)
	}

	return &BroadcastChannelStrategy{
		clients: clients,
	}, nil
}
