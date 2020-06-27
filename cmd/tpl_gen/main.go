package main

import (
	"bytes"
	"fmt"
	"go/format"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
	"text/template"

	"gopkg.in/yaml.v2"
)

func main() {
	if len(os.Args) != 2 {
		errorExit(-1, "need a generation file\n")
	}
	gd, err := loadGenData(os.Args[1])
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
	Imports           []string `yaml:"imports"`
	importedValueSets map[string]valueMap
	ValueSets         map[string]valueMap `yaml:"value_sets"`
	Templates         []*genValues        `yaml:"templates"`
}

type genValues struct {
	Template  string   `yaml:"template"`
	Out       string   `yaml:"out"`
	ValueSets []string `yaml:"value_sets"`
	Values    valueMap `yaml:"values"`
}

func loadGenData(f string) (*genData, error) {
	r, err := readGenDataFile(f)
	if err != nil {
		return nil, err
	}
	r.importedValueSets = make(map[string]valueMap, 16)
	for _, i := range r.Imports {
		fn := filepath.Join(filepath.Dir(f), i)
		gd, err := readGenDataFile(fn)
		if err != nil {
			return nil, err
		}
		mergeValueSets(gd.ValueSets, r.importedValueSets)
	}
	return r, nil
}

func readGenDataFile(f string) (*genData, error) {
	rc, err := os.Open(f)
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

func dashedToCamel(s string) string {
	parts := strings.Split(s, "-")
	for i := 1; i < len(parts); i++ {
		parts[i] = strings.Title(parts[i])
	}
	return strings.Join(parts, "")
}

var funcMap = template.FuncMap{
	"default":         defaultString,
	"lower":           strings.ToLower,
	"upper":           strings.ToUpper,
	"dashed_to_camel": dashedToCamel,
	"title":           strings.Title,
}

func generateCode(gd *genData) error {
	vs := make(map[string]valueMap, len(gd.ValueSets)+len(gd.importedValueSets))
	mergeValueSets(gd.importedValueSets, vs)
	mergeValueSets(gd.ValueSets, vs)
	for _, i := range gd.Templates {
		if i.Values == nil {
			i.Values = make(valueMap, 16)
		}
		for _, j := range i.ValueSets {
			mergeValues(vs[j], i.Values)
		}
		if err := generateFromTemplate(i); err != nil {
			return err
		}
	}
	return nil
}

func generateFromTemplate(gv *genValues) error {
	f, err := os.Create(gv.Out)
	if err != nil {
		return err
	}
	defer f.Close()
	b, err := ioutil.ReadFile(gv.Template)
	if err != nil {
		return err
	}
	t, err := template.New("main").Funcs(funcMap).Parse(string(b))
	if err != nil {
		return err
	}
	bb := bytes.NewBuffer(make([]byte, 0, 1024))
	if err = t.Execute(bb, valueMap{"Values": gv.Values}); err != nil {
		return err
	}
	if b, err = format.Source(bb.Bytes()); err != nil {
		return err
	}
	_, err = io.Copy(f, bytes.NewReader(b))
	return err
}

func mergeValues(src, dst valueMap) valueMap {
	for k, v := range src {
		dst[k] = v
	}
	return dst
}

func mergeValueSets(src, dst map[string]valueMap) map[string]valueMap {
	for k, v := range src {
		dst[k] = v
	}
	return dst
}
