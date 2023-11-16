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
func buildersCmd(showTranslator bool) {
	var builders []db.Builder
	if err := db.Find(&builders); err != nil {
		cLogger.Error("failed to retrieve installed builders: %v", err)
		return
	}
	header := "AGENT NAME\tVERSION\tAUTHOR\tINSTALLATION DATE\tID\t"
	if showTranslator {
		header += "TRANSLATOR ID\t"
	}
	_, _ = fmt.Fprintln(w, header)
	for _, builder := range builders {
		line := fmt.Sprintf("%s\t%s\t%s\t%s\t%s\t",
			builder.Name,
			builder.Version,
			builder.Author,
			builder.CreatedAt.Format(time.DateTime),
			builder.BuilderID,
		)
		if showTranslator {
			line += fmt.Sprintf("%s\t", builder.TranslatorID)
		}
		_, _ = fmt.Fprintln(w, line)
	}
	_ = w.Flush()
}
