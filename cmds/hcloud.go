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
			vars.ValidatePresence("cert-path", "cert-storage")
			vars.ValidateInclusion("cert-storage", []string{"fs", "consul", "vault"})

			if vars.Valid() {
				cfs := cloudh.TlsStorageById(vars.GetString("cert-storage"))
				tls := cloudh.AutoTls{
					Config: cloudh.TlsConfig{
						CertPathPrefix:    vars.GetString("cert-path"),
						AccountPathPrefix: "",
					},
					Storage:        cfs,
					AccountStorage: &cloudh.TlsNullStorage{}, // not needed for listing
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
				log.Fatal(vars.ErrorMessages())
			}
		},
	}

	hcloudTlsIssue = &cobra.Command{
		Use:   "issue",
		Short: "Issue new certificate using DNS challenge",
		Run: func(cmd *cobra.Command, args []string) {
			vars := tea.ParseEqArgs(args)
			vars.ValidatePresence("token", "email", "domains", "cert-path", "account-path", "cert-storage", "account-storage")
			vars.ValidateInclusion("cert-storage", []string{"fs", "consul", "vault"})
			vars.ValidateInclusion("account-storage", []string{"fs", "consul", "vault"})

			if vars.Valid() {
				cfs := cloudh.TlsStorageById(vars.GetString("cert-storage"))
				afs := cloudh.TlsStorageById(vars.GetString("account-storage"))
				tls := cloudh.AutoTls{
					Config: cloudh.TlsConfig{
						DnsToken:          vars.GetString("token"),
						Email:             vars.GetString("email"),
						Domains:           vars.GetStrings("domains", ","),
						CertPathPrefix:    vars.GetString("cert-path"),
						AccountPathPrefix: vars.GetString("account-path"),
						Debug:             vars.GetBoolDefault("debug", false),
					},
					Storage:        cfs,
					AccountStorage: afs,
				}

				err := tls.Issue()
				if err != nil {
					log.Fatal(err)
				}
			} else {
				log.Fatal(vars.ErrorMessages())
			}
		},
	}

	hcloudTlsRenew = &cobra.Command{
		Use:   "renew",
		Short: "Attempts certifcate renowal",
		Run: func(cmd *cobra.Command, args []string) {
			vars := tea.ParseEqArgs(args)
			vars.ValidatePresence("token", "email", "domains", "cert-path", "account-path", "cert-storage", "account-storage")
			vars.ValidateInclusion("cert-storage", []string{"fs", "consul", "vault"})
			vars.ValidateInclusion("account-storage", []string{"fs", "consul", "vault"})

			if vars.Valid() {
				cfs := cloudh.TlsStorageById(vars.GetString("cert-storage"))
				afs := cloudh.TlsStorageById(vars.GetString("account-storage"))
				tls := cloudh.AutoTls{
					Config: cloudh.TlsConfig{
						DnsToken:          vars.GetString("token"),
						Email:             vars.GetString("email"),
						Domains:           vars.GetStrings("domains", ","),
						CertPathPrefix:    vars.GetString("cert-path"),
						AccountPathPrefix: vars.GetString("account-path"),
						Debug:             vars.GetBoolDefault("debug", false),
					},
					Storage:        cfs,
					AccountStorage: afs,
				}

				err := tls.Renew(false)
				if err != nil {
					log.Fatal(err)
				}
			} else {
				log.Fatal(vars.ErrorMessages())
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
	cmdHCloudTls.AddCommand(hcloudTlsIssue)
	cmdHCloudTls.AddCommand(hcloudTlsRenew)
}
