package console

import (
	"context"
	"errors"
	"fmt"
	"github.com/desertbit/grumble"
	"github.com/fatih/color"
	"github.com/pygrum/monarch/pkg/types"
	"google.golang.org/grpc/metadata"
	"io"
	"net"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/pygrum/monarch/pkg/config"
	"github.com/pygrum/monarch/pkg/consts"
	"github.com/pygrum/monarch/pkg/crypto"
	"github.com/pygrum/monarch/pkg/handler/http"
	"github.com/pygrum/monarch/pkg/log"
	"github.com/pygrum/monarch/pkg/protobuf/clientpb"
	"github.com/pygrum/monarch/pkg/protobuf/rpcpb"
	"github.com/pygrum/monarch/pkg/teamserver"
	"github.com/sirupsen/logrus"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/credentials/insecure"
)

var (
	App *grumble.App
	Rpc rpcpb.MonarchClient
	CTX = context.Background()
)

// Run entrypoint for the entire application
func Run(isServer bool, rootCmds ...func() []*grumble.Command) {
	initCTX(isServer)
	var err error
	var clientConn *grpc.ClientConn
	App = grumble.New(&grumble.Config{
		Name:                  "Monarch",
		Prompt:                consts.DefaultPrompt,
		ASCIILogoColor:        color.New(color.FgHiWhite),
		PromptColor:           color.New(color.FgHiWhite),
		HistoryFile:           filepath.Join(config.Home(), ".monarch_history"),
		HelpHeadlineUnderline: true,
		HelpSubCommands:       true,
	})

	if isServer {
		clientConn, err = initMonarchServer()
		if err != nil {
			logrus.Fatal(err)
		}
	} else {
		clientConn, err = initMonarchClient()
		testServerConnectivity()
		if err != nil {
			logrus.Fatal(err)
		}
	}
	log.Initialize(App)
	Rpc = rpcpb.NewMonarchClient(clientConn)

	go getNotifications()
	go getMessages()
	start(rootCmds)

}

func start(rootCmds []func() []*grumble.Command) {
	for _, rootCmd := range rootCmds {
		for _, cmd := range rootCmd() {
			App.AddCommand(cmd)
		}
	}
	App.SetPrintASCIILogo(func(_ *grumble.App) {
		fmt.Print("\033[H\033[2J")
		fmt.Printf(`                  o 
               o^/|\^o
            o_^|\/*\/|^_o
           o\*¬'.\|/.'¬*/o
            \\\\\\|//////
             {><><@><><}
             |"""""""""|
               MONARCH
  ADVERSARY EMULATION TOOLKIT v%s
  ==================================

		`, consts.Version)
	})
	grumble.Main(App)
}

func testServerConnectivity() {
	if _, err := net.Dial("tcp",
		net.JoinHostPort(config.ClientConfig.RHost, strconv.Itoa(config.ClientConfig.RPort))); err != nil {
		logrus.Fatalf("server seems down (%v)", err)
	}
}

func getNotifications() {
	tl, _ := log.NewLogger(log.TransientLogger, "")
	stream, err := Rpc.Notify(CTX, &clientpb.Empty{})
	if err != nil {
		tl.Fatal("can't receive notifications (%v)", err)
	}
	for {
		notif, err := stream.Recv()
		if errors.Is(err, io.EOF) {
			tl.Error("server connection closed")
			return
		}
		if err != nil {
			tl.Error("notification error: %v", err)
			return
		}
		log.NumericalLevel(tl, uint16(notif.LogLevel), notif.Msg)
		if notif.Msg == types.NotificationKickPlayer {
			_ = stream.CloseSend()
			return
		}
	}
}

func getMessages() {
	tl, _ := log.NewLogger(log.TransientLogger, "")
	stream, err := Rpc.GetMessages(CTX, &clientpb.Empty{})
	if err != nil {
		tl.Error("can't receive messages (%v)", err)
		return
	}
	for {
		message, err := stream.Recv()
		if errors.Is(err, io.EOF) {
			tl.Error("server connection closed")
			return
		}
		if err != nil {
			tl.Error("messaging error: %v", err)
			return
		}
		msgFmt := "%s [%s] says: \033[36m%s\033[0m"
		msg := fmt.Sprintf(msgFmt, message.From, message.Role, message.Msg)
		_, _ = fmt.Fprintln(App.Stdout(), strings.ReplaceAll(msg, "%", "%%"))
	}
}

func initMonarchServer() (*grpc.ClientConn, error) {
	http.Initialize()
	lis, err := net.Listen("tcp", fmt.Sprintf("localhost:9999"))
	if err != nil {
		logrus.Fatalf("couldn't listen on localhost: %v", err)
	}
	grpcServer, err := newMonarchServer()
	if err != nil {
		return nil, err
	}
	go func() {
		if err = grpcServer.Serve(lis); err != nil {
			logrus.Fatal(err)
		}
	}()

	// new internal grpc client
	conn, err := grpc.Dial(fmt.Sprintf("localhost:9999"),
		grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		return nil, err
	}
	return conn, nil
}

func initMonarchClient() (*grpc.ClientConn, error) {
	http.ClientInitialize()
	c, err := crypto.ClientTLSConfig(&config.ClientConfig)
	if err != nil {
		return nil, err
	}
	conn, err := grpc.Dial(fmt.Sprintf("%s:%d", config.ClientConfig.RHost, config.ClientConfig.RPort),
		grpc.WithTransportCredentials(credentials.NewTLS(c)))
	if err != nil {
		return conn, err
	}
	return conn, nil
}

func newMonarchServer() (*grpc.Server, error) {
	types.NotifQueues = make(map[string]types.Queue)
	types.MessageQueues = make(map[string]types.Queue)

	// TODO: fetch key pair and create credentials with credentials.NewTLS
	grpcServer := grpc.NewServer()
	srv, err := teamserver.New()
	if err != nil {
		return nil, fmt.Errorf("failed to create new teamserver: %v", err)
	}
	rpcpb.RegisterMonarchServer(grpcServer, srv)
	return grpcServer, nil
}

func initCTX(isServer bool) {
	m := make(map[string]string)
	m["uid"] = config.ClientConfig.UUID
	if isServer {
		m["username"] = consts.ServerUser
		CTX = metadata.NewOutgoingContext(CTX, metadata.New(m))
		return
	}
	challenge, err := crypto.EncryptAES(config.ClientConfig.Secret, config.ClientConfig.Challenge)
	if err != nil {
		logrus.Fatalf("couldn't encrypt challenge for auth: %v", err)
	}
	m["challenge"] = challenge
	md := metadata.New(m)
	CTX = metadata.NewOutgoingContext(CTX, md)
}
