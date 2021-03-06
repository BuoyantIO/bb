package cmd

import (
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"github.com/buoyantio/bb/protocols"
	"github.com/buoyantio/bb/service"
	"github.com/buoyantio/bb/strategies"
	log "github.com/sirupsen/logrus"
)

func buildServers(config *service.Config, handler *service.RequestHandler) ([]service.Server, error) {
	servers := make([]service.Server, 0)
	grpcServer, err := protocols.NewGrpcServerIfConfigured(config, handler)
	if err != nil {
		return nil, err
	}

	if grpcServer != nil {
		servers = append(servers, grpcServer)
	}

	httpServer, err := protocols.NewHTTPServerIfConfigured(config, handler)
	if err != nil {
		return nil, err
	}

	if httpServer != nil {
		servers = append(servers, httpServer)
	}

	return servers, nil
}

func buildClients(config *service.Config) ([]service.Client, error) {
	clients := make([]service.Client, 0)
	grpcClients, err := protocols.NewGrpcClientsIfConfigured(config)
	if err != nil {
		return nil, err
	}
	clients = append(clients, grpcClients...)

	httpClients, err := protocols.NewHTTPClientsIfConfigured(config)
	if err != nil {
		return nil, err
	}
	clients = append(clients, httpClients...)

	if config.FireAndForget {
		wrappedClients := make([]service.Client, 0)
		for _, c := range clients {
			wrappedClients = append(wrappedClients, service.MakeFireAndForget(c))
		}
		clients = wrappedClients
	}

	return clients, err
}

func newService(config *service.Config, strategyName string) (*service.Service, error) {

	handler := service.NewRequestHandler(config)

	servers, err := buildServers(config, handler)
	if err != nil {
		return nil, err
	}

	clients, err := buildClients(config)
	if err != nil {
		return nil, err
	}

	strategy, err := newStrategyByName(strategyName, config, servers, clients)
	if err != nil {
		log.Fatalln(err)
	}

	//TODO: this is awful as there's a circular dep between server and strategy
	handler.Strategy = strategy

	service := &service.Service{
		Strategy: strategy,
		Servers:  servers,
		Clients:  clients,
	}

	log.Infof("Process configured as: %+v", service)

	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt, syscall.SIGTERM)
	log.Infof("Service [%s] is ready and waiting for incoming connections", config.ID)

	select {
	case <-stop:
		log.Infof("Stopping service [%s] due to interrupt", config.ID)
	case <-handler.Stopping():
		log.Infof("Stopping service [%s] due to handler", config.ID)
	}

	for _, server := range servers {
		server.Shutdown()
	}

	return service, nil
}

type strategyConstructor func(*service.Config, []service.Server, []service.Client) (service.Strategy, error)

var strategyByName = map[string]strategyConstructor{
	strategies.PointToPointStrategyName:     strategies.NewPointToPointChannel,
	strategies.BroadcastChannelStrategyName: strategies.NewBroadcastChannel,
	strategies.TerminusStrategyName:         strategies.NewTerminusStrategy,
	strategies.HTTPEgressStrategyName:       strategies.NewHTTPEgress,
}

func newStrategyByName(strategyName string, config *service.Config, servers []service.Server, clients []service.Client) (service.Strategy, error) {
	strategyConstructor := strategyByName[strategyName]
	if strategyConstructor == nil {
		return nil, fmt.Errorf("no strategy named [%s]", strategyName)
	}

	return strategyConstructor(config, servers, clients)
}
