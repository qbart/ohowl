package cmds

import (
	"io/ioutil"
	"log"
	"os"
	"strings"
	"text/template"

	"github.com/spf13/cobra"
)

var (
	cmdTpl = &cobra.Command{Use: "tpl", Short: "Templates"}

	tplRender = &cobra.Command{
		Use:   "render",
		Short: "Renders tpl file and replaces variables {{.var}}",
		Long:  `render FILE var=1 var2=... (do not use spaces between =)`,
		Args: func(cmd *cobra.Command, args []string) error {
			if err := cobra.MinimumNArgs(1)(cmd, args); err != nil {
				return err
			}

			return nil
		},
		Run: func(cmd *cobra.Command, args []string) {
			path := args[0]

			file, err := os.Open(path)
			if err != nil {
				log.Fatal(err)
			}
			defer file.Close()

			b, err := ioutil.ReadAll(file)

			vars := make(map[string]interface{}, 0)

			for _, arg := range args[1:] {
				kv := strings.SplitN(arg, "=", 2)
				vars[kv[0]] = kv[1]
			}

			tmpl, err := template.New(path).Parse(string(b))
			if err != nil {
				log.Fatal(err)
			}

			if err := tmpl.Execute(os.Stdout, vars); err != nil {
				log.Fatal(err)
			}
		},
	}
)

func init() {
	cmdTpl.AddCommand(tplRender)
}
