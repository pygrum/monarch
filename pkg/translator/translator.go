package translator

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
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
	ListenPort        = 2000
	serviceAddress    = "http://localhost:20000" // The user-made translation service, localhost:20000
	toAgentEndpoint   = "/to"
	fromAgentEndpoint = "/from"
)

type translatorServer struct {
	rpcpb.UnimplementedTranslatorServer
	serviceClient *http.Client
	config        *config.ProjectConfig
}

func newServer() (*translatorServer, error) {
	// /config/royal.yaml is where the config file must be placed, as shown in dockerfile
	royal := filepath.Join("/config", "royal.yaml")
	royalConfig := config.ProjectConfig{}
	if err := config.YamlConfig(royal, &royalConfig); err != nil {
		return nil, err
	}
	c := &http.Client{Timeout: 10}
	return &translatorServer{config: &royalConfig, serviceClient: c}, nil
}

// sendServiceRequest sends a request to the translation service and returns the response body
func (s *translatorServer) sendServiceRequest(method, endpoint string, body []byte, receiver interface{}) error {
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

func (s *translatorServer) GetCmdDescriptions(context.Context, *rpcpb.DescriptionsRequest) (*rpcpb.DescriptionsReply, error) {
	reply := &rpcpb.DescriptionsReply{
		Descriptions: make([]*rpcpb.Description, len(s.config.CmdSchema)),
	}
	for i, cmd := range s.config.CmdSchema {
		description := &rpcpb.Description{
			Name:             cmd.Name,
			Opcode:           cmd.Opcode,
			Usage:            cmd.Usage,
			DescriptionShort: cmd.DescriptionShort,
			DescriptionLong:  cmd.DescriptionLong,
			NumArgs:          cmd.NArgs,
			Admin:            cmd.Admin,
		}
		reply.Descriptions[i] = description
	}
	return reply, nil
}

func (s *translatorServer) TranslateFrom(_ context.Context, r *rpcpb.Data) (*rpcpb.Reply, error) {
	b, err := json.Marshal(r)
	if err != nil {
		return nil, err
	}
	respObj := &fromAgentServiceResponse{}
	if err = s.sendServiceRequest(http.MethodPost, fromAgentEndpoint, b, respObj); err != nil {
		return nil, err
	}
	if !respObj.Success {
		return nil, fmt.Errorf("translation from reply failed with error: %v", errors.New(respObj.ErrorMsg))
	}
	return &respObj.Reply, nil
}

func (s *translatorServer) TranslateTo(_ context.Context, r *rpcpb.Request) (*rpcpb.Data, error) {
	req := &toAgentServiceRequest{
		AgentID:   r.AgentId,
		RequestID: r.RequestId,
		Opcode:    r.Opcode,
		Args:      r.Args,
	}
	b, err := json.Marshal(req)
	if err != nil {
		return nil, err
	}
	respObj := &toAgentServiceResponse{}
	if err = s.sendServiceRequest(http.MethodPost, toAgentEndpoint, b, respObj); err != nil {
		return nil, err
	}
	if !respObj.Success {
		return nil, fmt.Errorf("translation to request failed with error: %v", errors.New(string(respObj.Message)))
	}
	return &rpcpb.Data{Message: respObj.Message}, nil
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
	rpcpb.RegisterTranslatorServer(grpcServer, srv)
	return grpcServer.Serve(lis)
}
