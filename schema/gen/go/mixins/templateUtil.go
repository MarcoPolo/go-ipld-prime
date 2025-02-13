package mixins

import (
	"io"
	"strings"
	"text/template"

	"github.com/ipld/go-ipld-prime/testutil"
)

func doTemplate(tmplstr string, w io.Writer, data interface{}) {
	tmpl := template.Must(template.New("").
		Funcs(template.FuncMap{
			"title": func(s string) string { return strings.Title(s) },
		}).
		Parse(testutil.Dedent(tmplstr)))
	if err := tmpl.Execute(w, data); err != nil {
		panic(err)
	}
}
