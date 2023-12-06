package console

import (
	"fmt"
	"github.com/pygrum/monarch/pkg/config"
	"github.com/pygrum/monarch/pkg/consts"
	"github.com/pygrum/monarch/pkg/db"
	"github.com/pygrum/monarch/pkg/log"
	"github.com/pygrum/monarch/pkg/protobuf/rpcpb"
	"github.com/pygrum/monarch/pkg/teamserver"
	"github.com/reeflective/console"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"google.golang.org/grpc"
	"net"
)

type server struct {
	App *console.Console
}

var (
	Rpc           rpcpb.MonarchClient
	monarchServer *server
)

func init() {
	monarchServer = &server{
		App: console.New("monarch"),
	}
	db.Initialize()
	log.Initialize(monarchServer.App.TransientPrintf)
}

// NamedMenu switches the console to a new menu with the provided name.
func NamedMenu(name string, commands func() *cobra.Command) {
	namedMenu := monarchServer.App.NewMenu(name)
	namedMenu.SetCommands(commands)
	monarchServer.App.SwitchMenu(name)
}

// Run entrypoint for the entire application
func Run(rootCmd func() *cobra.Command) error {
	lis, err := net.Listen("tcp", fmt.Sprintf("localhost:%d", config.MainConfig.MultiplayerPort))
	if err != nil {
		return err
	}
	grpcServer, err := newMonarchServer()
	if err != nil {
		return err
	}
	go func() {
		if err = grpcServer.Serve(lis); err != nil {
			logrus.Fatal(err)
		}
	}()

	// new internal grpc client
	conn, err := grpc.Dial(fmt.Sprintf("localhost:%d", config.MainConfig.MultiplayerPort))
	if err != nil {
		return err
	}
	Rpc = rpcpb.NewMonarchClient(conn)
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

func newMonarchServer() (*grpc.Server, error) {
	teamserver.NotifQueue = &teamserver.NotificationQueue{Channel: make(chan *rpcpb.PlayerNotification, 10)}
	var opts []grpc.ServerOption
	// TODO: fetch key pair and create credentials with credentials.NewTLS
	grpcServer := grpc.NewServer(opts...)
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
