package cmds

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

var (
	cmdRoot = &cobra.Command{
		Use:   "owl",
		Short: "Hetzner Cloud helper",
	}
)

func init() {
	cmdRoot.AddCommand(cmdAgent)
	cmdRoot.AddCommand(cmdHCloud)
	cmdRoot.AddCommand(cmdTpl)
}

func Run() {
	if err := cmdRoot.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}
