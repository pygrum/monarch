package console

import (
	"context"
	"fmt"
	"net"

	"github.com/pygrum/monarch/pkg/config"
	"github.com/pygrum/monarch/pkg/consts"
	"github.com/pygrum/monarch/pkg/crypto"
	"github.com/pygrum/monarch/pkg/db"
	"github.com/pygrum/monarch/pkg/handler/http"
	"github.com/pygrum/monarch/pkg/log"
	"github.com/pygrum/monarch/pkg/protobuf/clientpb"
	"github.com/pygrum/monarch/pkg/protobuf/rpcpb"
	"github.com/pygrum/monarch/pkg/teamserver"
	"github.com/reeflective/console"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/credentials/insecure"
)

type server struct {
	App *console.Console
}

var (
	Rpc           rpcpb.MonarchClient
	monarchServer *server
)

// NamedMenu switches the console to a new menu with the provided name.
func NamedMenu(name string, commands func() *cobra.Command) {
	namedMenu := monarchServer.App.NewMenu(name)
	namedMenu.SetCommands(commands)
	monarchServer.App.SwitchMenu(name)
}

// Run entrypoint for the entire application
func Run(rootCmd func() *cobra.Command, isServer bool) error {
	var err error
	var clientConn *grpc.ClientConn
	monarchServer = &server{
		App: console.New("monarch"),
	}
	if isServer {
		clientConn, err = initMonarchServer()
		if err != nil {
			return err
		}
	} else {
		clientConn, err = initMonarchClient()
		if err != nil {
			return err
		}
	}
	log.Initialize(monarchServer.App.TransientPrintf)
	Rpc = rpcpb.NewMonarchClient(clientConn)
	go getNotifications()
	srvMenu := monarchServer.App.ActiveMenu()
	srvMenu.SetCommands(rootCmd)
	monarchServer.App.SetPrintLogo(func(_ *console.Console) {
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
	return monarchServer.App.Start()
}

func getNotifications() {
	var playerID, playerName string
	playerID = config.ClientConfig.UUID
	playerName = config.ClientConfig.Name

	tl, _ := log.NewLogger(log.TransientLogger, "")
	stream, err := Rpc.Notify(context.Background(), &clientpb.NotifyRequest{
		PlayerId:   playerID,
		PlayerName: playerName,
	})
	if err != nil {
		tl.Fatal("can't receive notifications (%v)", err)
	}
	for {
		notif, err := stream.Recv()
		if err != nil {
			tl.Error("notification error: %v", err)
			return
		}
		log.NumericalLevel(tl, uint16(notif.LogLevel), notif.Msg)
	}
}

func initMonarchServer() (*grpc.ClientConn, error) {
	config.Initialize()
	uid := db.Initialize()
	http.Initialize()
	config.ClientConfig.UUID = uid
	config.ClientConfig.Name = "console"
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
	http.NotifQueues = make(map[string]http.Queue)
	// TODO: fetch key pair and create credentials with credentials.NewTLS
	grpcServer := grpc.NewServer()
	srv, err := teamserver.New()
	if err != nil {
		return nil, fmt.Errorf("failed to create new teamserver: %v", err)
	}
	rpcpb.RegisterMonarchServer(grpcServer, srv)
	return grpcServer, nil
}

// MainMenu switches back to the main menu
func MainMenu() {
	monarchServer.App.SwitchMenu("")
}
