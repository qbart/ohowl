package tea

import (
	"html/template"
	"io"
)

func TplRender(writer io.Writer, bytes []byte, data interface{}) error {
	tmpl, err := template.New("t").Parse(string(bytes))
	if err != nil {
		return err
	}

	return tmpl.Execute(writer, data)
}
