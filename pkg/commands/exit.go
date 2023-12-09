package commands

import (
	"os"
)

func exitCmd() {
	cLogger.Info("Goodbye!")
	_ = cLogger.Close()

	os.Exit(0)
}
