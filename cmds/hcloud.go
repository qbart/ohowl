package cmds

import (
	"fmt"
	"log"
	"os"

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

	hcloudWaitForIp = &cobra.Command{
		Use:   "wait",
		Short: "Wait for private IP to be assigned",
		Run: func(cmd *cobra.Command, args []string) {
			if ok := cloudh.WaitForIp(); !ok {
				os.Exit(1)
			}
		},
	}
)

func init() {
	cmdHCloud.AddCommand(hcloudMetadata)
	cmdHCloud.AddCommand(hcloudWaitForIp)
}
