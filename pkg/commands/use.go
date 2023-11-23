package commands

import (
	"context"
	"github.com/google/uuid"
	"github.com/pygrum/monarch/pkg/console"
	"github.com/pygrum/monarch/pkg/db"
	"github.com/pygrum/monarch/pkg/docker"
	"github.com/pygrum/monarch/pkg/handlers/xhttp"
	"github.com/pygrum/monarch/pkg/rpcpb"
	"github.com/pygrum/monarch/pkg/transport"
	"github.com/spf13/cobra"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"os"
	"strconv"
)

func useCmd(id int) {
	ctx := context.Background()
	sessionInfo := xhttp.Handler.SessionByID(id)
	if sessionInfo == nil {
		cLogger.Error("session '%d' not found", id)
		return
	}
	builder := &db.Builder{}
	if err := db.FindOneConditional("builder_id = ?", sessionInfo.Agent.Builder, builder); err != nil ||
		len(builder.BuilderID) == 0 {
		cLogger.Error("failed to acquire commands: %v", err)
		return
	}
	rpc, err := docker.RPCAddress(docker.Cli, ctx, builder.BuilderID)
	if err != nil {
		cLogger.Error("failed to acquire commands: %v", err)
		return
	}
	conn, err := grpc.Dial(rpc, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		cLogger.Error("RPC connection failed: %v", err)
		return
	}
	client := rpcpb.NewBuilderClient(conn)
	descriptions, err := client.GetCommands(ctx, &rpcpb.DescriptionsRequest{})
	if err != nil {
		cLogger.Error("failed to acquire command descriptions (rpc): %v", err)
		return
	}
	rootCmd := &cobra.Command{}
	for _, description := range descriptions.Descriptions {
		args := cobra.NoArgs
		if description.NumArgs > 0 {
			args = cobra.ExactArgs(int(description.NumArgs))
		} else if description.NumArgs < 0 {
			args = cobra.ArbitraryArgs
		}
		cmd := &cobra.Command{
			Use:   description.Usage,
			Short: description.DescriptionShort,
			Long:  description.DescriptionLong,
			Args:  args,
			Run: func(cmd *cobra.Command, args []string) {
				byteArgs := make([][]byte, len(args))
				for i, arg := range args {
					data := []byte(arg)
					// Refers to a file if prefixed and suffixed with @
					if arg[0] == '@' && arg[len(arg)-1] == '@' {
						filename := arg[1 : len(arg)-1]
						bytes, err := os.ReadFile(filename)
						if err != nil {
							cLogger.Error("failed to read file %s", filename)
							return
						}
						data = bytes
					}
					byteArgs[i] = data
				}
				req := &transport.GenericHTTPRequest{
					AgentID:   sessionInfo.Agent.AgentID,
					RequestID: uuid.New().String(),
					Opcode:    description.Opcode,
					Args:      byteArgs,
				}
				if err = xhttp.Handler.QueueRequest(sessionInfo.ID, req); err != nil {
					cLogger.Error("failed to queue request: %v", err)
				}
			},
		}
		rootCmd.AddCommand(cmd)
	}
	rootCmd.AddCommand(exit(""))
	console.NamedMenu(strconv.Itoa(sessionInfo.ID), func() *cobra.Command {
		return rootCmd
	})
}
