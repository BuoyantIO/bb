package protocols

import (
	"context"
	"fmt"
	"net"

	pb "github.com/buoyantio/conduit-test/gen"
	"github.com/buoyantio/conduit-test/service"
	log "github.com/sirupsen/logrus"
	"google.golang.org/grpc"
)

type theGrpcServer struct {
	port           int
	serviceHandler *service.RequestHandler
}

func (s *theGrpcServer) GetId() string {
	return fmt.Sprintf("grpc-%d", s.port)
}

func (s *theGrpcServer) TheFunction(ctx context.Context, req *pb.TheRequest) (*pb.TheResponse, error) {
	resp, err := s.serviceHandler.Handle(ctx, req)
	log.Infof("Received gRPC request [%s] [%s] Returning response [%+v]", req.RequestUid, req, resp)
	return resp, err
}

type theGrpcClient struct {
	id         string
	conn       *grpc.ClientConn
	grpcClient pb.TheServiceClient
}

func (c *theGrpcClient) GetId() string {
	return c.id
}

func (c *theGrpcClient) Send(req *pb.TheRequest) (*pb.TheResponse, error) {
	return c.grpcClient.TheFunction(context.Background(), req)
}

func (c *theGrpcClient) Close() error {
	log.Debugf("Closing client [%s]", c.id)
	return c.conn.Close()
}

func NewGrpcServerIfConfigured(config *service.Config, serviceHandler *service.RequestHandler) (service.Server, error) {
	if config.GrpcServerPort == -1 {
		return nil, nil
	}

	grpcServerPort := config.GrpcServerPort
	lis, err := net.Listen("tcp", fmt.Sprintf(":%d", grpcServerPort))
	if err != nil {
		return nil, err
	}
	grpcServer := grpc.NewServer()

	theGrpcServer := &theGrpcServer{
		port:           grpcServerPort,
		serviceHandler: serviceHandler,
	}

	pb.RegisterTheServiceServer(grpcServer, theGrpcServer)
	log.Infof("gRPC server listening on port [%d]", grpcServerPort)
	go func() { grpcServer.Serve(lis) }()
	return theGrpcServer, nil
}

func NewGrpcClientsIfConfigured(config *service.Config) ([]service.Client, error) {
	clients := make([]service.Client, 0)
	for _, serverUrl := range config.GrpcDownstreamServers {
		conn, err := grpc.Dial(serverUrl, grpc.WithInsecure())
		if err != nil {
			return nil, err
		}

		client := pb.NewTheServiceClient(conn)
		clients = append(clients, &theGrpcClient{id: serverUrl, conn: conn, grpcClient: client})
	}

	return clients, nil
}
