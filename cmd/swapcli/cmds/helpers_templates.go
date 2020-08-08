package cmds

var cryptosListTemplates = []string{
	"{{ .Name }}\n",
	"{{ .Name }}, {{ .Short }}\n",
	"{{ .Name }}, {{ .Short }}, {{ .Decimals }} decimal places\n",
	"{{ .Name }}, {{ .Short }}, {{ .Decimals }} decimal places, {{ .Type }}\n",
}
