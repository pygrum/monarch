package lint

import (
	"errors"
	"fmt"
	"github.com/pygrum/monarch/pkg/commands"
	"github.com/pygrum/monarch/pkg/config"
	"github.com/sirupsen/logrus"
	"net/url"
	"regexp"
	"slices"
	"strings"
)

var (
	err               error
	allOpcodes        []int
	alphaNumericRegex = regexp.MustCompile(`^[a-zA-Z0-9_-]+$`)
	semVerRegex       = regexp.MustCompile(`^(0|[1-9]\d*)\.(0|[1-9]\d*)\.(0|[1-9]\d*)(?:-((?:0|[1-9]\d*|\d*[a-zA-Z-][0-9a-zA-Z-]*)(?:\.(?:0|[1-9]\d*|\d*[a-zA-Z-][0-9a-zA-Z-]*))*))?(?:\+([0-9a-zA-Z-]+(?:\.[0-9a-zA-Z-]+)*))?$`)
)

func Lint(configFile string) {
	royal := &config.ProjectConfig{}
	if err := config.YamlConfig(configFile, royal); err != nil {
		logrus.Errorf("failed to parse configuration: %v", err)
		return
	}
	if err := lint(royal); err != nil {
		logrus.Fatalf("INVALID royal.yaml")
	}
	logrus.Info("VALID royal.yaml")
}

func check(str string, checkErr error) {
	if checkErr != nil {
		logrus.Errorf("linter detected an issue with field '%s': %v", str, checkErr)
		err = checkErr
	}
}

func lint(cfg *config.ProjectConfig) error {
	check(naming("name", cfg.Name))
	check(versioning(cfg.Version))
	check(sourceUrl(cfg.URL))
	for _, cmd := range cfg.CmdSchema {
		check(cmdSchema(&cmd))
	}
	for _, arg := range cfg.Builder.BuildArgs {
		check(buildArg(&arg))
	}
	return err
}

func naming(field, value string) (string, error) {
	// Check for alphanumeric
	if !alphaNumericRegex.MatchString(value) {
		return field, errors.New("name must be alphanumeric with special characters ( _ - ) allowed")
	}
	return "", nil
}

func versioning(version string) (string, error) {
	if !semVerRegex.MatchString(version) {
		return "version", errors.New("invalid versioning: must be semantic (https://semver.org/)")
	}
	return "", nil
}

func sourceUrl(u string) (string, error) {
	if len(u) == 0 {
		return "", nil
	}
	if _, err := url.ParseRequestURI(u); err != nil {
		return "url", errors.New("invalid project url provided")
	}
	return "", nil
}

func cmdSchema(cmd *config.ProjectConfigCmd) (string, error) {
	if slices.Contains(allOpcodes, int(cmd.Opcode)) {
		return "cmd_schema." + cmd.Name + ".opcode", errors.New("duplicate opcode found")
	}
	allOpcodes = append(allOpcodes, int(cmd.Opcode))
	if cmd.MinArgs < 0 {
		return "cmd_schema." + cmd.Name + ".min_args", errors.New("cannot be below 0")
	}
	if cmd.MinArgs > cmd.MaxArgs && cmd.MaxArgs >= 0 {
		return "cmd_schema." + cmd.Name + ".min_args", errors.New("cannot be greater than max_args")
	}
	return "", nil
}

func buildArg(arg *config.ProjectConfigBuildArg) (string, error) {
	field := "builder.build_args." + arg.Name
	if !alphaNumericRegex.MatchString(arg.Name) {
		return field, errors.New("name must be alphanumeric with special characters ( _ - ) allowed")
	}
	if !slices.Contains(commands.ValidTypes, arg.Type) {
		return field, errors.New(arg.Type + " is not a valid type ( must be " + strings.Join(commands.ValidTypes, ", ") + ")")
	}
	if err := commands.TypeVerify(arg.Type, arg.Default); err != nil {
		return field, fmt.Errorf("bad default value '%s': %v", arg.Default, err)
	}
	for _, choice := range arg.Choices {
		if err := commands.TypeVerify(arg.Type, choice); err != nil {
			return field + ".choices", fmt.Errorf("invalid choice '%s': %v", choice, err)
		}
	}
	return "", nil
}
