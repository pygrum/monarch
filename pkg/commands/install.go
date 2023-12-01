package commands

import (
	"github.com/pygrum/monarch/pkg/config"
	"github.com/pygrum/monarch/pkg/install"
	"os"
	"path/filepath"
	"strings"
)

func installCmd(repoUrl, branch string, useCreds bool) {
	if err := install.NewRepo(repoUrl, branch, useCreds); err != nil {
		cLogger.Error("failed to install %s: %v", repoUrl, err)
		clonePath := filepath.Join(config.MainConfig.InstallDir, strings.TrimSuffix(filepath.Base(repoUrl),
			filepath.Ext(filepath.Base(repoUrl))))
		if err = os.RemoveAll(clonePath); err != nil {
			cLogger.Error("failed to remove %s: %v. must be manually removed", clonePath, err)
		}
		return
	}
	cLogger.Success("installation successful")
}
