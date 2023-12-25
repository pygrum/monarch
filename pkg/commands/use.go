package commands

import (
	"encoding/binary"
	"github.com/desertbit/grumble"
	"github.com/google/uuid"
	"github.com/pygrum/monarch/pkg/config"
	"github.com/pygrum/monarch/pkg/console"
	"github.com/pygrum/monarch/pkg/consts"
	"github.com/pygrum/monarch/pkg/handler/http"
	"github.com/pygrum/monarch/pkg/protobuf/builderpb"
	"github.com/pygrum/monarch/pkg/protobuf/clientpb"
	"github.com/pygrum/monarch/pkg/transport"
	"golang.org/x/text/cases"
	"golang.org/x/text/language"
	"google.golang.org/grpc"
	"os"
	"path/filepath"
)

var cmds []string

func removeSessionCommands() {
	for _, cmd := range cmds {
		console.App.Commands().Remove(cmd)
	}
	console.App.Commands().Remove("info")
	console.App.Commands().Remove("end-session")
	cmds = nil
}

func endSessionCmd(id int32) *grumble.Command {
	cmd := &grumble.Command{
		Name:      "end-session",
		Help:      "end an interactive agent session",
		HelpGroup: consts.GeneralHelpGroup,
		Run: func(c *grumble.Context) error {
			if _, err := console.Rpc.FreeSession(ctx, &clientpb.FreeSessionRequest{
				SessionId: id, PlayerName: config.ClientConfig.Name,
			}); err != nil {
				cLogger.Error("couldn't end session: %v", err)
			}
			removeSessionCommands()
			console.App.SetDefaultPrompt()
			return nil
		},
	}
	return cmd
}

func useHelpGroup(s string) string {
	return cases.Title(language.English, cases.Compact).String(s)
}

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
	if _, err = console.Rpc.LockSession(ctx, &clientpb.LockSessionRequest{
		SessionId: sessionInfo.Id, PlayerName: config.ClientConfig.Name}); err != nil {
		cLogger.Error("couldn't acquire session: %v", err)
		return
	}
	prompt := "(" + sessionInfo.AgentName + ") > "

	console.App.AddCommand(endSessionCmd(sessionInfo.Id))
	console.App.AddCommand(info(sessionInfo.Info))

	console.App.SetPrompt("monarch " + prompt)

	for _, description := range descriptions.Descriptions {
		if c := console.App.Commands().Get(description.Name); c != nil {
			cLogger.Error("cannot load commands for %s: duplicate command %s",
				sessionInfo.AgentName,
				description.Name)
			return
		}
		cmds = append(cmds, description.Name)
		op := description.Opcode // must copy out to use in cobra Command otherwise it will be the last
		cmd := &grumble.Command{
			Name:      description.Name,
			Usage:     description.Usage,
			Help:      description.DescriptionShort,
			LongHelp:  description.DescriptionLong,
			HelpGroup: useHelpGroup(sessionInfo.AgentName),
			Args: func(a *grumble.Args) {
				minA := int(description.MinArgs)
				maxA := int(description.MaxArgs)
				if maxA < 0 {
					maxA = 1<<30 - 1
				}
				if minA < 0 {
					// reset if invalid
					minA = 0
				}
				if maxA > 0 && minA < maxA {
					a.StringList("args", "command arguments", grumble.Min(minA),
						grumble.Max(maxA))
				} else {
					a.StringList("args", "command arguments")
				}
			},
			Run: func(c *grumble.Context) error {
				args := c.Args.StringList("args")
				byteArgs := make([][]byte, len(args))
				for i, arg := range args {
					data := []byte(arg)
					// Refers to a file if prefixed with @
					if arg[0] == '@' {
						filename := arg[1:]
						info, err := os.Stat(filename)
						if err != nil {
							cLogger.Error("%v", err)
							return nil
						}
						bytes, err := os.ReadFile(filename)
						if err != nil {
							cLogger.Error("failed to read file %s", filename)
							return nil
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
					maxSizeOption := grpc.MaxCallRecvMsgSize(32 * 10e6)
					resp, err := console.Rpc.Send(ctx, req, maxSizeOption)
					if err != nil {
						http.TranLogger.Error("%v", err)
						return
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
				return nil
			},
		}
		console.App.AddCommand(cmd)
	}
}
