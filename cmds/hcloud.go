package cmds

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"

	"github.com/olekukonko/tablewriter"
	"github.com/qbart/ohowl/cloudh"
	"github.com/qbart/ohowl/tea"
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
			fmt.Println(string(tea.MustJson(data)))
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

	hcloudTlsList = &cobra.Command{
		Use:   "list",
		Short: "List certificates",
		Run: func(cmd *cobra.Command, args []string) {
			vars := tea.ParseEqArgs(args)
			if vars.Exist("path") {
				tls := cloudh.AutoTls{
					Config: cloudh.TlsConfig{
						Path: vars.GetString("path"),
					},
					Storage: &cloudh.TlsFileStorage{},
				}

				certs, err := tls.List()
				if err != nil {
					log.Fatal(err)
				}

				table := tablewriter.NewWriter(os.Stdout)
				table.SetHeader([]string{"Common Name", "DNS", "Expiry", "File"})

				for _, cert := range certs {
					table.Append([]string{
						cert.CommonName,
						strings.Join(cert.DNS, ", "),
						cert.Expiry.String(),
						filepath.Base(cert.Path),
					})
				}
				table.Render()

			} else {
				log.Fatal("Missing required params: path=")
			}
		},
	}

	hcloudTlsAuto = &cobra.Command{
		Use:   "auto",
		Short: "DNS challenge",
		Run: func(cmd *cobra.Command, args []string) {
			vars := tea.ParseEqArgs(args)
			if vars.Exist("token", "email", "zones", "path") {
				tls := cloudh.AutoTls{
					Config: cloudh.TlsConfig{
						Token:   vars.GetString("token"),
						Email:   vars.GetString("email"),
						Domains: vars.GetStrings("zones", ","),
						Path:    vars.GetString("path"),
						Debug:   vars.GetBoolDefault("debug", false),
					},
					Storage: &cloudh.TlsFileStorage{},
				}

				err := tls.IssueNew()
				if err != nil {
					log.Fatal(err)
				}
			} else {
				log.Fatal("Missing required params: token= email= zones= path=")
			}
		},
	}
)

func init() {
	cmdHCloud.AddCommand(hcloudMetadata)
	cmdHCloud.AddCommand(hcloudWaitForIp)
	cmdHCloud.AddCommand(hcloudServers)
	cmdHCloud.AddCommand(cmdHCloudTls)
	cmdHCloudTls.AddCommand(hcloudTlsList)
	cmdHCloudTls.AddCommand(hcloudTlsAuto)
}
