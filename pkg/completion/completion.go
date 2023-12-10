package completion

import (
	"context"
	"github.com/pygrum/monarch/pkg/console"
	"github.com/pygrum/monarch/pkg/consts"
	"github.com/pygrum/monarch/pkg/db"
	"github.com/pygrum/monarch/pkg/protobuf/clientpb"
	"github.com/rsteube/carapace"
	"strconv"
)

func Builders(mctx context.Context) carapace.Action {
	c := func(ctx carapace.Context) carapace.Action {
		var results []string
		bs, err := console.Rpc.Builders(mctx, &clientpb.BuilderRequest{BuilderId: make([]string, 0)})
		if err == nil {
			for _, b := range bs.Builders {
				results = append(results, b.Name)
			}
		}
		return carapace.ActionValues(results...).Tag("builders")
	}
	return carapace.ActionCallback(c)
}

func Agents(mctx context.Context) carapace.Action {
	c := func(ctx carapace.Context) carapace.Action {
		var results []string
		as, err := console.Rpc.Agents(mctx, &clientpb.AgentRequest{AgentId: make([]string, 0)})
		if err == nil {
			for _, a := range as.Agents {
				results = append(results, a.Name)
			}
		}
		return carapace.ActionValues(results...).Tag("agents")
	}
	return carapace.ActionCallback(c)
}

func Profiles(mctx context.Context, builderId string) carapace.Action {
	c := func(ctx carapace.Context) carapace.Action {
		var results []string
		ps, err := console.Rpc.Profiles(mctx, &clientpb.ProfileRequest{Name: make([]string, 0), BuilderId: builderId})
		if err == nil {
			for _, p := range ps.Profiles {
				results = append(results, p.Name)
			}
		}
		return carapace.ActionValues(results...).Tag("profiles")
	}
	return carapace.ActionCallback(c)
}

func Options(options []string) carapace.Action {
	c := func(ctx carapace.Context) carapace.Action {
		return carapace.ActionValues(options...).Tag("options")
	}
	return carapace.ActionCallback(c)
}

func Players() carapace.Action {
	c := func(ctx carapace.Context) carapace.Action {
		var results []string
		var players []db.Player
		if err := db.Find(&players); err == nil {
			for _, p := range players {
				if p.Username != consts.UserConsole {
					results = append(results, p.Username)
				}
			}
		}
		return carapace.ActionValues(results...).Tag("players")
	}
	return carapace.ActionCallback(c)
}

func Sessions(mctx context.Context) carapace.Action {
	c := func(ctx carapace.Context) carapace.Action {
		var results []string
		sessions, err := console.Rpc.Sessions(mctx, &clientpb.SessionsRequest{IDs: make([]int32, 0)})
		if err == nil {
			for _, s := range sessions.Sessions {
				results = append(results, strconv.Itoa(int(s.Id)))
			}
		}
		return carapace.ActionValues(results...).Tag("sessions")
	}
	return carapace.ActionCallback(c)
}

func UnStage(mctx context.Context) carapace.Action {
	c := func(ctx carapace.Context) carapace.Action {
		var results []string
		stage, err := console.Rpc.StageView(mctx, &clientpb.Empty{})
		if err == nil {
			for s := range stage.Stage {
				results = append(results, s)
			}
		}
		return carapace.ActionValues(results...).Tag("aliases")
	}
	return carapace.ActionCallback(c)
}
