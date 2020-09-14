package cmds

import (
	"io/ioutil"
	"log"
	"os"

	"github.com/qbart/ohowl/tea"
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

			vars := tea.ParseEqArgs(args[1:])
			err = tea.TplRender(os.Stdout, b, vars.Raw)
			if err != nil {
				log.Fatal(err)
			}
		},
	}
)

func init() {
	cmdTpl.AddCommand(tplRender)
}
