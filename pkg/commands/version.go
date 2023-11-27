package commands

import (
	"fmt"
	"github.com/pygrum/monarch/pkg/consts"
	"runtime"
)

func versionCmd() {
	fmt.Printf("monarch v%s %s/%s\n", consts.Version, runtime.GOOS, runtime.GOARCH)
}
