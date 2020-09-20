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
		if bootArgs.Exist("acltoken", "account-path", "cert-path") {
			app := web.App{
				AclToken:          bootArgs.GetString("acltoken"),
				Debug:             bootArgs.GetBoolDefault("debug", false),
				AccountPathPrefix: bootArgs.GetString("account-path"),
				CertPathPrefix:    bootArgs.GetString("cert-path"),
			}
			app.Run()
		} else {
			log.Println("Missing one of acltoken= account-path= cert-path=")
		}
	},
}
