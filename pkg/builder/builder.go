package builder

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"github.com/pygrum/monarch/pkg/config"
	"github.com/pygrum/monarch/pkg/rpcpb"
	"google.golang.org/grpc"
	"io"
	"net"
	"net/http"
	"path/filepath"
)

const (
	ListenPort     = 2001
	serviceAddress = "http://localhost:20000"
)

type builderServer struct {
	rpcpb.UnimplementedBuilderServer
	serviceClient *http.Client
	config        *config.ProjectConfig
}

func newServer() (*builderServer, error) {
	// /config/royal.yaml is where the config file must be placed, as shown in dockerfile
	royal := filepath.Join("/config", "royal.yaml")
	royalConfig := config.ProjectConfig{}
	if err := config.YamlConfig(royal, &royalConfig); err != nil {
		return nil, err
	}
	c := &http.Client{Timeout: 10}
	return &builderServer{config: &royalConfig, serviceClient: c}, nil
}

func (s *builderServer) BuildAgent(_ context.Context, r *rpcpb.BuildRequest) (*rpcpb.BuildReply, error) {
	buildReply := &rpcpb.BuildReply{}
	b, err := json.Marshal(r.GetParams())
	if err != nil {
		return nil, err
	}
	if err = s.sendServiceRequest(http.MethodPost, "/build", b, buildReply); err != nil {
		return nil, err
	}
	return buildReply, nil
}

// sendServiceRequest sends a request to the translation service and returns the response body
func (s *builderServer) sendServiceRequest(method, endpoint string, body []byte, receiver interface{}) error {
	httpReq, err := http.NewRequest(method, serviceAddress+endpoint, bytes.NewReader(body))
	if err != nil {
		return err
	}
	resp, err := s.serviceClient.Do(httpReq)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	b, _ := io.ReadAll(resp.Body)
	return json.Unmarshal(b, receiver)
}

func (s *builderServer) GetParams(context.Context, *rpcpb.ParamsRequest) (*rpcpb.ParamsReply, error) {
	reply := &rpcpb.ParamsReply{
		Params: make([]*rpcpb.Param, len(s.config.Builder.BuildArgs)),
	}
	for i, a := range s.config.Builder.BuildArgs {
		reply.Params[i] = &rpcpb.Param{
			Name:        a.Name,
			Description: a.Description,
			Required:    a.Required,
		}
	}
	return reply, nil
}

func Run() error {
	lis, err := net.Listen("tcp", fmt.Sprintf("0.0.0.0:%d", ListenPort))
	if err != nil {
		return err
	}
	var opts []grpc.ServerOption

	grpcServer := grpc.NewServer(opts...)
	srv, err := newServer()
	if err != nil {
		return err
	}
	rpcpb.RegisterBuilderServer(grpcServer, srv)
	return grpcServer.Serve(lis)
}
