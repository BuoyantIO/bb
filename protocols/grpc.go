package protocols

import (
	"context"
	"fmt"
	"net"

	pb "github.com/buoyantio/bb/gen"
	"github.com/buoyantio/bb/service"
	log "github.com/sirupsen/logrus"
	"google.golang.org/grpc"
)

type theGrpcServer struct {
	grpcServer     *grpc.Server
	port           int
	serviceHandler *service.RequestHandler
}

func (s *theGrpcServer) GetID() string {
	return fmt.Sprintf("grpc-%d", s.port)
}

func (s *theGrpcServer) Shutdown() error {
	log.Infof("Shutting down [%s]", s.GetID())
	s.grpcServer.GracefulStop()
	return nil
}

func (s *theGrpcServer) TheFunction(ctx context.Context, req *pb.TheRequest) (*pb.TheResponse, error) {
	resp, err := s.serviceHandler.Handle(ctx, req)
	log.Infof("Received gRPC request [%s] [%s] Returning response [%+v]", req.RequestUID, req, resp)
	return resp, err
}

type theGrpcClient struct {
	id         string
	conn       *grpc.ClientConn
	grpcClient pb.TheServiceClient
}

func (c *theGrpcClient) GetID() string {
	return c.id
}

func (c *theGrpcClient) Send(req *pb.TheRequest) (*pb.TheResponse, error) {
	return c.grpcClient.TheFunction(context.Background(), req)
}

func (c *theGrpcClient) Close() error {
	log.Debugf("Closing client [%s]", c.id)
	return c.conn.Close()
}

// NewGrpcServerIfConfigured returns a gRPC-backed Server
func NewGrpcServerIfConfigured(config *service.Config, serviceHandler *service.RequestHandler) (service.Server, error) {
	if config.GRPCServerPort == -1 {
		return nil, nil
	}

	grpcServerPort := config.GRPCServerPort
	lis, err := net.Listen("tcp", fmt.Sprintf(":%d", grpcServerPort))
	if err != nil {
		return nil, err
	}
	grpcServer := grpc.NewServer()

	theGrpcServer := &theGrpcServer{
		grpcServer:     grpcServer,
		port:           grpcServerPort,
		serviceHandler: serviceHandler,
	}

	pb.RegisterTheServiceServer(grpcServer, theGrpcServer)
	log.Infof("gRPC server listening on port [%d]", grpcServerPort)
	go func() { grpcServer.Serve(lis) }()
	return theGrpcServer, nil
}

// NewGrpcClientsIfConfigured takes in a Config and returns an instance of gRPC-backed Client for every configured gRPC
// downstream service
func NewGrpcClientsIfConfigured(config *service.Config) ([]service.Client, error) {
	clients := make([]service.Client, 0)

	for _, serverURL := range config.GRPCDownstreamServers {
		target := serverURL
		authority := ""
		clientID := serverURL
		if config.GRPCProxy != "" {
			target = config.GRPCProxy
			authority = serverURL
			clientID = config.GRPCProxy + " / " + serverURL
		}

		conn, err := grpc.Dial(
			target,
			grpc.WithTimeout(config.DownstreamConnectionTimeout),
			grpc.WithInsecure(),
			grpc.WithAuthority(authority),
		)
		if err != nil {
			return nil, err
		}

		client := pb.NewTheServiceClient(conn)
		clients = append(clients, &theGrpcClient{id: clientID, conn: conn, grpcClient: client})
	}

	return clients, nil
}
