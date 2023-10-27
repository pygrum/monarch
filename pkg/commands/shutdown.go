package commands

import (
	"os"

	"github.com/AlecAivazis/survey/v2"
)

func shutdownCmd() {
	y := false
	prompt := &survey.Confirm{
		Message: "are you sure you want to quit? All listeners and sessions will be closed",
	}
	survey.AskOne(prompt, &y)
	if y {
		cLogger.Info("Goodbye!")
		cLogger.Close()

		os.Exit(0)
	}
}
