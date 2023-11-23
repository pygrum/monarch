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
	"log"
	"net"
	"net/http"
	"os"
	"os/exec"
	"time"
)

const (
	ListenPort     = 2000
	serviceAddress = "http://127.0.0.1:20000"
)

type builderServer struct {
	rpcpb.UnimplementedBuilderServer
	serviceClient *http.Client
	config        *config.ProjectConfig
}

func newServer(conf string) (*builderServer, error) {
	// /config/royal.yaml is where the config file must be placed, as shown in dockerfile
	royalConfig := config.ProjectConfig{}
	if err := config.YamlConfig(conf, &royalConfig); err != nil {
		return nil, err
	}
	c := &http.Client{Timeout: 10 * time.Second}
	return &builderServer{config: &royalConfig, serviceClient: c}, nil
}

func (s *builderServer) BuildAgent(_ context.Context, r *rpcpb.BuildRequest) (*rpcpb.BuildReply, error) {
	buildReply := &rpcpb.BuildReply{}
	r.Options["src_directory"] = s.config.Builder.SourceDir
	b, err := json.Marshal(r.GetOptions())
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

func (s *builderServer) GetOptions(context.Context, *rpcpb.OptionsRequest) (*rpcpb.OptionsReply, error) {
	reply := &rpcpb.OptionsReply{
		Options: make([]*rpcpb.Option, len(s.config.Builder.BuildArgs)),
	}
	for i, a := range s.config.Builder.BuildArgs {
		reply.Options[i] = &rpcpb.Option{
			Name:        a.Name,
			Description: a.Description,
			Required:    a.Required,
			Default:     a.Default,
		}
	}
	return reply, nil
}

func (s *builderServer) GetCommands(context.Context, *rpcpb.DescriptionsRequest) (*rpcpb.DescriptionsReply, error) {
	reply := &rpcpb.DescriptionsReply{
		Descriptions: make([]*rpcpb.Description, len(s.config.CmdSchema)),
	}
	for i, sch := range s.config.CmdSchema {
		reply.Descriptions[i] = &rpcpb.Description{
			Name:             sch.Name,
			Opcode:           sch.Opcode,
			Usage:            sch.Usage,
			DescriptionShort: sch.DescriptionShort,
			DescriptionLong:  sch.DescriptionLong,
			NumArgs:          sch.NArgs,
			Admin:            sch.Admin,
		}
	}
	return reply, nil
}

func Run() error {
	if len(os.Args) < 3 {
		return fmt.Errorf("usage: %s /path/to/royal.yml /path/to/service.py /path/to/requirements.txt",
			os.Args[0])
	}
	var cerr bytes.Buffer
	if len(os.Args) > 3 {
		cmd := exec.Command("pip3", "install", "-r", os.Args[3])
		cmd.Stderr = &cerr
		if err := cmd.Run(); err != nil {
			return fmt.Errorf("failed to install service dependencies: %v: %v", err, cerr.String())
		}
	}
	go func() {
		cmd := exec.Command("python3", os.Args[2])
		cmd.Stderr = &cerr
		if err := cmd.Run(); err != nil {
			log.Fatalf("failed to start service service: %v: %v", err, cerr.String())
		}
	}()
	lis, err := net.Listen("tcp", fmt.Sprintf("0.0.0.0:%d", ListenPort))
	if err != nil {
		return err
	}
	var opts []grpc.ServerOption

	grpcServer := grpc.NewServer(opts...)
	srv, err := newServer(os.Args[1])
	if err != nil {
		return err
	}
	rpcpb.RegisterBuilderServer(grpcServer, srv)
	// deliberately blocking
	return grpcServer.Serve(lis)
}
