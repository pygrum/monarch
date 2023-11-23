package commands

import "github.com/pygrum/monarch/pkg/install"

func installCmd(repoUrl string, useCreds bool) {
	if err := install.NewRepo(repoUrl, useCreds); err != nil {
		cLogger.Error("failed to install %s: %v", repoUrl, err)
		return
	}
	cLogger.Success("installation successful")
}
