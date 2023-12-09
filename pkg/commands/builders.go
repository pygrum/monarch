package commands

import (
	"fmt"
	"github.com/pygrum/monarch/pkg/console"
	"github.com/pygrum/monarch/pkg/protobuf/clientpb"
	"os"
	"text/tabwriter"
)

var w *tabwriter.Writer

func init() {
	w = tabwriter.NewWriter(os.Stdout, 1, 1, 2, ' ', 0)
}

// buildersCmd lists installed builders
func buildersCmd(args []string) {
	builders, err := console.Rpc.Builders(ctx, &clientpb.BuilderRequest{BuilderId: args})
	if err != nil {
		cLogger.Error("failed to retrieve builders: %v", err)
		return
	}
	header := "AGENT NAME\tVERSION\tAUTHOR\tINSTALLATION DATE\tID\tRUNS ON\t"
	_, _ = fmt.Fprintln(w, header)
	for _, builder := range builders.Builders {
		line := fmt.Sprintf("%s\t%s\t%s\t%s\t%s\t%s\t",
			builder.Name,
			builder.Version,
			builder.Author,
			builder.CreatedAt,
			builder.BuilderId,
			builder.Supported_OS,
		)
		_, _ = fmt.Fprintln(w, line)
	}
	_ = w.Flush()
}
