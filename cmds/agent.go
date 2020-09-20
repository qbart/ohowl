package cmds

import (
	"log"

	"github.com/qbart/ohowl/tea"
	"github.com/qbart/ohowl/web"
	"github.com/spf13/cobra"
)

var cmdAgent = &cobra.Command{
	Use:   "agent",
	Short: "Start HTTP server on 1914 port",
	Run: func(cmd *cobra.Command, args []string) {
		bootArgs := tea.ParseEqArgs(args)
		if bootArgs.Exist("acltoken") {
			app := web.App{
				Token: bootArgs.GetString("acltoken"),
				Debug: bootArgs.GetBoolDefault("debug", false),
			}
			app.Run()
		} else {
			log.Println("Missing acltoken=")
		}
	},
}
