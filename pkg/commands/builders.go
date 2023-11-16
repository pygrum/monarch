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
func buildersCmd() {
	var builders []db.Builder
	if err := db.Find(&builders); err != nil {
		cLogger.Error("failed to retrieve installed builders: %v", err)
		return
	}
	header := "AGENT NAME\tVERSION\tID\tAUTHOR\tINSTALLATION DATE\t"
	_, _ = fmt.Fprintln(w, header)
	for _, builder := range builders {
		line := fmt.Sprintf("%s\t%s\t%s\t%s\t%s\t",
			builder.Name,
			builder.Version,
			builder.BuilderID,
			builder.Author,
			builder.CreatedAt.Format(time.DateTime),
		)
		_, _ = fmt.Fprintln(w, line)
	}
	_ = w.Flush()
}
