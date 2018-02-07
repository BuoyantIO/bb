package protocols

import (
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	pb "github.com/buoyantio/conduit-test/building_blocks/gen"
	"github.com/buoyantio/conduit-test/building_blocks/service"
	"github.com/gogo/protobuf/jsonpb"
	"github.com/golang/protobuf/proto"
	log "github.com/sirupsen/logrus"
)

var marshaller = &jsonpb.Marshaler{}

type theHttpServer struct {
	port int
}

type httpHandler struct {
	serviceHandler *service.RequestHandler
}

func (s *theHttpServer) GetId() string {
	return fmt.Sprintf("h1-%d", s.port)
}

func (h *httpHandler) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	var protoReq pb.TheRequest

	if req.ContentLength > 0 {
		r, err := unmarshalProtoRequest(req)
		if err != nil {
			dealWithErrorDuringHandling(w, fmt.Errorf("error unmarshalling the request: %v", err))
			return
		}
		protoReq = r
	} else {
		protoReq = pb.TheRequest{
			RequestUid: newRequestUid("http", h.serviceHandler.Config),
		}
		log.Info("Received request with empty body, creating new request")
	}

	protoResponse, err := h.serviceHandler.Handle(req.Context(), &protoReq)
	if err != nil {
		dealWithErrorDuringHandling(w, fmt.Errorf("error in handler: %v", err))
		return
	}

	log.Infof("Received HTTP request [%s] [%s %s] Body [%+v] Returning response [%+v]", protoReq.RequestUid, req.Method, req.URL, protoReq, protoResponse)

	if err = marshalProtoResponse(w, protoResponse); err != nil {
		dealWithErrorDuringHandling(w, fmt.Errorf("error marshalling the response: %v", err))
		return
	}
}

type httpClient struct {
	id        string
	serverUrl string
}

func (c *httpClient) Close() error { return nil }

func (c *httpClient) GetId() string { return c.id }

func (c *httpClient) Send(req *pb.TheRequest) (*pb.TheResponse, error) {
	json, err := marshallProtobufToJson(req)
	if err != nil {
		return nil, err
	}
	resp, err := http.Post(c.serverUrl, "application/json", strings.NewReader(json))
	if err != nil {
		return nil, err
	}
	var protoResp pb.TheResponse
	err = unmarshalJsonToProtobuf(resp.Body, &protoResp)
	return &protoResp, err
}

func newRequestUid(inboundType string, config *service.Config) string {
	return fmt.Sprintf("in:%s-sid:%s-%d", inboundType, config.Id, time.Now().Nanosecond())
}

func marshallProtobufToJson(msg proto.Message) (string, error) {
	json, err := marshaller.MarshalToString(msg)
	if err != nil {
		return "", err
	}
	return json, nil
}

func marshalProtoResponse(httpResp http.ResponseWriter, protoResp proto.Message) error {
	jsonResponse, err := marshallProtobufToJson(protoResp)
	if err != nil {
		return err
	}

	_, err = fmt.Fprint(httpResp, jsonResponse)
	if err != nil {
		return err
	}
	return nil
}

func unmarshalProtoRequest(httpReq *http.Request) (pb.TheRequest, error) {
	var protoReq pb.TheRequest
	err := unmarshalJsonToProtobuf(httpReq.Body, &protoReq)
	return protoReq, err
}

func unmarshalJsonToProtobuf(r io.Reader, out proto.Message) error {
	return jsonpb.Unmarshal(r, out)
}

func dealWithErrorDuringHandling(w http.ResponseWriter, err error) {
	log.Errorf("Error while handling HTTP request: %v", err)
	http.Error(w, err.Error(), http.StatusInternalServerError)
}

func newHttpHandler(serviceHandler *service.RequestHandler) *httpHandler {
	return &httpHandler{
		serviceHandler: serviceHandler,
	}
}

func NewHttpServerIfConfigured(config *service.Config, serviceHandler *service.RequestHandler) (service.Server, error) {
	if config.H1ServerPort == -1 {
		return nil, nil
	}

	handler := newHttpHandler(serviceHandler)
	go func() {
		log.Infof("HTTP 1.1 server listening on port [%d]", config.H1ServerPort)
		http.ListenAndServe(fmt.Sprintf(":%d", config.H1ServerPort), handler)
	}()

	return &theHttpServer{
		port: config.H1ServerPort,
	}, nil
}

func NewHttpClientsIfConfigured(config *service.Config) ([]service.Client, error) {
	clients := make([]service.Client, 0)
	for _, serverUrl := range config.H1DownstreamServers {
		clients = append(clients, &httpClient{
			id:        serverUrl,
			serverUrl: serverUrl,
		})
	}

	return clients, nil
}
