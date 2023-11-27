package commands

import (
	"github.com/pygrum/monarch/pkg/db"
	"github.com/pygrum/monarch/pkg/install"
)

// installs local repositories / folders
func localCmd(path string) {
	builder, err := install.Setup(path)
	if err != nil {
		l.Error("failed to setup local repository: %v", err)
		return
	}
	if err = db.Create(builder); err != nil {
		l.Error("failed to save new builder: %v", err)
		return
	}
	l.Info("successfully installed %s", path)
}
