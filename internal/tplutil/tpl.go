package tplutil

import (
	"text/template"

	"github.com/spf13/pflag"
	"github.com/transmutate-io/atomicswap/internal/cmdutil"
	"github.com/transmutate-io/atomicswap/internal/flagutil"
	"github.com/transmutate-io/atomicswap/internal/flagutil/exitcodes"
)

type TemplateData = map[string]interface{}

func OpenTemplate(format string, verbose int, tpls []string, funcs template.FuncMap) (*template.Template, error) {
	var tplStr string
	if format != "" {
		tplStr = format
	} else {
		tplStr = tpls[verbose]
	}
	r := template.New("main")
	if funcs != nil {
		r = r.Funcs(funcs)
	}
	return r.Parse(tplStr)
}

func MustOpenTemplate(fs *pflag.FlagSet, tpls []string, funcs template.FuncMap) *template.Template {
	r, err := OpenTemplate(
		flagutil.MustFormat(fs),
		flagutil.MustVerboseLevel(fs, len(tpls)-1),
		tpls,
		funcs,
	)
	if err != nil {
		cmdutil.ErrorExit(exitcodes.BadTemplate, err)
	}
	return r
}
