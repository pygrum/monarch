package commands

import (
	"fmt"
	"github.com/pygrum/monarch/pkg/completion"
	"github.com/pygrum/monarch/pkg/console"
	"github.com/pygrum/monarch/pkg/protobuf/clientpb"
	"github.com/rsteube/carapace"
	"github.com/spf13/cobra"
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
		Immutables: immutables,
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

func cobraProfilesCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "profiles [flags] NAMES...",
		Short: "list all created profiles",
		Run: func(cmd *cobra.Command, args []string) {
			profilesCmd(args)
		},
	}
	carapace.Gen(cmd).PositionalCompletion(completion.Profiles(ctx, builderConfig.builderID))
	saveCmd := &cobra.Command{
		Use:   "save [flags] NAME",
		Short: "save current build configuration options as a new profile",
		Args:  cobra.ExactArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			profilesSaveCmd(args[0])
		},
	}
	loadCmd := &cobra.Command{
		Use:   "load [flags] NAME",
		Short: "load an existing profile",
		Args:  cobra.ExactArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			profilesLoadCmd(args[0])
		},
	}
	carapace.Gen(loadCmd).PositionalCompletion(completion.Profiles(ctx, builderConfig.builderID))

	rmCmd := &cobra.Command{
		Use:   "rm [flags] NAMES...",
		Short: "remove one or more saved profiles",
		Args:  cobra.MinimumNArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			profilesRmCmd(args)
		},
	}
	carapace.Gen(rmCmd).PositionalCompletion(completion.Profiles(ctx, builderConfig.builderID))

	cmd.AddCommand(saveCmd, loadCmd, rmCmd)
	return cmd
}
