package cmds

import (
	"fmt"
	"os"
	"strings"
	"text/template"

	"github.com/c-bata/go-prompt"
	"github.com/spf13/cobra"
	"github.com/transmutate-io/atomicswap/cryptos"
)

var interactiveActionsHandlers = map[string]func(cmd *cobra.Command){
	"cryptos": func(cmd *cobra.Command) {
		tpl, err := template.New("main").Parse("{{ .Name }} ({{ .Short }})\n")
		if err != nil {
			fmt.Printf("can't list cryptos: %s\n", err)
			return
		}
		fmt.Println("available cryptocurrencies:")
		if err = listCryptos(os.Stdout, tpl); err != nil {
			fmt.Printf("can't list cryptos: %s\n", err)
		}
	},
	"config/network": func(_ *cobra.Command) {
		fmt.Printf("config network\n")
	},
	"config/save": func(_ *cobra.Command) {
		fmt.Printf("save config\n")
	},
	"config/load": func(_ *cobra.Command) {
		fmt.Printf("load config\n")
	},
	"trade/new": func(_ *cobra.Command) {
		fmt.Printf("new trade\n")
	},
}

func init() {
	for c := range cryptos.Cryptos {
		interactiveActionsHandlers["config/clients/"+c] = func(_ *cobra.Command) {
			fmt.Printf("config %s client\n", c)
		}
	}
}

var InteractiveConsoleCmd = &cobra.Command{
	Use:     "interactive",
	Short:   "start the interactive console",
	Aliases: []string{"i", "console", "c"},
	Args:    cobra.NoArgs,
	Run:     cmdInteractiveConsole,
}

func cmdInteractiveConsole(cmd *cobra.Command, args []string) {
	rootNode := newRootNode()
	node := rootNode
	inputOpts := prompt.OptionAddKeyBind(
		prompt.KeyBind{
			Key: prompt.Key(prompt.ControlC),
			Fn: func(d *prompt.Buffer) {
				fmt.Println(`please terminate with the "exit" command`)
			},
		},
	)
Outer:
	for {
		command := prompt.Input(node.prompt(">"), node.completer, inputOpts)
		switch c := strings.TrimSpace(command); c {
		case exitCommand.suggestion.Text:
			break Outer
		case helpCommand.suggestion.Text:
			var menuName string
			if node == rootNode {
				menuName = "main"
			} else {
				menuName = node.name()
			}
			fmt.Printf("available commands (%s menu):\n", menuName)
			var maxNameLen int
			for _, i := range node.sub {
				if sz := len(i.suggestion.Text); sz > maxNameLen {
					maxNameLen = sz
				}
			}
			for _, i := range node.sub {
				name := i.suggestion.Text + strings.Repeat(" ", maxNameLen-len(i.suggestion.Text))
				fmt.Printf("  %s    %s\n", name, i.suggestion.Description)
			}
		case upCommand.suggestion.Text:
			if node == rootNode {
				fmt.Println("already at the top")
			} else {
				node = node.parent
			}
		case "":
		default:
			for _, i := range node.sub {
				if i.suggestion != nil && i.suggestion.Text == c {
					iaFunc, ok := interactiveActionsHandlers[i.name()]
					if ok {
						iaFunc(cmd)
					} else {
						node = i
					}
					continue Outer
				}
			}
			fmt.Printf("unknown command: %s\n", c)
		}
	}
}
