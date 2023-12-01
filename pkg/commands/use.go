package commands

import (
	"context"
	"encoding/binary"
	"github.com/google/uuid"
	"github.com/pygrum/monarch/pkg/console"
	"github.com/pygrum/monarch/pkg/db"
	"github.com/pygrum/monarch/pkg/docker"
	"github.com/pygrum/monarch/pkg/handler/http"
	"github.com/pygrum/monarch/pkg/rpcpb"
	"github.com/pygrum/monarch/pkg/transport"
	"github.com/spf13/cobra"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"os"
	"path/filepath"
)

func useCmd(id int) {
	ctx := context.Background()
	sessionInfo := http.MainHandler.SessionByID(id)
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
	console.NamedMenu("\033[35m"+sessionInfo.Agent.Name+"\033[0m", func() *cobra.Command {
		// rootCmd must be defined in here to prevent help flag bug
		rootCmd := &cobra.Command{}
		for _, description := range descriptions.Descriptions {
			if description.MinArgs < 0 {
				// reset if invalid
				description.MinArgs = 0
			}
			args := cobra.MinimumNArgs(int(description.MinArgs))
			if description.MaxArgs >= description.MinArgs {
				args = cobra.RangeArgs(int(description.MinArgs), int(description.MaxArgs))
			}
			if description.MinArgs == description.MaxArgs {
				args = cobra.ExactArgs(int(description.MinArgs))
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
						// Refers to a file if prefixed with @
						if arg[0] == '@' {
							filename := arg[1:]
							bytes, err := os.ReadFile(filename)
							if err != nil {
								cLogger.Error("failed to read file %s", filename)
								return
							}
							sizeBytes := make([]byte, 8)
							binary.BigEndian.PutUint64(sizeBytes, uint64(len(bytes))) // enforce network byte order
							// packet looks like: [filesize][......file-data.......][filename]
							data = append(sizeBytes, append(bytes, []byte(filepath.Base(filename))...)...)
						}
						byteArgs[i] = data
					}
					req := &transport.GenericHTTPRequest{
						AgentID:   sessionInfo.Agent.AgentID,
						RequestID: uuid.New().String(),
						Opcode:    description.Opcode,
						Args:      byteArgs,
					}
					if err = http.MainHandler.QueueRequest(sessionInfo.ID, req); err != nil {
						cLogger.Error("%v", err)
						console.MainMenu()
						return
					}
					l.Info("queued request %s for %s", http.ShortID(req.RequestID), sessionInfo.Agent.Name)
					//resp := http.MainHandler.AwaitResponse(sessionInfo.ID)
					//http.HandleResponse(sessionInfo, resp)
				},
			}
			rootCmd.AddCommand(cmd)
		}
		rootCmd.AddCommand(exit(""))
		rootCmd.AddCommand(info(sessionInfo.Info))
		rootCmd.CompletionOptions.HiddenDefaultCmd = true
		return rootCmd
	})
}

// TODO: SPECIFY A FLAG TO NOT WAIT FOR A RESPONSE (AwaitResponse) and handle resp in separate goroutine instead
// TODO: multiplayer?
