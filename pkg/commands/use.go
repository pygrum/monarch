package commands

import (
	"encoding/binary"
	"github.com/google/uuid"
	"github.com/pygrum/monarch/pkg/console"
	"github.com/pygrum/monarch/pkg/handler/http"
	"github.com/pygrum/monarch/pkg/protobuf/builderpb"
	"github.com/pygrum/monarch/pkg/protobuf/clientpb"
	"github.com/pygrum/monarch/pkg/transport"
	"github.com/spf13/cobra"
	"os"
	"path/filepath"
)

func useCmd(id int) {
	ss, err := console.Rpc.Sessions(ctx, &clientpb.SessionsRequest{IDs: []int32{int32(id)}})
	if err != nil {
		cLogger.Error("%v", err)
		return
	}
	if len(ss.Sessions) == 0 {
		cLogger.Error("no active sessions")
		return
	}
	sessionInfo := ss.Sessions[0]
	descriptions, err := console.Rpc.Commands(ctx,
		&builderpb.DescriptionsRequest{BuilderId: sessionInfo.AgentId + sessionInfo.BuilderId})
	if err != nil {
		cLogger.Error("failed to acquire command descriptions (rpc): %v", err)
		return
	}
	console.NamedMenu("\033[35m"+sessionInfo.AgentName+"\033[0m", func() *cobra.Command {
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
				if description.MinArgs > 0 {
					args = cobra.ExactArgs(int(description.MinArgs))
				} else {
					args = cobra.NoArgs
				}
			}
			op := description.Opcode // must copy out to use in cobra Command otherwise it will be the last
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
							info, err := os.Stat(filename)
							if err != nil {
								cLogger.Error("%v", err)
								return
							}
							bytes, err := os.ReadFile(filename)
							if err != nil {
								cLogger.Error("failed to read file %s", filename)
								return
							}
							sizeBytes := make([]byte, 8)
							binary.BigEndian.PutUint64(sizeBytes, uint64(info.Size())) // enforce network byte order
							// packet looks like: [filesize][......file-data.......][filename]
							data = append(sizeBytes, append(bytes, []byte(filepath.Base(filename))...)...)
						}
						byteArgs[i] = data
					}
					req := &clientpb.HTTPRequest{
						SessionId: sessionInfo.Id,
						AgentId:   sessionInfo.AgentId,
						RequestId: uuid.New().String(),
						Opcode:    op,
						Args:      byteArgs,
					}
					go func() {
						resp, err := console.Rpc.Send(ctx, req)
						if err != nil {
							http.TranLogger.Error("%v", err)
						}
						r := &transport.GenericHTTPResponse{
							AgentID:   resp.AgentId,
							RequestID: resp.RequestId,
						}
						for _, response := range resp.Responses {
							r.Responses = append(r.Responses, transport.ResponseDetail{
								Status: response.Status,
								Dest:   int32(response.Dest),
								Name:   response.Name,
								Data:   response.Data,
							})
						}
						http.HandleResponse(sessionInfo, r)
					}()
				},
			}
			rootCmd.AddCommand(cmd)
		}
		rootCmd.AddCommand(exit("", "use"))
		rootCmd.AddCommand(info(sessionInfo.Info))
		rootCmd.CompletionOptions.HiddenDefaultCmd = true
		return rootCmd
	})
}
