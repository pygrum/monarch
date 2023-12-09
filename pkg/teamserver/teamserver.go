package teamserver

import (
	"context"
	"crypto/tls"
	"crypto/x509"
	"errors"
	"fmt"
	"github.com/pygrum/monarch/pkg/types"
	"math"
	"net"
	"os"
	"path/filepath"
	"slices"
	"strings"
	"time"

	"github.com/pygrum/monarch/pkg/config"
	"github.com/pygrum/monarch/pkg/crypto"
	"github.com/pygrum/monarch/pkg/db"
	"github.com/pygrum/monarch/pkg/docker"
	"github.com/pygrum/monarch/pkg/handler/http"
	"github.com/pygrum/monarch/pkg/install"
	"github.com/pygrum/monarch/pkg/protobuf/builderpb"
	"github.com/pygrum/monarch/pkg/protobuf/clientpb"
	"github.com/pygrum/monarch/pkg/protobuf/rpcpb"
	"github.com/pygrum/monarch/pkg/transport"
	"github.com/pygrum/monarch/pkg/utils"
	"github.com/sirupsen/logrus"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/credentials/insecure"
)

type MonarchServer struct {
	rpcpb.UnimplementedMonarchServer
	builderClients map[string]rpcpb.BuilderClient
}

var (
	grpcServer *grpc.Server
)

func New() (*MonarchServer, error) {
	return &MonarchServer{
		builderClients: make(map[string]rpcpb.BuilderClient),
	}, nil
}

func (s *MonarchServer) Agents(_ context.Context, req *clientpb.AgentRequest) (*clientpb.Agents, error) {
	pbAgents := &clientpb.Agents{}
	var agents []db.Agent
	if len(req.AgentId) > 0 {
		if err := db.FindConditional("agent_id IN ?", req.AgentId, &agents); err != nil {
			return nil, fmt.Errorf("failed to retrieve the specified agents: %v", err)
		}
		if len(agents) == 0 {
			if err := db.FindConditional("name IN ?", req.AgentId, &agents); err != nil {
				return nil, fmt.Errorf("failed to retrieve the specified agents: %v", err)
			}
		}
	} else {
		if err := db.Find(&agents); err != nil {
			return nil, fmt.Errorf("failed to find agent(s): %v", err)
		}
	}
	for _, a := range agents {
		pbAgents.Agents = append(pbAgents.Agents, &clientpb.Agent{
			AgentId:   a.AgentID,
			Name:      a.Name,
			Version:   a.Version,
			OS:        a.OS,
			Arch:      a.Arch,
			Host:      a.Host,
			Port:      a.Port,
			Builder:   a.Builder,
			File:      a.File,
			CreatedAt: a.CreatedAt.Format(time.RFC850),
		})
	}
	return pbAgents, nil
}

func (s *MonarchServer) NewAgent(_ context.Context, agent *clientpb.Agent) (*clientpb.Empty, error) {
	a := &db.Agent{}
	if err := db.FindOneConditional("name = ?", agent.Name, a); err == nil {
		// just to check that we actually returned sum
		if a.Name == agent.Name {
			return nil, fmt.Errorf("duplicate agent names - choose a different name, or delete the other agent")
		}
	}
	a = &db.Agent{
		AgentID:   agent.AgentId,
		Name:      agent.Name,
		Version:   agent.Version,
		OS:        agent.OS,
		Arch:      agent.Arch,
		Host:      agent.Host,
		Port:      agent.Port,
		Builder:   agent.Builder,
		File:      agent.File,
		CreatedBy: agent.CreatedBy,
	}
	if err := db.Create(a); err != nil {
		return nil, fmt.Errorf("failed to save agent instance: %v", err)
	}
	return &clientpb.Empty{}, nil
}

func (s *MonarchServer) RmAgents(_ context.Context, req *clientpb.AgentRequest) (*clientpb.Empty, error) {
	var agents []db.Agent
	if err := db.FindConditional("agent_id IN ?", req.AgentId, &agents); err != nil {
		return nil, fmt.Errorf("failed to retrieve the specified agents: %v", err)
	}
	if len(agents) == 0 {
		if err := db.FindConditional("name IN ?", req.AgentId, &agents); err != nil {
			return nil, fmt.Errorf("failed to retrieve the specified agents: %v", err)
		}
	}
	if len(agents) == 0 {
		return nil, fmt.Errorf("no agents with the provided names exist")
	}
	for _, agent := range agents {
		if err := db.DeleteOne(&agent); err != nil {
			return nil, fmt.Errorf("failed to delete %s: %v", agent.Name, err)
		}
	}
	return &clientpb.Empty{}, nil
}

func (s *MonarchServer) Builders(_ context.Context, req *clientpb.BuilderRequest) (*clientpb.Builders, error) {
	var builders []db.Builder
	pbBuilders := &clientpb.Builders{}
	if len(req.BuilderId) == 0 {
		if err := db.Find(&builders); err != nil {
			return nil, fmt.Errorf("failed to retrieve installed builders: %v", err)
		}
	} else {
		if err := db.FindConditional("builder_id IN ?", req.BuilderId, &builders); err != nil {
			return nil, fmt.Errorf("failed to retrieve the specified builders: %v", err)
		}
		if len(builders) == 0 {
			if err := db.FindConditional("name IN ?", req.BuilderId, &builders); err != nil {
				return nil, fmt.Errorf("failed to retrieve the specified builders: %v", err)
			}
		}
	}
	for _, b := range builders {
		pbBuilders.Builders = append(pbBuilders.Builders, &clientpb.Builder{
			BuilderId:    b.BuilderID,
			CreatedAt:    b.CreatedAt.Format(time.RFC850),
			UpdatedAt:    b.UpdatedAt.Format(time.RFC850),
			Name:         b.Name,
			Version:      b.Version,
			Author:       b.Author,
			Url:          b.Url,
			Supported_OS: b.SupportedOS,
			InstalledAt:  b.InstalledAt,
			ImageId:      b.ImageID,
			ContainerId:  b.ContainerID,
		})
	}
	return pbBuilders, nil
}

func (s *MonarchServer) Profiles(_ context.Context, req *clientpb.ProfileRequest) (*clientpb.Profiles, error) {
	var profiles []db.Profile
	pbProfiles := &clientpb.Profiles{}
	if len(req.Name) > 0 {
		if err := db.Where("name IN ? AND builder_id = ?", req.Name, req.BuilderId).Find(&profiles); err != nil {
			return nil, fmt.Errorf("failed to find profiles(s): %v", err)
		}
	} else {
		if err := db.Where("builder_id = ?", req.BuilderId).Find(&profiles).Error; err != nil {
			return nil, fmt.Errorf("failed to find profiles(s): %v", err)
		}
	}
	for _, p := range profiles {
		pbProfiles.Profiles = append(pbProfiles.Profiles, &clientpb.Profile{
			Id:        int32(p.ID),
			CreatedAt: p.CreatedAt.Format(time.RFC850),
			Name:      p.Name,
			BuilderId: p.BuilderID,
		})
	}
	return pbProfiles, nil
}

func (s *MonarchServer) SaveProfile(_ context.Context, req *clientpb.SaveProfileRequest) (*clientpb.Empty, error) {
	profile := &db.Profile{}
	if db.Where("name = ? AND builder_id = ?", req.Name, req.BuilderId).Find(&profile); len(profile.Name) != 0 {
		return nil, fmt.Errorf("a profile for this build named '%s' already exists", req.Name)
	}
	profile = &db.Profile{
		Name:      req.Name,
		BuilderID: req.BuilderId,
	}
	var records []db.ProfileRecord
	for k, v := range req.Options {
		record := db.ProfileRecord{
			Profile: req.Name,
			Name:    k,
			Value:   v,
		}
		records = append(records, record)
	}
	if err := db.Create(profile); err != nil {
		return nil, fmt.Errorf("failed to create new profile: %v", err)
	}
	if err := db.Create(records); err != nil {
		return nil, fmt.Errorf("failed to save profile values: %v", err)
	}
	return &clientpb.Empty{}, nil
}

func (s *MonarchServer) LoadProfile(_ context.Context, req *clientpb.SaveProfileRequest) (*clientpb.ProfileData, error) {
	profile := &db.Profile{}
	profileData := &clientpb.ProfileData{}
	if err := db.Where("name = ? AND builder_id = ?", req.Name, req.BuilderId).Find(profile).Error; err != nil {
		return nil, fmt.Errorf("failed to find %s: %v", req.Name, err)
	}
	var records []db.ProfileRecord
	if err := db.FindConditional("profile = ?", req.Name, &records); err != nil {
		return nil, fmt.Errorf("failed to get profile values: %v", err)
	}
	profileData.Profile = &clientpb.Profile{
		Id:        int32(profile.ID),
		Name:      profile.Name,
		CreatedAt: profile.CreatedAt.Format(time.RFC850),
		BuilderId: profile.BuilderID,
	}
	for _, r := range records {
		if slices.Contains(req.Immutables, r.Name) {
			continue
		}
		profileData.Records = append(profileData.Records, &clientpb.ProfileRecord{
			Profile: profile.Name,
			Name:    r.Name,
			Value:   r.Value,
		})
	}
	return profileData, nil
}

func (s *MonarchServer) RmProfiles(ctx context.Context, req *clientpb.ProfileRequest) (*clientpb.Empty, error) {
	profiles, err := s.Profiles(ctx, req)
	if err != nil {
		return nil, err
	}
	var records []db.ProfileRecord
	if err := db.FindConditional("profile IN ?", req.Name, &records); err != nil {
		return nil, fmt.Errorf("failed to find profile values: %v", err)
	}
	if err := db.Delete(records); err != nil {
		return nil, fmt.Errorf("failed to delete profile values: %v", err)
	}
	if err := db.Delete(profiles); err != nil {
		return nil, fmt.Errorf("failed to delete profiles: %v", err)
	}
	return &clientpb.Empty{}, nil
}

func (s *MonarchServer) newBuilderClient(bid string) (rpcpb.BuilderClient, error) {
	if len(bid) == 0 {
		return nil, errors.New("agentID+builderID pair was not passed to packet")
	}
	if client, ok := s.builderClients[bid]; ok {
		return client, nil
	}
	realBid := bid[16:] // first 16 bytes are the agent ID, which we can ignore
	builderRPC, err := docker.RPCAddress(docker.Cli, context.Background(), realBid)
	if err != nil {
		return nil, err
	}
	conn, err := grpc.Dial(builderRPC, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		return nil, err
	}
	client := rpcpb.NewBuilderClient(conn)
	s.builderClients[bid] = client
	return client, nil
}

func (s *MonarchServer) killBuilderClient(bid string) error {
	if len(bid) == 0 {
		return errors.New("agentID+builderID pair was not passed to context with key 'builder_id'")
	}
	delete(s.builderClients, bid)
	return nil
}

// Options returns build options for each request to start the build process
// A builder client MUST be sent via ctx otherwise an error is returned.
func (s *MonarchServer) Options(ctx context.Context, o *builderpb.OptionsRequest) (*builderpb.OptionsReply, error) {
	client, err := s.newBuilderClient(o.BuilderId)
	if err != nil {
		return nil, fmt.Errorf("failed to get builder client: %v", err)
	}
	return client.GetOptions(ctx, o)
}

// Build returns a reply for a build request issued by a client.
// A builder client MUST be sent via ctx otherwise an error is returned.
func (s *MonarchServer) Build(ctx context.Context, req *builderpb.BuildRequest) (*builderpb.BuildReply, error) {
	client, err := s.newBuilderClient(req.BuilderId)
	if err != nil {
		return nil, fmt.Errorf("failed to get builder client: %v", err)
	}
	maxSizeOption := grpc.MaxCallRecvMsgSize(32 * 10e6)
	return client.BuildAgent(ctx, req, maxSizeOption)
}

func (s *MonarchServer) EndBuild(_ context.Context, req *builderpb.BuildRequest) (*clientpb.Empty, error) {
	return &clientpb.Empty{}, s.killBuilderClient(req.BuilderId)
}

func (s *MonarchServer) Install(req *clientpb.InstallRequest, stream rpcpb.Monarch_InstallServer) error {
	switch req.Source {
	case clientpb.InstallRequest_Local:
		builder, err := install.Setup(req.Path, stream)
		if err != nil {
			return fmt.Errorf("failed to setup local repository: %v", err)
		}
		if err = db.Create(builder); err != nil {
			return fmt.Errorf("failed to save new builder: %v", err)
		}
	case clientpb.InstallRequest_Git:
		if err := install.NewRepo(req.Path, req.Branch, req.UseCreds, stream); err != nil {
			return fmt.Errorf("failed to install %s: %v", req.Path, err)
		}
		clonePath := filepath.Join(config.MainConfig.InstallDir, strings.TrimSuffix(filepath.Base(req.Path),
			filepath.Ext(filepath.Base(req.Path))))
		if err := os.RemoveAll(clonePath); err != nil {
			return fmt.Errorf("failed to remove %s: %v. must be manually removed", clonePath, err)
		}
	}
	_ = stream.Send(&rpcpb.Notification{
		LogLevel: rpcpb.LogLevel_LevelInfo,
		Msg:      fmt.Sprintf("successfully installed %s", req.Path),
	})
	return nil
}

func (s *MonarchServer) Uninstall(req *clientpb.UninstallRequest, stream rpcpb.Monarch_UninstallServer) error {
	builders, err := s.Builders(context.Background(), req.Builders)
	if err != nil {
		return err
	}
	for _, b := range builders.Builders {
		_ = stream.Send(&rpcpb.Notification{
			LogLevel: rpcpb.LogLevel_LevelInfo,
			Msg:      fmt.Sprintf("deleting %s...", b.Name),
		})
		if req.RemoveSource {
			if err := os.RemoveAll(b.InstalledAt); err != nil {
				return fmt.Errorf("failed to remove install folder: %v", err)
			}
		}
		builder := &db.Builder{}
		if err = db.FindOneConditional("builder_id = ?", b.BuilderId, &builder); err != nil {
			return err
		}
		if err := utils.Cleanup(builder); err != nil {
			return fmt.Errorf("%v", err)
		}
		_ = stream.Send(&rpcpb.Notification{
			LogLevel: rpcpb.LogLevel_LevelSuccess,
			Msg:      fmt.Sprintf("%s v%s deleted", b.Name, b.Version),
		})
	}
	return nil
}

func (s *MonarchServer) HttpOpen(context.Context, *clientpb.Empty) (*rpcpb.Notification, error) {
	if http.MainHandler.IsActive() {
		return &rpcpb.Notification{LogLevel: rpcpb.LogLevel_LevelWarn, Msg: "http listener is already active"}, nil
	}
	go http.MainHandler.Serve()
	return &rpcpb.Notification{
		LogLevel: rpcpb.LogLevel_LevelInfo,
		Msg: fmt.Sprintf("started http listener on %s:%d",
			config.MainConfig.Interface, config.MainConfig.HttpPort),
	}, nil
}

func (s *MonarchServer) HttpsOpen(context.Context, *clientpb.Empty) (*rpcpb.Notification, error) {
	if http.MainHandler.IsActiveTLS() {
		return &rpcpb.Notification{LogLevel: rpcpb.LogLevel_LevelWarn, Msg: "https listener is already active"}, nil
	}
	go http.MainHandler.ServeTLS()
	return &rpcpb.Notification{
		LogLevel: rpcpb.LogLevel_LevelInfo,
		Msg: fmt.Sprintf("started https listener on %s:%d",
			config.MainConfig.Interface, config.MainConfig.HttpsPort),
	}, nil
}

func (s *MonarchServer) HttpClose(context.Context, *clientpb.Empty) (*clientpb.Empty, error) {
	return &clientpb.Empty{}, http.MainHandler.Stop()
}

func (s *MonarchServer) HttpsClose(context.Context, *clientpb.Empty) (*clientpb.Empty, error) {
	return &clientpb.Empty{}, http.MainHandler.StopTLS()
}

func (s *MonarchServer) Sessions(_ context.Context, req *clientpb.SessionsRequest) (*clientpb.Sessions, error) {
	var sessionsInt []int
	pbSessions := &clientpb.Sessions{}
	for _, i := range req.IDs {
		sessionsInt = append(sessionsInt, int(i))
	}
	sessions := http.MainHandler.Sessions(sessionsInt)
	for _, ss := range sessions {
		info := ss.Info
		pbSessions.Sessions = append(pbSessions.Sessions, &clientpb.Session{
			Id:         int32(ss.ID),
			AgentId:    ss.Agent.AgentID,
			AgentName:  ss.Agent.Name,
			AgentOwner: ss.Agent.CreatedBy,
			QueueSize:  int32(ss.RequestQueue.Size()),
			LastActive: ss.LastActive.Format(time.RFC850),
			Status:     ss.Status,
			BuilderId:  ss.Agent.Builder,
			Info: &clientpb.Registration{
				AgentId:   info.AgentID,
				Os:        info.OS,
				Arch:      info.Arch,
				Username:  info.Username,
				Hostname:  info.Hostname,
				UID:       info.UID,
				GID:       info.GID,
				PID:       info.PID,
				HomeDir:   info.HomeDir,
				IPAddress: info.IPAddress,
			},
		})
	}
	return pbSessions, nil
}

func (s *MonarchServer) LockSession(_ context.Context, r *clientpb.LockSessionRequest) (*clientpb.Empty, error) {
	session := http.MainHandler.SessionByID(int(r.SessionId))
	if session == nil {
		return &clientpb.Empty{}, errors.New("session not found")
	}
	if session.UsedBy != "" {
		return &clientpb.Empty{}, fmt.Errorf("session is in use by %s", session.UsedBy)
	}
	session.UsedBy = r.PlayerName
	return &clientpb.Empty{}, nil
}

func (s *MonarchServer) FreeSession(_ context.Context, r *clientpb.FreeSessionRequest) (*clientpb.Empty, error) {
	session := http.MainHandler.SessionByID(int(r.SessionId))
	if session == nil {
		return &clientpb.Empty{}, errors.New("session not found")
	}
	if session.UsedBy != r.PlayerName {
		return &clientpb.Empty{}, fmt.Errorf("unauthorized free (%s != %s)", session.UsedBy, r.PlayerName)
	}
	session.UsedBy = ""
	return &clientpb.Empty{}, nil
}

func (s *MonarchServer) Commands(_ context.Context, req *builderpb.DescriptionsRequest) (*builderpb.DescriptionsReply, error) {
	client, err := s.newBuilderClient(req.BuilderId)
	if err != nil {
		return nil, err
	}
	return client.GetCommands(context.Background(), req)
}

func (s *MonarchServer) Send(_ context.Context, req *clientpb.HTTPRequest) (*clientpb.HTTPResponse, error) {
	httpReq := &transport.GenericHTTPRequest{
		AgentID:   req.AgentId,
		RequestID: req.RequestId,
		Opcode:    req.Opcode,
		Args:      req.Args,
	}
	if err := http.MainHandler.QueueRequest(int(req.SessionId), httpReq); err != nil {
		return nil, err
	}
	httpResp := http.MainHandler.AwaitResponse(int(req.SessionId))
	resp := &clientpb.HTTPResponse{
		AgentId:   httpResp.AgentID,
		RequestId: httpResp.RequestID,
	}
	for _, d := range httpResp.Responses {
		resp.Responses = append(resp.Responses, &clientpb.ResponseDetail{
			Name:   d.Name,
			Status: d.Status,
			Dest:   clientpb.ResponseDetail_Dest(d.Dest),
			Data:   d.Data,
		})
	}
	return resp, nil
}

func (s *MonarchServer) StageView(context.Context, *clientpb.Empty) (*clientpb.Stage, error) {
	stage := &clientpb.Stage{
		Endpoint: config.MainConfig.StageEndpoint,
		Stage:    make(map[string]*clientpb.StageItem),
	}
	for k, v := range *http.Stage.View() {
		stage.Stage[k] = &clientpb.StageItem{
			Path:  v.Path,
			Agent: v.Agent,
		}
	}
	return stage, nil
}

func (s *MonarchServer) StageAdd(_ context.Context, r *clientpb.StageAddRequest) (*rpcpb.Notification, error) {
	agent := &db.Agent{}
	if err := db.FindOneConditional("agent_id = ?", r.Agent, &agent); err != nil {
		if err = db.FindOneConditional("name = ?", r.Agent, &agent); err != nil {
			return nil, fmt.Errorf("failed to retrieve the specified agent: %v", err)
		}
	}
	if len(r.Alias) == 0 {
		r.Alias = filepath.Base(agent.File)
	}
	r.Alias = filepath.Base(r.Alias)
	http.Stage.Add(r.Alias, agent.Name, agent.File)
	return &rpcpb.Notification{
		LogLevel: rpcpb.LogLevel_LevelInfo,
		Msg: fmt.Sprintf(
			"staged %s on %s",
			agent.File,
			strings.ReplaceAll(config.MainConfig.StageEndpoint, "{file}", r.Alias)),
	}, nil
}

func (s *MonarchServer) Unstage(_ context.Context, r *clientpb.UnstageRequest) (*clientpb.Empty, error) {
	http.Stage.Rm(r.Alias)
	return &clientpb.Empty{}, nil
}

func (s *MonarchServer) Notify(req *clientpb.NotifyRequest, stream rpcpb.Monarch_NotifyServer) error {
	playerID := req.PlayerId
	var authed, kicked bool
	p := &db.Player{}
	if len(playerID) == 0 {
		return errors.New("player ID cannot be blank")
	}
	if err := db.FindOneConditional("uuid = ?", playerID, &p); err != nil {
		_ = stream.Send(&rpcpb.Notification{
			LogLevel: rpcpb.LogLevel_LevelError,
			Msg:      "you are not registered to this server",
		})
		return nil
	}
	authed = true
	// notify all that you have joined the game (this is done after subbing for notifications, by calling this func)
	notifyAll(&rpcpb.Notification{
		LogLevel: rpcpb.LogLevel_LevelInfo,
		Msg:      fmt.Sprintf("%s has joined the operation", req.PlayerName),
	})
	defer func() {
		delete(types.NotifQueues, playerID)
		if !kicked && authed {
			notifyAll(&rpcpb.Notification{
				LogLevel: rpcpb.LogLevel_LevelInfo,
				Msg:      fmt.Sprintf("%s has left the operation", req.PlayerName),
			})
		}
	}()
	// Implement a notification queue
	queue := &types.NotificationQueue{Channel: make(chan *rpcpb.Notification, 10)}
	types.NotifQueues[playerID] = queue
	for {
		select {
		case <-stream.Context().Done():
			return nil
		case notification := <-queue.Channel:
			_ = stream.Send(notification)
			if notification.Msg == types.NotificationKickPlayer {
				// name and shame!
				notifyAll(&rpcpb.Notification{
					LogLevel: rpcpb.LogLevel_LevelInfo,
					Msg:      fmt.Sprintf("%s has been kicked from the operation", req.PlayerName),
				}, p.Username, config.ClientConfig.UUID)
				kicked = true
				break
			}
		}
	}
}

func notifyAll(n *rpcpb.Notification, excludes ...string) {
	for k, q := range types.NotifQueues {
		if !slices.Contains(excludes, k) {
			_ = q.Enqueue(n)
		}
	}
}

func Stop() {
	grpcServer.Stop()
}

func Start() error {
	lis, err := net.Listen("tcp", fmt.Sprintf("0.0.0.0:%d", config.MainConfig.MultiplayerPort))
	if err != nil {
		return err
	}
	creds := credentials.NewTLS(serverTlsConfig())
	interceptor := NewAuthInterceptor()

	opts := []grpc.ServerOption{
		grpc.Creds(creds),
		grpc.UnaryInterceptor(interceptor.Unary()),
		grpc.MaxRecvMsgSize(math.MaxInt32),
	}
	grpcServer = grpc.NewServer(opts...)
	srv, err := New()
	if err != nil {
		return err
	}
	rpcpb.RegisterMonarchServer(grpcServer, srv)
	// deliberately blocking
	return grpcServer.Serve(lis)
}

func serverTlsConfig() *tls.Config {
	caCert, _, err := crypto.CertificateAuthority()
	if err != nil {
		logrus.Fatalf("could not retrieve CA certificate: %v", err)
	}
	caCertPool := x509.NewCertPool()
	caCertPool.AddCert(caCert)

	certBlock, keyBlock, err := config.ServerCertificates()
	if err != nil {
		logrus.Fatalf("couldn't fetch server server certificate: %v", err)
	}
	cert, err := tls.X509KeyPair(certBlock, keyBlock)
	if err != nil {
		logrus.Fatalf("couldn't load key pair: %v", err)
	}
	return &tls.Config{
		RootCAs: caCertPool,
		// below checks whether client certs were signed by a RootCA, and nothing else
		ClientAuth:   tls.RequireAndVerifyClientCert,
		ClientCAs:    caCertPool,
		Certificates: []tls.Certificate{cert},
		MinVersion:   tls.VersionTLS13,
	}
}
