package cmds

import (
	"bytes"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"syscall"
	"text/template"

	"github.com/c-bata/go-prompt"
	"github.com/spf13/cobra"
	"github.com/transmutate-io/atomicswap/cryptos"
	"github.com/transmutate-io/cryptocore"
	"gopkg.in/yaml.v2"
)

var InteractiveConsoleCmd = &cobra.Command{
	Use:     "interactive",
	Short:   "start the interactive console",
	Aliases: []string{"i", "console", "c"},
	Args:    cobra.NoArgs,
	Run:     cmdInteractiveConsole,
}

func cmdInteractiveConsole(cmd *cobra.Command, args []string) {
	if err := loadConfigFile(cmd, ""); err != nil {
		errorExit(ecCantLoadConfig, err)
	}
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
			fmt.Printf("\navailable commands (%s menu):\n", menuName)
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
			fmt.Println()
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
	"cryptos":        actionListCryptos,
	"config/network": actionConfigNetwork,
	"config/save":    actionSaveConfig,
	"config/load":    actionLoadConfig,
	"config/show":    actionShowConfig,
	"trade/new":      actionNewTrade,
}

func init() {
	for c := range cryptos.Cryptos {
		interactiveActionsHandlers["config/clients/"+c] = newCryptoConfigMenu(c)
	}
}

func newCryptoConfigMenu(name string) func(*cobra.Command) {
	return func(_ *cobra.Command) {
		fmt.Printf("config %s client\n", name)
	}
}

func actionListCryptos(cmd *cobra.Command) {
	tpl, err := template.New("main").Parse("{{ .Name }} ({{ .Short }})\n")
	if err != nil {
		fmt.Printf("can't list cryptos: %s\n", err)
		return
	}
	fmt.Println("available cryptocurrencies:")
	if err = listCryptos(os.Stdout, tpl); err != nil {
		fmt.Printf("can't list cryptos: %s\n", err)
	}
}

type clientConfig struct {
	Addr string
	cryptocore.TLSConfig
}

type consoleConfig struct {
	Network string
	Clients map[string]*clientConfig
}

var mainConfig = &consoleConfig{}

const DEFAULT_CONSOLE_CONFIG_NAME = "console_defaults.yaml"

func configDir(cmd *cobra.Command) string { return filepath.Join(dataDir(cmd), "config") }

func consoleConfigPath(cmd *cobra.Command, name string) string {
	if name == "" {
		name = DEFAULT_CONSOLE_CONFIG_NAME
	}
	return filepath.Join(configDir(cmd), name)
}

func loadConfigFile(cmd *cobra.Command, name string) error {
	f, err := os.Open(consoleConfigPath(cmd, name))
	if err != nil {
		e, ok := err.(*os.PathError)
		if !ok || e.Err != syscall.ENOENT {
			return err
		}
		mainConfig = &consoleConfig{Network: "mainnet"}
		return nil
	}
	defer f.Close()
	return yaml.NewDecoder(f).Decode(mainConfig)
}

func actionLoadConfig(cmd *cobra.Command) {
	name, err := inputFilename("file to load: ", configDir(cmd), true)
	if err != nil {
		fmt.Printf("can't load config file: %s\n", err)
	}
	if err := loadConfigFile(cmd, name); err != nil {
		fmt.Printf("can't load config file: %s\n", err)
	}
}

func saveConfigFile(cmd *cobra.Command, name string) error {
	f, err := createFile(consoleConfigPath(cmd, name))
	if err != nil {
		return err
	}
	defer f.Close()
	return yaml.NewEncoder(f).Encode(mainConfig)
}

func actionSaveConfig(cmd *cobra.Command) {
	if err := saveConfigFile(cmd, ""); err != nil {
		fmt.Printf("can't save config file: %s\n", err)
	}
}

func actionConfigNetwork(cmd *cobra.Command) {
	mainConfig.Network = inputNetwork()
}

func actionNewTrade(cmd *cobra.Command) {
	fmt.Printf("new trade\n")
}

func actionShowConfig(cmd *cobra.Command) {
	b := bytes.NewBuffer(make([]byte, 0, 1024))
	if err := yaml.NewEncoder(b).Encode(mainConfig); err != nil {
		fmt.Printf("\ncan't encode main config: %s\n\n", err)
		return
	}
	fmt.Printf("\ncurrent configuration:\n\n%s\n\n", b.String())
}

// func action (cmd*cobra.Command){}
