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
	"github.com/transmutate-io/atomicswap/cryptos"
)

var (
	exitCommand = &commandCompleter{
		suggestion: &prompt.Suggest{
			Text:        "exit",
			Description: "exit the interactive console",
		},
	}
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
	tailCommands  = []*commandCompleter{upCommand, helpCommand, exitCommand}
	configCommand = &commandCompleter{
		suggestion: &prompt.Suggest{
			Text:        "config",
			Description: "configuration",
		},
	}
)

func init() {
	configClientsCommand := &commandCompleter{
		parent: configCommand,
		suggestion: &prompt.Suggest{
			Text:        "clients",
			Description: "configure cryptocurrency clients",
		},
	}
	for c := range cryptos.Cryptos {
		configClientsCommand.sub = append(configClientsCommand.sub, &commandCompleter{
			parent: configClientsCommand,
			suggestion: &prompt.Suggest{
				Text:        c,
				Description: fmt.Sprintf("configure %s client", c),
			},
		})
	}
	configClientsCommand.sub = append(configClientsCommand.sub, tailCommands...)
	configCommand.sub = append(configCommand.sub,
		&commandCompleter{
			parent: configCommand,
			suggestion: &prompt.Suggest{
				Text:        "network",
				Description: "configure which network to use",
			},
		},
		configClientsCommand,
		&commandCompleter{
			parent: configCommand,
			suggestion: &prompt.Suggest{
				Text:        "save",
				Description: "save configuration",
			},
		},
		&commandCompleter{
			parent: configCommand,
			suggestion: &prompt.Suggest{
				Text:        "load",
				Description: "load configuration",
			},
		},
		&commandCompleter{
			parent: configCommand,
			suggestion: &prompt.Suggest{
				Text:        "show",
				Description: "show the current configuration",
			},
		},
	)
	configCommand.sub = append(configCommand.sub, tailCommands...)
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

func inputNetwork() string {
	choices := make([]prompt.Suggest, 0, 3)
	for _, i := range []string{"mainnet", "testnet", "locanet"} {
		choices = append(choices, prompt.Suggest{
			Text:        i,
			Description: "set the network to " + i,
		})
	}
	pr := fmt.Sprintf("new network (%s)> ", mainConfig.Network)
	r, ok := inputMultiChoice(pr, mainConfig.Network, choices, func(_ []prompt.Suggest) {
		fmt.Printf("\navailable networks:\n")
		for _, i := range choices {
			fmt.Printf("  %s\n", i.Text)
		}
		fmt.Printf("\n.. to cancel and move up\n\n")
	})
	if !ok {
		return mainConfig.Network
	}
	return r
}

func inputMultiChoice(pr string, def string, choices []prompt.Suggest, helpFunc func(c []prompt.Suggest)) (string, bool) {
	choices = append(choices, *tailCommands[0].suggestion, *tailCommands[1].suggestion)
	for {
		input := prompt.Input(pr, func(doc prompt.Document) []prompt.Suggest {
			return prompt.FilterHasPrefix(choices, doc.GetWordBeforeCursor(), false)
		})
		switch ii := strings.TrimSpace(input); ii {
		case "":
		case "..":
			return "", false
		case "help":
			helpFunc(choices[:len(choices)-1])
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

var (
	errInvalidPath = errors.New("invalid path")
	pathSep        = string([]rune{filepath.Separator})
)

func trimPath(s, p string) string {
	return strings.TrimPrefix(strings.TrimPrefix(s, p), pathSep)
}

func inputFilename(pr string, basePath string, mustExist bool) (string, error) {
	for {
		input := prompt.Input(pr, func(doc prompt.Document) []prompt.Suggest {
			r := make([]prompt.Suggest, 0, 0)
			text := doc.TextBeforeCursor()
			fullPath := filepath.Clean(filepath.Join(basePath, text))
			if strings.Contains(fullPath, "..") {
				return nil
			}
			var dirName string
			if text == "" {
				dirName = filepath.Clean(basePath)
			} else if strings.HasSuffix(text, pathSep) {
				dirName = fullPath
			} else {
				dirName, _ = filepath.Split(fullPath)
			}
			err := filepath.Walk(dirName, func(path string, info os.FileInfo, err error) error {
				if err != nil {
					e, ok := err.(*os.PathError)
					if !ok || e.Err != syscall.ENOENT {
						return err
					}
					return nil
				}
				var isDir bool
				relPath := trimPath(path, dirName)
				if info.IsDir() {
					if len(relPath) == 0 {
						return nil
					} else {
						isDir = true
					}
				}
				path = trimPath(path, basePath)
				if strings.HasPrefix(path, text) {
					var s string
					if isDir {
						s = "/"
					}
					r = append(r, prompt.Suggest{Text: fmt.Sprintf("%s%s", path, s)})
					if isDir {
						return filepath.SkipDir
					}
				}
				return nil
			})
			if err != nil {
				return nil
			}
			return r
		})
		if strings.TrimSpace(input) == "" {
			return "", nil
		}
		r := filepath.Clean(filepath.Join(basePath, input))
		if mustExist {
			if _, err := os.Stat(r); err != nil {
				return "", err
			}
		}
		return r, nil
	}
}
