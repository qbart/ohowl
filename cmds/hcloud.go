package cmds

import (
	"fmt"
	"log"
	"os"
	"strings"

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

	hcloudServers = &cobra.Command{
		Use:   "servers",
		Short: "Get servers by labels",
		Run: func(cmd *cobra.Command, args []string) {
			byLabels := cloudh.LabelSelector{
				MustLabels:    make(map[string]string, 0),
				MustNotLabels: make(map[string]string, 0),
			}

			// build filters from args:
			// key=aaa other!=bbb
			for _, arg := range args {
				kv := strings.SplitN(arg, "=", 2)
				last := kv[0][len(kv[0])-1]
				if last == '!' {
					byLabels.MustNotLabels[kv[0][:len(kv[0])-1]] = kv[1]
				} else {
					byLabels.MustLabels[kv[0]] = kv[1]
				}
			}

			data, err := cloudh.GetServers(os.Getenv("HCLOUD_TOKEN"), cloudh.ServerFilter{
				ByLabel: byLabels,
			})
			if err != nil {
				log.Fatal(err)
			}
			fmt.Println(string(data))
		},
	}

	cmdHCloudTls = &cobra.Command{Use: "tls", Short: "Certificates"}

	hcloudTlsAuto = &cobra.Command{
		Use:   "auto",
		Short: "DNS challenge",
		Run: func(cmd *cobra.Command, args []string) {
			autoTls := cloudh.ConfigAutoTls(cloudh.Dns{
				Token: os.Getenv("HCLOUD_DNS_TOKEN"),
				Email: os.Getenv("EMAIL"),
			}, true)

			if err := autoTls.Start(strings.Split(os.Getenv("ZONE"), ",")); err != nil {
				log.Fatal(err)
			}
		},
	}
)

func init() {
	cmdHCloud.AddCommand(hcloudMetadata)
	cmdHCloud.AddCommand(hcloudWaitForIp)
	cmdHCloud.AddCommand(hcloudServers)
	cmdHCloud.AddCommand(cmdHCloudTls)
	cmdHCloudTls.AddCommand(hcloudTlsAuto)
}
