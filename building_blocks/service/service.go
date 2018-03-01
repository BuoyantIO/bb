package service

import (
	"context"
	"fmt"
	"math/rand"
	"time"

	pb "github.com/buoyantio/conduit-test/building_blocks/gen"
	log "github.com/sirupsen/logrus"
)

type Config struct {
	Id                          string
	GrpcServerPort              int
	H1ServerPort                int
	GrpcDownstreamServers       []string
	H1DownstreamServers         []string
	PercentageFailedRequests    int
	SleepInMillis               int
	FireAndForget               bool
	DownstreamConnectionTimeout time.Duration
	ExtraArguments              map[string]string
}

type Client interface {
	Close() error
	GetId() string
	Send(*pb.TheRequest) (*pb.TheResponse, error)
}

type fireAndForgetClient struct {
	underlyingClient Client
}

func (f *fireAndForgetClient) Close() error { return f.underlyingClient.Close() }

func (f *fireAndForgetClient) GetId() string { return f.underlyingClient.GetId() }

func (f *fireAndForgetClient) Send(req *pb.TheRequest) (*pb.TheResponse, error) {
	go func(c Client, req *pb.TheRequest) {
		log.Infof("Sending fire-and-forget request to [%s] for request uid [%s]", f.GetId(), req.RequestUid)
		response, err := c.Send(req)
		log.Infof("Response from fire-and-forget request to [%s] for request uid [%s] was: %s error %v", f.GetId(), req.RequestUid, response, err)
	}(f.underlyingClient, req)

	stubResponse := &pb.TheResponse{
		Payload: fmt.Sprintf("Stub response for fire-and-forget request to [%s] for request uid [%s]", f.GetId(), req.RequestUid),
	}
	return stubResponse, nil
}

func MakeFireAndForget(client Client) Client {
	return &fireAndForgetClient{underlyingClient: client}
}

type Server interface {
	GetId() string
}

type Strategy interface {
	Do(context.Context, *pb.TheRequest) (*pb.TheResponse, error)
}

type RequestHandler struct {
	Config   *Config
	Strategy Strategy
}

func (h *RequestHandler) Handle(ctx context.Context, req *pb.TheRequest) (*pb.TheResponse, error) {
	sleepForConfiguredTime(h)

	if shouldFailThisRequest(h) {
		return nil, fmt.Errorf("this error was injected by [%s]", h.Config.Id)
	} else {
		reqId := req.RequestUid

		resp, err := h.Strategy.Do(ctx, req)
		if resp != nil {
			resp.RequestUid = reqId
		}
		return resp, err
	}
}

func sleepForConfiguredTime(h *RequestHandler) {
	time.Sleep(time.Duration(int64(h.Config.SleepInMillis)) * time.Millisecond)
}

func shouldFailThisRequest(h *RequestHandler) bool {
	perc := h.Config.PercentageFailedRequests
	rnd := rand.Intn(100)
	return rnd < perc
}

type Service struct {
	Servers  []Server
	Clients  []Client
	Strategy Strategy
}

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
