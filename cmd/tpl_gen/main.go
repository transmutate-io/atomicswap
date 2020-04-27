package main

import (
	"fmt"
	"io/ioutil"
	"os"
	"strings"
	"text/template"

	"gopkg.in/yaml.v2"
)

func main() {
	if len(os.Args) != 2 {
		errorExit(-1, "need a generation file\n")
	}
	gd, err := readGenData(os.Args[1])
	if err != nil {
		errorExit(-2, "can't read generation data: %s\n", err.Error())
	}
	if err = generateCode(gd); err != nil {
		errorExit(-3, "can't generate code: %s\n", err.Error())
	}
}

func errorExit(code int, f string, a ...interface{}) {
	fmt.Fprintf(os.Stderr, f, a...)
	os.Exit(code)
}

type valueMap = map[string]interface{}

type genData struct {
	ValueSets map[string]valueMap `yaml:"value_sets"`
	Templates []*genValues        `yaml:"templates"`
}

type genValues struct {
	Template  string   `yaml:"template"`
	Out       string   `yaml:"out"`
	ValueSets []string `yaml:"value_sets"`
	Values    valueMap `yaml:"values"`
}

func readGenData(f string) (*genData, error) {
	rc, err := os.Open(os.Args[1])
	if err != nil {
		return nil, err
	}
	defer rc.Close()
	r := &genData{}
	if err = yaml.NewDecoder(rc).Decode(r); err != nil {
		return nil, err
	}
	return r, nil
}

func defaultString(d, s string) string {
	if s != "" {
		return s
	}
	return d
}

var funcMap = template.FuncMap{
	"default": defaultString,
	"lower":   strings.ToLower,
	"upper":   strings.ToUpper,
}

func generateCode(gd *genData) error {
	for i := range gd.Templates {
		if err := generateFromTemplate(gd, i); err != nil {
			return err
		}
	}
	return nil
}

func generateFromTemplate(gd *genData, idx int) error {
	f, err := os.Create(gd.Templates[idx].Out)
	if err != nil {
		return err
	}
	defer f.Close()
	b, err := ioutil.ReadFile(gd.Templates[idx].Template)
	if err != nil {
		return err
	}
	t, err := template.New("main").Funcs(funcMap).Parse(string(b))
	if err != nil {
		return err
	}

	vals := make(map[string]interface{}, 32)
	for _, i := range gd.Templates[idx].ValueSets {
		mergeValues(gd.ValueSets[i], vals)
	}
	mergeValues(gd.Templates[idx].Values, vals)

	return t.Execute(f, map[string]interface{}{"All": gd, "Values": vals})
}

func mergeValues(src, dst map[string]interface{}) {
	for k, v := range src {
		dst[k] = v
	}
}
