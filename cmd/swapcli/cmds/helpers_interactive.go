package cmds

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"syscall"

	"github.com/c-bata/go-prompt"
	"github.com/spf13/cobra"
)

var (
	upCommand = &commandCompleter{
		suggestion: &prompt.Suggest{
			Text:        "..",
			Description: "move to the parent menu",
		},
	}
	helpCommand = &commandCompleter{
		suggestion: &prompt.Suggest{
			Text:        "help",
			Description: "show the help for the current menu",
		},
	}
	exitCommand = &commandCompleter{
		suggestion: &prompt.Suggest{
			Text:        "exit",
			Description: "exit the interactive console",
		},
	}
	tailCommands  = []*commandCompleter{upCommand, helpCommand, exitCommand}
	configCommand = &commandCompleter{
		suggestion: &prompt.Suggest{
			Text:        "config",
			Description: "configuration",
		},
	}
)

func newClientTLSConfigMenu(parent *commandCompleter) *commandCompleter {
	r := &commandCompleter{
		parent: parent,
		suggestion: &prompt.Suggest{
			Text:        "tls",
			Description: "configure tls",
		},
	}
	r.sub = append(r.sub,
		&commandCompleter{
			parent: r,
			suggestion: &prompt.Suggest{
				Text:        "ca",
				Description: "configure CA certificate",
			},
		},
		&commandCompleter{
			parent: r,
			suggestion: &prompt.Suggest{
				Text:        "cert",
				Description: "configure client certificate",
			},
		},
		&commandCompleter{
			parent: r,
			suggestion: &prompt.Suggest{
				Text:        "key",
				Description: "configure client key",
			},
		},
		&commandCompleter{
			parent: r,
			suggestion: &prompt.Suggest{
				Text:        "skipverify",
				Description: "configure TLS to skip certificate verification",
			},
		},
		&commandCompleter{
			parent: r,
			suggestion: &prompt.Suggest{
				Text:        "show",
				Description: "show configuration",
			},
		},
	)
	r.sub = append(r.sub, tailCommands...)
	return r
}

func newClientConfigMenu(name string, parent *commandCompleter) *commandCompleter {
	r := &commandCompleter{
		parent: parent,
		suggestion: &prompt.Suggest{
			Text:        name,
			Description: fmt.Sprintf("configure %s client", name),
		},
	}
	r.sub = append(r.sub,
		&commandCompleter{
			parent: r,
			suggestion: &prompt.Suggest{
				Text:        "address",
				Description: "configure the address",
			},
		},
		&commandCompleter{
			parent: r,
			suggestion: &prompt.Suggest{
				Text:        "username",
				Description: "configure the username",
			},
		},
		&commandCompleter{
			parent: r,
			suggestion: &prompt.Suggest{
				Text:        "password",
				Description: "configure the password",
			},
		},
		newClientTLSConfigMenu(r),
		&commandCompleter{
			parent: r,
			suggestion: &prompt.Suggest{
				Text:        "show",
				Description: "show configuration",
			},
		},
	)
	r.sub = append(r.sub, tailCommands...)
	return r
}

type commandCompleter struct {
	suggestion *prompt.Suggest
	sub        []*commandCompleter
	parent     *commandCompleter
}

func newRootNode() *commandCompleter {
	rootNode := &commandCompleter{sub: append(make([]*commandCompleter, 0, 0), &commandCompleter{
		suggestion: &prompt.Suggest{
			Text:        "cryptos",
			Description: "list available cryptos",
		},
	})}
	configCommand.parent = rootNode
	for _, i := range []*cobra.Command{
		TradeCmd,
		ProposalCmd,
		LockSetCmd,
		WatchCmd,
		RedeemCmd,
		RecoverCmd,
	} {
		rootNode.sub = append(rootNode.sub, newCommandCompleter(i, rootNode))
	}
	rootNode.sub = append(rootNode.sub,
		configCommand,
		helpCommand,
		exitCommand,
	)
	return rootNode
}

func newCommandCompleter(cmd *cobra.Command, parent *commandCompleter) *commandCompleter {
	name := strings.SplitN(cmd.Use, " ", 2)[0]
	cmds := cmd.Commands()
	r := &commandCompleter{
		suggestion: &prompt.Suggest{
			Text:        name,
			Description: cmd.Short,
		},
		parent: parent,
		sub:    make([]*commandCompleter, 0, len(cmds)+3),
	}
	for _, i := range cmd.Commands() {
		r.sub = append(r.sub, newCommandCompleter(i, r))
	}
	r.sub = append(r.sub, tailCommands...)
	return r
}

const nameSep = "/"

func (cc *commandCompleter) name() string {
	c := cc
	parts := make([]string, 0, 8)
	for {
		if c == nil {
			break
		}
		if c.suggestion != nil {
			parts = append(parts, c.suggestion.Text)
		}
		c = c.parent
	}
	sz := len(parts)
	for i := 0; i < sz/2; i++ {
		parts[i], parts[sz-i-1] = parts[sz-i-1], parts[i]
	}
	return strings.Join(parts, nameSep)
}

func (cc *commandCompleter) prompt(p string) string {
	return cc.name() + p + " "
}

func (cc *commandCompleter) completer(doc prompt.Document) []prompt.Suggest {
	r := make([]prompt.Suggest, 0, len(cc.sub))
	for _, i := range cc.sub {
		r = append(r, *i.suggestion)
	}
	return prompt.FilterHasPrefix(r, doc.GetWordBeforeCursor(), false)
}

func inputTextWithDefault(pr, def string) string {
	r := prompt.Input(
		fmt.Sprintf("%s (%s): ", pr, def),
		func(prompt.Document) []prompt.Suggest { return nil },
	)
	if strings.TrimSpace(r) == "" {
		return def
	}
	return r
}

func inputText(pr string) string {
	return prompt.Input(
		fmt.Sprintf("%s: ", pr),
		func(prompt.Document) []prompt.Suggest { return nil },
	)
}

var yesNoMap = map[string]bool{"yes": true, "no": false}

func inputYesNo(pr string, def bool) (bool, bool) {
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
	r, ok := inputMultiChoice(pr, d, choices, func(_ []prompt.Suggest) {
		fmt.Printf("\nchoose yes or no\n")
	})
	if !ok {
		return false, false
	}
	return yesNoMap[r], true
}

func inputMultiChoice(pr string, def string, choices []prompt.Suggest, helpFunc func(c []prompt.Suggest)) (string, bool) {
	choices = append(choices, *tailCommands[0].suggestion, *tailCommands[1].suggestion)
	for {
		input := prompt.Input(fmt.Sprintf("%s (%s): ", pr, def), func(doc prompt.Document) []prompt.Suggest {
			return prompt.FilterHasPrefix(choices, doc.GetWordBeforeCursor(), false)
		})
		switch ii := strings.TrimSpace(input); ii {
		case "":
			return def, true
		case "..":
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

var pathSep = string([]rune{filepath.Separator})

func trimPath(s, p string) string {
	return strings.TrimPrefix(strings.TrimPrefix(s, p), pathSep)
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
				if strings.HasSuffix(text, pathSep) {
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

func inputSandboxedFilename(pr string, rootPath string, mustExist bool) (string, error) {
	return inputPath(pr, rootPath, mustExist, newSandboxedPathFilter(rootPath))
}

func inputFilename(pr string, rootPath string, mustExit bool) (string, error) {
	return inputPath(pr, rootPath, mustExit, newAbsolutePathFilter(rootPath))
}

func newSandboxedPathFilter(rootPath string) func(string, string) (prompt.Suggest, bool) {
	return func(path string, text string) (prompt.Suggest, bool) {
		cRootPath := filepath.Clean(rootPath)
		cPath := filepath.Clean(path)
		if !strings.HasPrefix(cPath, cRootPath) {
			return prompt.Suggest{}, false
		}
		p := trimPath(cPath, cRootPath)
		if !strings.HasPrefix(p, text) {
			return prompt.Suggest{}, false
		}
		return prompt.Suggest{Text: p}, true
	}
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
				if relPath := trimPath(path, rootPath); strings.HasPrefix(relPath, text) {
					return prompt.Suggest{Text: relPath}, true
				}
			}
		}
		return prompt.Suggest{}, false
	}
}

func inputTradeName(cmd *cobra.Command, pr string, mustExist bool) (string, error) {
	return inputSandboxedFilename(pr, tradesDir(cmd), mustExist)
}
