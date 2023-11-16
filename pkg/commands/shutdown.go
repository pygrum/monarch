package commands

import (
	"os"

	"github.com/AlecAivazis/survey/v2"
)

func exitCmd() {
	y := false
	prompt := &survey.Confirm{
		Message: "are you sure you want to quit? All listeners and sessions will be closed",
	}
	_ = survey.AskOne(prompt, &y)
	if y {
		cLogger.Info("Goodbye!")
		_ = cLogger.Close()

		os.Exit(0)
	}
}
