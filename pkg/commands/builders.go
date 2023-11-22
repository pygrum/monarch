package commands

import (
	"fmt"
	"github.com/pygrum/monarch/pkg/db"
	"os"
	"text/tabwriter"
	"time"
)

var w *tabwriter.Writer

func init() {
	w = tabwriter.NewWriter(os.Stdout, 1, 1, 2, ' ', 0)
}

// buildersCmd lists installed builders
func buildersCmd(args []string) {
	var builders []db.Builder
	if len(args) == 0 {
		if err := db.Find(&builders); err != nil {
			cLogger.Error("failed to retrieve installed builders: %v", err)
			return
		}
	} else {
		if err := db.FindConditional("builder_id IN ?", args, &builders); err != nil {
			if err = db.FindConditional("name IN ?", args, &builders); err != nil {
				cLogger.Error("failed to retrieve the specified builders: %v", err)
				return
			}
		}
	}
	header := "AGENT NAME\tVERSION\tAUTHOR\tINSTALLATION DATE\tID\t"
	_, _ = fmt.Fprintln(w, header)
	for _, builder := range builders {
		line := fmt.Sprintf("%s\t%s\t%s\t%s\t%s\t",
			builder.Name,
			builder.Version,
			builder.Author,
			builder.CreatedAt.Format(time.DateTime),
			builder.BuilderID,
		)
		_, _ = fmt.Fprintln(w, line)
	}
	_ = w.Flush()
}
