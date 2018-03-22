package protocols

import (
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"strings"
	"time"

	pb "github.com/buoyantio/bb/gen"
	"github.com/buoyantio/bb/service"
	"github.com/gogo/protobuf/jsonpb"
	"github.com/golang/protobuf/proto"
	log "github.com/sirupsen/logrus"
)

var marshaller = &jsonpb.Marshaler{}

type theHTTPServer struct {
	port int
}

type httpHandler struct {
	serviceHandler *service.RequestHandler
}

func (s *theHTTPServer) GetID() string {
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
		newRequestUID := newRequestUID("http", h.serviceHandler.Config)
		log.Infof("Received request with empty body, assigning new request UID [%s] to it", newRequestUID)
		protoReq = pb.TheRequest{
			RequestUID: newRequestUID,
		}
	}

	protoResponse, err := h.serviceHandler.Handle(req.Context(), &protoReq)
	if err != nil {
		dealWithErrorDuringHandling(w, fmt.Errorf("error handling http request: %v", err))
		return
	}

	log.Infof("Received HTTP request [%s] [%s %s] Body [%+v] Returning response [%+v]", protoReq.RequestUID, req.Method, req.URL, protoReq, protoResponse)

	if err = marshalProtoResponse(w, protoResponse); err != nil {
		dealWithErrorDuringHandling(w, fmt.Errorf("error marshalling the response: %v", err))
		return
	}
}

type httpClient struct {
	id                        string
	serverURL                 string
	clientForDownsteamServers *http.Client
}

func (c *httpClient) Close() error { return nil }

func (c *httpClient) GetID() string { return c.id }

func (c *httpClient) Send(req *pb.TheRequest) (*pb.TheResponse, error) {
	json, err := marshallProtobufToJSON(req)
	if err != nil {
		return nil, err
	}

	resp, err := c.clientForDownsteamServers.Post(c.serverURL, "application/json", strings.NewReader(json))
	if err != nil {
		return nil, err
	}

	var protoResp pb.TheResponse
	defer resp.Body.Close()
	err = unmarshalJSONToProtobuf(resp.Body, &protoResp)

	return &protoResp, err
}

func newRequestUID(inboundType string, config *service.Config) string {
	return fmt.Sprintf("in:%s-sid:%s-%d", inboundType, config.ID, time.Now().Nanosecond())
}

func marshallProtobufToJSON(msg proto.Message) (string, error) {
	json, err := marshaller.MarshalToString(msg)
	if err != nil {
		return "", err
	}
	return json, nil
}

func marshalProtoResponse(httpResp http.ResponseWriter, protoResp proto.Message) error {
	jsonResponse, err := marshallProtobufToJSON(protoResp)
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
	err := unmarshalJSONToProtobuf(httpReq.Body, &protoReq)
	return protoReq, err
}

func unmarshalJSONToProtobuf(r io.Reader, out proto.Message) error {
	bytes, err := ioutil.ReadAll(r)
	if err != nil {
		return err
	}

	bodyAsString := string(bytes)
	err = jsonpb.UnmarshalString(bodyAsString, out)
	if err != nil {
		return errors.New(bodyAsString)
	}

	return nil
}

func dealWithErrorDuringHandling(w http.ResponseWriter, err error) {
	log.Errorf("Error while handling HTTP request: %v", err)
	http.Error(w, err.Error(), http.StatusInternalServerError)
}

func newHTTPHandler(serviceHandler *service.RequestHandler) *httpHandler {
	return &httpHandler{
		serviceHandler: serviceHandler,
	}
}

// NewHTTPServerIfConfigured returns a HTTP-backed Server
func NewHTTPServerIfConfigured(config *service.Config, serviceHandler *service.RequestHandler) (service.Server, error) {
	if config.H1ServerPort == -1 {
		return nil, nil
	}

	handler := newHTTPHandler(serviceHandler)
	go func() {
		log.Infof("HTTP 1.1 server listening on port [%d]", config.H1ServerPort)
		http.ListenAndServe(fmt.Sprintf(":%d", config.H1ServerPort), handler)
	}()

	return &theHTTPServer{
		port: config.H1ServerPort,
	}, nil
}

// NewHTTPClientsIfConfigured takes in a Config and returns an instance of HTTP-backed Client for every configured HTTP
// downstream service
func NewHTTPClientsIfConfigured(config *service.Config) ([]service.Client, error) {
	clients := make([]service.Client, 0)

	httpClientToUse := &http.Client{
		Timeout: config.DownstreamConnectionTimeout,
	}

	for _, serverURL := range config.H1DownstreamServers {
		clients = append(clients, &httpClient{
			id:                        serverURL,
			serverURL:                 serverURL,
			clientForDownsteamServers: httpClientToUse,
		})
	}

	return clients, nil
}
