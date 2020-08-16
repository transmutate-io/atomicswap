package cmds

import (
	"fmt"
	"strings"

	"github.com/c-bata/go-prompt"
	"github.com/spf13/cobra"
)

var (
	InteractiveConsoleCmd = &cobra.Command{
		Use:     "interactive",
		Short:   "start the interactive console",
		Aliases: []string{"i", "console", "c"},
		Args:    cobra.NoArgs,
		Run:     cmdInteractiveConsole,
	}
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
)

type commandCompleter struct {
	suggestion *prompt.Suggest
	sub        []*commandCompleter
	parent     *commandCompleter
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
	r.sub = append(r.sub, exitCommand, upCommand, helpCommand)
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

func cmdInteractiveConsole(cmd *cobra.Command, args []string) {
	rootNode := &commandCompleter{sub: []*commandCompleter{exitCommand}}
	for _, i := range []*cobra.Command{
		RedeemCmd,
		ProposalCmd,
		RecoverCmd,
		TradeCmd,
		LockSetCmd,
	} {
		rootNode.sub = append(rootNode.sub, newCommandCompleter(i, rootNode))
	}
	rootNode.sub = append(rootNode.sub, helpCommand)
	node := rootNode
	inputOpts := prompt.OptionAddKeyBind(
		prompt.KeyBind{
			Key: prompt.Key(prompt.ControlC),
			Fn: func(d *prompt.Buffer) {
				fmt.Println(`please terminate with the "exit" command`)
				d.Document().Text = "exit\n"
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

var interactiveActionsHandlers = map[string]func(cmd *cobra.Command){
	"trade/new": func(_ *cobra.Command) {
		fmt.Printf("new trade\n")
	},
}
