package commands

import (
	"github.com/pygrum/monarch/pkg/db"
	"github.com/pygrum/monarch/pkg/utils"
	"os"
	"strings"
)

func uninstallCmd(args []string) {
	var builders []db.Builder
	if err := db.FindConditional("builder_id IN ?", args, &builders); err != nil {
		cLogger.Error("failed to retrieve the specified builders: %v", err)
		return
	}
	if len(builders) == 0 {
		if err := db.FindConditional("name IN ?", args, &builders); err != nil {
			cLogger.Error("failed to retrieve the specified builders: %v", err)
			return
		}
		if len(builders) == 0 {
			cLogger.Error("no builders named %s", strings.Join(args, ", "))
			return
		}
	}
	for _, b := range builders {
		cLogger.Info("deleting %s...", b.Name)
		if err := os.RemoveAll(b.InstalledAt); err != nil {
			cLogger.Error("failed to remove install folder: %v", err)
			return
		}
		if err := utils.Cleanup(&b); err != nil {
			cLogger.Error("%v", err)
			return
		}
		cLogger.Success("%s v%s deleted", b.Name, b.Version)
	}
}
