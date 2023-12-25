package completion

import (
	"context"
	"github.com/pygrum/monarch/pkg/console"
	"github.com/pygrum/monarch/pkg/protobuf/clientpb"
	"os"
	"path/filepath"
	"strconv"
	"strings"
)

func Builders(prefix string, mctx context.Context) []string {
	var results []string
	bs, err := console.Rpc.Builders(mctx, &clientpb.BuilderRequest{BuilderId: make([]string, 0)})
	if err == nil {
		for _, b := range bs.Builders {
			if strings.HasPrefix(b.Name, prefix) {
				results = append(results, b.Name)
			}
		}
	}
	return results
}

func Agents(prefix string, mctx context.Context) []string {
	var results []string
	as, err := console.Rpc.Agents(mctx, &clientpb.AgentRequest{AgentId: make([]string, 0)})
	if err == nil {
		for _, a := range as.Agents {
			if strings.HasPrefix(a.Name, prefix) {
				results = append(results, a.Name)
			}
		}
	}
	return results
}

func Profiles(mctx context.Context, prefix, builderId string) []string {
	var results []string
	ps, err := console.Rpc.Profiles(mctx, &clientpb.ProfileRequest{Name: make([]string, 0), BuilderId: builderId})
	if err == nil {
		for _, p := range ps.Profiles {
			if strings.HasPrefix(p.Name, prefix) {
				results = append(results, p.Name)
			}
		}
	}
	return results
}

func Options(prefix string, allOptions []string) []string {
	var results []string
	for _, o := range allOptions {
		if strings.HasPrefix(o, prefix) {
			results = append(results, o)
		}
	}
	return results
}

func Players(prefix string, mctx context.Context) []string {
	var results []string
	players, err := console.Rpc.Players(mctx, &clientpb.PlayerRequest{Names: make([]string, 0)})
	if err == nil {
		for _, p := range players.Players {
			if strings.HasPrefix(p.Username, prefix) {
				results = append(results, p.Username)
			}
		}
	}
	return results
}

func Sessions(prefix string, mctx context.Context) []string {
	var results []string
	sessions, err := console.Rpc.Sessions(mctx, &clientpb.SessionsRequest{IDs: make([]int32, 0)})
	if err == nil {
		for _, s := range sessions.Sessions {
			strId := strconv.Itoa(int(s.Id))
			if strings.HasPrefix(strId, prefix) {
				results = append(results, strId)
			}
		}
	}
	return results
}

func UnStage(prefix string, mctx context.Context) []string {
	var results []string
	stage, err := console.Rpc.StageView(mctx, &clientpb.Empty{})
	if err == nil {
		for s := range stage.Stage {
			if strings.HasPrefix(s, prefix) {
				results = append(results, s)
			}
		}
	}
	return results
}

// LocalPathCompleter :P https://github.com/BishopFox/sliver/blob/v1.5.41/client/command/completers/completers.go
func LocalPathCompleter(prefix string) []string {
	var parent string
	var partial string
	fi, err := os.Stat(prefix)
	if os.IsNotExist(err) {
		parent = filepath.Dir(prefix)
		partial = filepath.Base(prefix)
	} else {
		if fi.IsDir() {
			parent = prefix
			partial = ""
		} else {
			parent = filepath.Dir(prefix)
			partial = filepath.Base(prefix)
		}
	}

	results := []string{}
	ls, err := os.ReadDir(parent)
	if err != nil {
		return results
	}
	for _, dirent := range ls {
		fi, err = dirent.Info()
		if err == nil {
			if 0 < len(partial) {
				if strings.HasPrefix(fi.Name(), partial) {
					results = append(results, filepath.Join(parent, fi.Name()))
				}
			} else {
				results = append(results, filepath.Join(parent, fi.Name()))
			}
		}
	}
	return results
}
