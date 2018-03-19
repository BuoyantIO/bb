package protocols

import (
	"context"

	pb "github.com/buoyantio/bb/gen"
)

type stubStrategy struct {
	theRequestReceived  *pb.TheRequest
	theResponseToReturn *pb.TheResponse
	theErrorToReturn    error
}

func (h *stubStrategy) Do(ctx context.Context, req *pb.TheRequest) (*pb.TheResponse, error) {
	h.theRequestReceived = req
	return h.theResponseToReturn, h.theErrorToReturn
}
