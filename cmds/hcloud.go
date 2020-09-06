package cmds

import (
	"fmt"
	"log"

	"github.com/qbart/ohowl/cloudh"
	"github.com/qbart/ohowl/utils"
	"github.com/spf13/cobra"
)

var (
	cmdHCloud = &cobra.Command{Use: "hcloud", Short: "Hetzner Cloud"}

	hcloudMetadata = &cobra.Command{
		Use:   "metadata",
		Short: "Get server metadata",
		Run: func(cmd *cobra.Command, args []string) {
			data, err := cloudh.GetMetadata()
			if err != nil {
				log.Fatal(err)
			}
			fmt.Println(string(utils.Json(data)))
		},
	}
)

func init() {
	cmdHCloud.AddCommand(hcloudMetadata)
}
