package cmds

import (
	"os"

	"github.com/qbart/ohowl/cloudh"
	"github.com/spf13/cobra"
)

var (
	cmdTls = &cobra.Command{Use: "tls", Short: "Certificates"}

	tlsAuto = &cobra.Command{
		Use:   "auto",
		Short: "Download cert",
		Run: func(cmd *cobra.Command, args []string) {
			autoTls := cloudh.ConfigAutoTls(cloudh.Dns{
				Token: os.Getenv("HCLOUD_DNS_TOKEN"),
				Email: os.Getenv("EMAIL"),
			}, true)

			autoTls.Start([]string{os.Getenv("ZONE")})
		},
	}
)

func init() {
	cmdTls.AddCommand(tlsAuto)
}
