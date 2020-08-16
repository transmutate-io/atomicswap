package cmds

import (
	"fmt"
	"strings"

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
		rootNode.sub = append(rootNode.sub, cobraCommandToCompleter(i, rootNode))
	}
	rootNode.sub = append(rootNode.sub,
		configCommand,
		helpCommand,
		exitCommand,
	)
	return rootNode
}

func cobraCommandToCompleter(cmd *cobra.Command, parent *commandCompleter) *commandCompleter {
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
		r.sub = append(r.sub, cobraCommandToCompleter(i, r))
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

type consoleConfig struct {
}
