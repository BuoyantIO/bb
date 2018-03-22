package service

import (
	"context"
	"fmt"
	"math/rand"
	"time"

	pb "github.com/buoyantio/bb/gen"
	log "github.com/sirupsen/logrus"
)

// Config holds the ,ain configuration for this service.
type Config struct {
	ID                          string
	GRPCServerPort              int
	H1ServerPort                int
	GRPCDownstreamServers       []string
	H1DownstreamServers         []string
	PercentageFailedRequests    int
	SleepInMillis               int
	FireAndForget               bool
	DownstreamConnectionTimeout time.Duration
	ExtraArguments              map[string]string
}

// Client is an abstraction representing a client connection to each downstream service.
type Client interface {
	Close() error
	GetID() string
	Send(*pb.TheRequest) (*pb.TheResponse, error)
}

type fireAndForgetClient struct {
	underlyingClient Client
}

func (f *fireAndForgetClient) Close() error { return f.underlyingClient.Close() }

func (f *fireAndForgetClient) GetID() string { return f.underlyingClient.GetID() }

func (f *fireAndForgetClient) Send(req *pb.TheRequest) (*pb.TheResponse, error) {
	go func(c Client, req *pb.TheRequest) {
		log.Infof("Sending fire-and-forget request to [%s] for request UID [%s]", f.GetID(), req.RequestUID)
		response, err := c.Send(req)
		log.Infof("Response from fire-and-forget request to [%s] for request UID [%s] was: %s error %v", f.GetID(), req.RequestUID, response, err)
	}(f.underlyingClient, req)

	stubResponse := &pb.TheResponse{
		Payload: fmt.Sprintf("Stub response for fire-and-forget request to [%s] for request UID [%s]", f.GetID(), req.RequestUID),
	}
	return stubResponse, nil
}

// MakeFireAndForget creates a new Client that will send requests and not wait for a response.
func MakeFireAndForget(client Client) Client {
	return &fireAndForgetClient{underlyingClient: client}
}

// Server is an abstraction representing each server made available to receive inbound connections.
type Server interface {
	GetID() string
}

// Strategy is the algorithm applied by this service when it receives requests (c.f. http://wiki.c2.com/?StrategyPattern)
type Strategy interface {
	Do(context.Context, *pb.TheRequest) (*pb.TheResponse, error)
}

// RequestHandler is a protocol-independent request/response handler interface
type RequestHandler struct {
	Config   *Config
	Strategy Strategy
}

// Handle takes in a request, processes it accordingly to its Strategy, an returns the response.
func (h *RequestHandler) Handle(ctx context.Context, req *pb.TheRequest) (*pb.TheResponse, error) {
	sleepForConfiguredTime(h)

	if shouldFailThisRequest(h) {
		return nil, fmt.Errorf("this error was injected by [%s]", h.Config.ID)
	}

	reqID := req.RequestUID

	resp, err := h.Strategy.Do(ctx, req)
	if resp != nil {
		resp.RequestUID = reqID
	}
	return resp, err
}

func sleepForConfiguredTime(h *RequestHandler) {
	time.Sleep(time.Duration(int64(h.Config.SleepInMillis)) * time.Millisecond)
}

func shouldFailThisRequest(h *RequestHandler) bool {
	perc := h.Config.PercentageFailedRequests
	rnd := rand.Intn(100)
	return rnd < perc
}

// Service is the aggregate of all Client, Server, and the Strategy.
type Service struct {
	Servers  []Server
	Clients  []Client
	Strategy Strategy
}

// Close closes any open connections with Clients.
func (s *Service) Close() error {
	errors := make([]error, 0)
	for _, c := range s.Clients {
		err := c.Close()
		if err != nil {
			log.Errorln(err)
			errors = append(errors, err)
		}
	}

	if len(errors) > 0 {
		return fmt.Errorf("errors found closing connections: %+v", errors)
	}

	return nil
}
