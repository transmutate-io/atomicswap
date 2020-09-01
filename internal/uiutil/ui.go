package uiutil

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
	"syscall"

	"github.com/c-bata/go-prompt"
	"github.com/spf13/cobra"
	"github.com/transmutate-io/atomicswap/cryptos"
	"github.com/transmutate-io/atomicswap/internal/cmdutil"
	"github.com/transmutate-io/cryptocore/types"
)

var (
	UpCommand = &MenuCompleter{
		Suggestion: &prompt.Suggest{
			Text:        "..",
			Description: "move to the Parent menu",
		},
	}
	HelpCommand = &MenuCompleter{
		Suggestion: &prompt.Suggest{
			Text:        "help",
			Description: "show the help for the current menu",
		},
	}
	ExitCommand = &MenuCompleter{
		Suggestion: &prompt.Suggest{
			Text:        "exit",
			Description: "exit the interactive console",
		},
	}
	TailCommands = []*MenuCompleter{UpCommand, HelpCommand, ExitCommand}
)

type MenuCompleter struct {
	Suggestion *prompt.Suggest
	Sub        []*MenuCompleter
	Parent     *MenuCompleter
}

func NewRootNode(entries []*MenuCompleter) *MenuCompleter {
	r := &MenuCompleter{Sub: append(make([]*MenuCompleter, 0, len(entries)+3), entries...)}
	for _, i := range r.Sub {
		i.Parent = r
	}
	r.Sub = append(r.Sub, newConfigMenu(r), HelpCommand, ExitCommand)
	return r
}

func NewMenuCompleter(cmd *cobra.Command, parent *MenuCompleter) *MenuCompleter {
	name := strings.SplitN(cmd.Use, " ", 2)[0]
	cmds := cmd.Commands()
	r := &MenuCompleter{
		Suggestion: &prompt.Suggest{
			Text:        name,
			Description: cmd.Short,
		},
		Parent: parent,
		Sub:    make([]*MenuCompleter, 0, len(cmds)+3),
	}
	for _, i := range cmd.Commands() {
		r.Sub = append(r.Sub, NewMenuCompleter(i, r))
	}
	r.Sub = append(r.Sub, TailCommands...)
	return r
}

const nameSep = "/"

func (cc *MenuCompleter) Name() string {
	c := cc
	parts := make([]string, 0, 8)
	for {
		if c == nil {
			break
		}
		if c.Suggestion != nil {
			parts = append(parts, c.Suggestion.Text)
		}
		c = c.Parent
	}
	sz := len(parts)
	for i := 0; i < sz/2; i++ {
		parts[i], parts[sz-i-1] = parts[sz-i-1], parts[i]
	}
	return strings.Join(parts, nameSep)
}

func (cc *MenuCompleter) Prompt(p string) string {
	return cc.Name() + p + " "
}

func (cc *MenuCompleter) Completer(doc prompt.Document) []prompt.Suggest {
	r := make([]prompt.Suggest, 0, len(cc.Sub))
	for _, i := range cc.Sub {
		r = append(r, *i.Suggestion)
	}
	return prompt.FilterHasPrefix(r, doc.GetWordBeforeCursor(), false)
}

func newConfigMenu(parent *MenuCompleter) *MenuCompleter {
	r := &MenuCompleter{
		Parent: parent,
		Suggestion: &prompt.Suggest{
			Text:        "config",
			Description: "configuration",
		},
	}
	configClientsCommand := &MenuCompleter{
		Parent: r,
		Suggestion: &prompt.Suggest{
			Text:        "clients",
			Description: "configure cryptocurrency clients",
		},
	}
	cc := make([]string, 0, len(cryptos.Cryptos))
	for c := range cryptos.Cryptos {
		cc = append(cc, c)
	}
	sort.Strings(cc)
	for _, c := range cc {
		configClientsCommand.Sub = append(configClientsCommand.Sub, newClientConfigMenu(c, configClientsCommand))
	}
	configClientsCommand.Sub = append(configClientsCommand.Sub, TailCommands...)
	r.Sub = append(r.Sub,
		configClientsCommand,
		&MenuCompleter{
			Parent: r,
			Suggestion: &prompt.Suggest{
				Text:        "save",
				Description: "save configuration",
			},
		},
		&MenuCompleter{
			Parent: r,
			Suggestion: &prompt.Suggest{
				Text:        "load",
				Description: "load configuration",
			},
		},
		&MenuCompleter{
			Parent: r,
			Suggestion: &prompt.Suggest{
				Text:        "show",
				Description: "show the current configuration",
			},
		},
	)
	r.Sub = append(r.Sub, TailCommands...)
	return r
}

func newClientConfigMenu(name string, parent *MenuCompleter) *MenuCompleter {
	r := &MenuCompleter{
		Parent: parent,
		Suggestion: &prompt.Suggest{
			Text:        name,
			Description: fmt.Sprintf("configure %s client", name),
		},
	}
	r.Sub = append(r.Sub,
		&MenuCompleter{
			Parent: r,
			Suggestion: &prompt.Suggest{
				Text:        "address",
				Description: "configure the address",
			},
		},
		&MenuCompleter{
			Parent: r,
			Suggestion: &prompt.Suggest{
				Text:        "username",
				Description: "configure the username",
			},
		},
		&MenuCompleter{
			Parent: r,
			Suggestion: &prompt.Suggest{
				Text:        "password",
				Description: "configure the password",
			},
		},
		newClientTLSConfigMenu(r),
		&MenuCompleter{
			Parent: r,
			Suggestion: &prompt.Suggest{
				Text:        "show",
				Description: "show configuration",
			},
		},
	)
	r.Sub = append(r.Sub, TailCommands...)
	return r
}

func newClientTLSConfigMenu(parent *MenuCompleter) *MenuCompleter {
	r := &MenuCompleter{
		Parent: parent,
		Suggestion: &prompt.Suggest{
			Text:        "tls",
			Description: "configure tls",
		},
	}
	r.Sub = append(r.Sub,
		&MenuCompleter{
			Parent: r,
			Suggestion: &prompt.Suggest{
				Text:        "ca",
				Description: "configure CA certificate",
			},
		},
		&MenuCompleter{
			Parent: r,
			Suggestion: &prompt.Suggest{
				Text:        "cert",
				Description: "configure client certificate",
			},
		},
		&MenuCompleter{
			Parent: r,
			Suggestion: &prompt.Suggest{
				Text:        "key",
				Description: "configure client key",
			},
		},
		&MenuCompleter{
			Parent: r,
			Suggestion: &prompt.Suggest{
				Text:        "skipverify",
				Description: "configure TLS to skip certificate verification",
			},
		},
		&MenuCompleter{
			Parent: r,
			Suggestion: &prompt.Suggest{
				Text:        "show",
				Description: "show configuration",
			},
		},
	)
	r.Sub = append(r.Sub, TailCommands...)
	return r
}

func InputText(pr string) string {
	return prompt.Input(
		fmt.Sprintf("%s: ", pr),
		func(prompt.Document) []prompt.Suggest { return nil },
	)
}

func InputTextWithDefault(pr, def string) string {
	r := InputText(fmt.Sprintf("%s (%s)", pr, def))
	if strings.TrimSpace(r) == "" {
		return def
	}
	return r
}

func InputMultiChoice(pr string, def string, choices []prompt.Suggest, helpFunc func(c []prompt.Suggest)) (string, bool) {
	choices = append(choices, *TailCommands[0].Suggestion, *TailCommands[1].Suggestion)
	for {
		input := prompt.Input(fmt.Sprintf("%s (%s): ", pr, def), func(doc prompt.Document) []prompt.Suggest {
			return prompt.FilterHasPrefix(choices, doc.GetWordBeforeCursor(), false)
		})
		switch ii := strings.TrimSpace(input); ii {
		case "":
			return def, true
		case "..":
			fmt.Println("aborted")
			return "", false
		case "help":
			helpFunc(choices[:len(choices)-2])
		default:
			for _, i := range choices[:len(choices)-2] {
				if input == i.Text {
					return input, true
				}
			}
			fmt.Printf("invalid choice: %s\n", input)
		}
	}
}

var yesNoMap = map[string]bool{"yes": true, "no": false}

func InputYesNo(pr string, def bool) (bool, bool) {
	choices := make([]prompt.Suggest, 0, 2)
	for _, i := range []string{"no", "yes"} {
		choices = append(choices, prompt.Suggest{Text: i})
	}
	var d string
	if def {
		d = choices[1].Text
	} else {
		d = choices[0].Text
	}
	r, ok := InputMultiChoice(pr, d, choices, func(_ []prompt.Suggest) {
		fmt.Printf("\nchoose yes or no\n")
	})
	if !ok {
		return false, false
	}
	return yesNoMap[r], true
}

func inputPath(pr string, rootPath string, mustExist bool, pathToSuggestionFn func(path string, text string) (prompt.Suggest, bool)) (string, error) {
	for {
		input := prompt.Input(pr, func(doc prompt.Document) []prompt.Suggest {
			r := make([]prompt.Suggest, 0, 0)
			text := doc.TextBeforeCursor()
			var fullPath string
			if filepath.IsAbs(text) {
				fullPath = text
			} else {
				fullPath = filepath.Join(rootPath, text)
			}
			var dirName string
			if text == "" {
				dirName = rootPath
			} else if filepath.IsAbs(text) {
				dirName, _ = filepath.Split(text)
			} else {
				if strings.HasSuffix(text, cmdutil.PathSep) {
					dirName = fullPath
				} else {
					dirName, _ = filepath.Split(fullPath)
				}
			}
			err := listPath(dirName, func(path string) {
				sug, ok := pathToSuggestionFn(path, text)
				if !ok {
					return
				}
				r = append(r, sug)
			})
			if err != nil {
				return nil
			}
			return r
		})
		if strings.TrimSpace(input) == "" {
			return "", nil
		}
		var r string
		if filepath.IsAbs(input) {
			r = input
		} else {
			r = filepath.Join(rootPath, input)
		}
		if mustExist {
			info, err := os.Stat(r)
			if err != nil {
				return "", err
			}
			if info.IsDir() {
				return "", errors.New("not a file")
			}
		}
		return r, nil
	}
}

func listPath(rootPath string, entryFn func(string)) error {
	return filepath.Walk(rootPath, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			e, ok := err.(*os.PathError)
			if !ok {
				return err
			}
			if e.Err != syscall.ENOENT && e.Err != syscall.EACCES {
				return err
			}
			return nil
		}
		if info.IsDir() {
			if filepath.Clean(rootPath) == filepath.Clean(path) {
				return nil
			}
		}
		entryFn(path)
		if info.IsDir() {
			return filepath.SkipDir
		}
		return nil
	})
}

func newSandboxedPathFilter(rootPath string) func(string, string) (prompt.Suggest, bool) {
	return func(path string, text string) (prompt.Suggest, bool) {
		cRootPath := filepath.Clean(rootPath)
		cPath := filepath.Clean(path)
		if !strings.HasPrefix(cPath, cRootPath) {
			return prompt.Suggest{}, false
		}
		p := cmdutil.TrimPath(cPath, cRootPath)
		if !strings.HasPrefix(p, text) {
			return prompt.Suggest{}, false
		}
		return prompt.Suggest{Text: p}, true
	}
}

func InputSandboxedFilename(pr string, rootPath string, mustExist bool) (string, error) {
	return inputPath(pr, rootPath, mustExist, newSandboxedPathFilter(rootPath))
}

func newAbsolutePathFilter(rootPath string) func(string, string) (prompt.Suggest, bool) {
	return func(path string, text string) (prompt.Suggest, bool) {
		if filepath.IsAbs(text) && strings.HasPrefix(path, text) {
			return prompt.Suggest{Text: path}, true
		} else {
			if strings.HasPrefix(filepath.Clean(text), "..") {
				_, pathName := filepath.Split(path)
				textDir, textName := filepath.Split(text)
				if strings.HasPrefix(pathName, textName) {
					return prompt.Suggest{Text: textDir + pathName}, true
				}
			} else {
				if relPath := cmdutil.TrimPath(path, rootPath); strings.HasPrefix(relPath, text) {
					return prompt.Suggest{Text: relPath}, true
				}
			}
		}
		return prompt.Suggest{}, false
	}
}

func InputFilename(pr string, rootPath string, mustExit bool) (string, error) {
	return inputPath(pr, rootPath, mustExit, newAbsolutePathFilter(rootPath))
}

func InputIntWithDefault(pr string, def int) (int, bool) {
	for {
		v := InputText(fmt.Sprintf("%s (%v)", pr, def))
		if v == "" {
			return def, true
		} else if v == ".." {
			fmt.Println("aborted")
			return 0, false
		}
		i, err := strconv.Atoi(v)
		if err != nil {
			fmt.Printf("%#v is not a number: %s\n", v, err)
			continue
		}
		return i, true
	}
}

func InputAmount(pr string) (types.Amount, bool) {
	for {
		if v := InputText(pr); v != "" {
			if v == "" {
				fmt.Println("aborted")
				return "", false
			}
			r := types.Amount(v)
			if !r.Valid() {
				fmt.Printf("invalid amount\n")
				continue
			}
			return r, true
		}
	}
}
