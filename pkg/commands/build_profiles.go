package commands

import (
	"fmt"
	"github.com/desertbit/grumble"
	"github.com/pygrum/monarch/pkg/completion"
	"github.com/pygrum/monarch/pkg/console"
	"github.com/pygrum/monarch/pkg/consts"
	"github.com/pygrum/monarch/pkg/protobuf/clientpb"
	"strings"
)

func profilesCmd(names []string) {
	profiles, err := console.Rpc.Profiles(ctx, &clientpb.ProfileRequest{Name: names, BuilderId: builderConfig.builderID})
	if err != nil {
		cLogger.Error("%v", err.Error())
		return
	}
	headers := "NAME\tCREATION TIME\t"
	_, _ = fmt.Fprintln(w, headers)
	for _, p := range profiles.Profiles {
		line := fmt.Sprintf("%s\t%s\t", p.Name, p.CreatedAt)
		_, _ = fmt.Fprintln(w, line)
	}
	_ = w.Flush()
}

func profilesSaveCmd(name string) {
	if _, err := console.Rpc.SaveProfile(ctx, &clientpb.SaveProfileRequest{
		Name:      name,
		BuilderId: builderConfig.builderID,
		Options:   builderConfig.request.Options,
	}); err != nil {
		cLogger.Error("%v", err)
		return
	}
	cLogger.Success("profile saved as '%s'", name)
}

func profilesLoadCmd(name string) {
	profile, err := console.Rpc.LoadProfile(ctx, &clientpb.SaveProfileRequest{
		Name:       name,
		BuilderId:  builderConfig.builderID,
		Immutables: internals,
	})
	if err != nil {
		cLogger.Error("%v", err)
		return
	}
	for _, r := range profile.Records {
		setCmd(r.Name, r.Value)
	}
	l.Success("profile %s loaded", profile.Profile.Name)
}

func profilesRmCmd(names []string) {
	if _, err := console.Rpc.RmProfiles(ctx, &clientpb.ProfileRequest{Name: names, BuilderId: builderConfig.builderID}); err != nil {
		cLogger.Error("%v", err)
		return
	}
	cLogger.Success("successfully deleted %s", strings.Join(names, ", "))
}

func cobraProfilesCmd() *grumble.Command {
	cmd := &grumble.Command{
		Name: "profiles",
		Help: "list all created profiles",
		Args: func(a *grumble.Args) {
			a.StringList("names", "names of created profiles")
		},
		HelpGroup: consts.BuildHelpGroup,
		Run: func(c *grumble.Context) error {
			if err := buildCheck(); err != nil {
				return err
			}
			profilesCmd(c.Args.StringList("names"))
			return nil
		},
		Completer: func(prefix string, args []string) []string {
			if err := buildCheck(); err != nil {
				return nil
			}
			return completion.Profiles(ctx, prefix, builderConfig.builderID)
		},
	}

	cmd.AddCommand(&grumble.Command{
		Name: "save",
		Help: "save current build configuration options as a new profile",
		Args: func(a *grumble.Args) {
			a.String("name", "name to save the current settings under")
		},
		HelpGroup: consts.BuildHelpGroup,
		Run: func(c *grumble.Context) error {
			if err := buildCheck(); err != nil {
				return err
			}
			profilesSaveCmd(c.Args.String("name"))
			return nil
		},
	})
	cmd.AddCommand(&grumble.Command{
		Name: "load",
		Help: "load an existing profile",
		Args: func(a *grumble.Args) {
			a.String("name", "name of profile")
		},
		HelpGroup: consts.BuildHelpGroup,
		Run: func(c *grumble.Context) error {
			if err := buildCheck(); err != nil {
				return err
			}
			profilesLoadCmd(c.Args.String("name"))
			return nil
		},
		Completer: func(prefix string, args []string) []string {
			if err := buildCheck(); err != nil {
				return nil
			}
			return completion.Profiles(ctx, prefix, builderConfig.builderID)
		},
	})

	cmd.AddCommand(&grumble.Command{
		Name: "rm",
		Help: "remove one or more saved profiles",
		Args: func(a *grumble.Args) {
			a.StringList("names", "names of existing profiles", grumble.Min(1))
		},
		HelpGroup: consts.BuildHelpGroup,
		Run: func(c *grumble.Context) error {
			if err := buildCheck(); err != nil {
				return err
			}
			profilesRmCmd(c.Args.StringList("names"))
			return nil
		},
		Completer: func(prefix string, args []string) []string {
			if err := buildCheck(); err != nil {
				return nil
			}
			return completion.Profiles(ctx, prefix, builderConfig.builderID)
		},
	})
	return cmd
}
