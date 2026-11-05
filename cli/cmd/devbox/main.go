// Command devbox is the command-line client for the DevBox Sandbox Manager.
package main

import (
	"errors"
	"os"
	"path/filepath"

	devboxcmd "github.com/huaweicloud/devspace-devbox/cli/cmd"
	"github.com/huaweicloud/devspace-devbox/cli/internal/helpers"
)

func main() {
	devboxcmd.RootCmd.Use = filepath.Base(os.Args[0])
	if err := devboxcmd.RootCmd.Execute(); err != nil {
		var ee *helpers.ExitCodeError
		if errors.As(err, &ee) {
			os.Exit(ee.ExitCode)
		}
		os.Exit(1)
	}
}
