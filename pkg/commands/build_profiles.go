package commands

import (
	"fmt"
	"github.com/pygrum/monarch/pkg/db"
	"github.com/spf13/cobra"
	"slices"
	"strings"
	"time"
)

func profilesCmd(names []string) {
	var profiles []db.Profile
	if len(names) > 0 {
		if err := db.FindConditional("name IN ?", names, &profiles); err != nil {
			cLogger.Error("failed to find profiles(s): %v", err)
			return
		}
	} else {
		if err := db.Find(&profiles); err != nil {
			cLogger.Error("failed to find profiles(s): %v", err)
			return
		}
	}
	headers := "NAME\tCREATION TIME\t"
	_, _ = fmt.Fprintln(w, headers)
	for _, p := range profiles {
		line := fmt.Sprintf("%s\t%s\t", p.Name, p.CreatedAt.Format(time.DateTime))
		_, _ = fmt.Fprintln(w, line)
	}
	_ = w.Flush()
}

func profilesSaveCmd(name string) {
	profile := &db.Profile{}
	if db.Where("name = ? AND builder_id = ?", name, builderConfig.builderID).Find(&profile); len(profile.Name) != 0 {
		cLogger.Error("a profile for this build named '%s' already exists", name)
		return
	}
	profile = &db.Profile{
		Name:      name,
		BuilderID: builderConfig.builderID,
	}
	var records []db.ProfileRecord
	for k, v := range builderConfig.request.Options {
		record := db.ProfileRecord{
			Profile: name,
			Name:    k,
			Value:   v,
		}
		records = append(records, record)
	}
	if err := db.Create(profile); err != nil {
		cLogger.Error("failed to create new profile: %v", err)
		return
	}
	if err := db.Create(records); err != nil {
		cLogger.Error("failed to save profile values: %v", err)
		return
	}
	cLogger.Success("profile saved as '%s'", name)
}

func profilesLoadCmd(name string) {
	profile := &db.Profile{}
	if err := db.Where("name = ? AND builder_id = ?", name, builderConfig.builderID).Find(profile).Error; err != nil {
		cLogger.Error("failed to find %s: %v", name, err)
		return
	}
	var records []db.ProfileRecord
	if err := db.FindConditional("profile = ?", name, &records); err != nil {
		cLogger.Error("failed to get profile values: %v", err)
		return
	}
	for _, r := range records {
		if slices.Contains(immutables, r.Name) {
			continue
		}
		setCmd(r.Name, r.Value)
	}
	l.Success("profile %s loaded", profile.Name)
}

func profilesRmCmd(names []string) {
	var profiles []db.Profile
	if err := db.Where("name IN ? AND builder_id = ?", names, builderConfig.builderID).Find(&profiles).Error; err != nil {
		cLogger.Error("failed to find profiles(s): %v", err)
		return
	}
	var records []db.ProfileRecord
	if err := db.FindConditional("profile IN ?", names, &records); err != nil {
		cLogger.Error("failed to find profile values: %v", err)
		return
	}
	if err := db.Delete(records); err != nil {
		cLogger.Error("failed to delete profile values: %v", err)
	}
	if err := db.Delete(profiles); err != nil {
		cLogger.Error("failed to delete profiles: %v", err)
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
	rmCmd := &cobra.Command{
		Use:   "rm [flags] NAMES...",
		Short: "remove one or more saved profiles",
		Args:  cobra.MinimumNArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			profilesRmCmd(args)
		},
	}
	cmd.AddCommand(saveCmd, loadCmd, rmCmd)
	return cmd
}
