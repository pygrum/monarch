package commands

import (
	"fmt"
	"github.com/pygrum/monarch/pkg/config"
	"github.com/pygrum/monarch/pkg/db"
	"github.com/pygrum/monarch/pkg/handler/http"
	"path/filepath"
	"strings"
)

func stageCmd(args []string, format, as string) {
	if len(args) == 0 {
		m := http.Stage.View()
		stagefmt := "%s (%s)\tstaged as %s\t%s\t"
		for k, v := range *m {
			fileFormat := ": " + v.Format
			if len(fileFormat) == 2 {
				fileFormat = ""
			}
			_, _ = fmt.Fprintln(w, fmt.Sprintf(stagefmt, v.Path, v.Agent,
				strings.ReplaceAll(config.MainConfig.StageEndpoint, "{file}", k), fileFormat))
		}
		w.Flush()
		return
	}
	name := args[0]
	agent := &db.Agent{}
	// only finds one so shouldn't be an issue
	if err := db.FindOneConditional("agent_id = ?", name, &agent); err != nil {
		if err = db.FindOneConditional("name = ?", name, &agent); err != nil {
			cLogger.Error("failed to retrieve the specified agent: %v", err)
			return
		}
	}
	if len(as) == 0 {
		as = filepath.Base(agent.File)
	}
	http.Stage.Add(as, agent.Name, agent.File, format)
	l.Info("staged %s on %s", agent.File, strings.ReplaceAll(config.MainConfig.StageEndpoint, "{file}", as))
}

func unstageCmd(name string) {
	http.Stage.Rm(name)
}
