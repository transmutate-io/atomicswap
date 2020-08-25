package cmds

import (
	"bytes"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"syscall"
	"text/template"
	"time"

	"github.com/c-bata/go-prompt"
	"github.com/spf13/cobra"
	"github.com/transmutate-io/atomicswap/cryptos"
	"github.com/transmutate-io/atomicswap/trade"
	"github.com/transmutate-io/cryptocore"
	"github.com/transmutate-io/cryptocore/types"
	"gopkg.in/yaml.v2"
)

var InteractiveConsoleCmd = &cobra.Command{
	Use:     "interactive",
	Short:   "start the interactive console",
	Aliases: []string{"i", "console", "c"},
	Args:    cobra.NoArgs,
	Run:     cmdInteractiveConsole,
}

func init() {
	addFlagCryptoChain(InteractiveConsoleCmd.Flags())
	for c := range cryptos.Cryptos {
		interactiveActionsHandlers["config/clients/"+c+"/address"] = newActionConfigClientAddress(c)
		interactiveActionsHandlers["config/clients/"+c+"/username"] = newActionConfigClientUsername(c)
		interactiveActionsHandlers["config/clients/"+c+"/password"] = newActionConfigClientPassword(c)
		interactiveActionsHandlers["config/clients/"+c+"/show"] = newActionConfigClientShow(c)
		interactiveActionsHandlers["config/clients/"+c+"/tls/ca"] = newActionConfigClientTLSCaCert(c)
		interactiveActionsHandlers["config/clients/"+c+"/tls/cert"] = newActionConfigClientTLSCert(c)
		interactiveActionsHandlers["config/clients/"+c+"/tls/key"] = newActionConfigClientTLSKey(c)
		interactiveActionsHandlers["config/clients/"+c+"/tls/skipverify"] = newActionConfigClientTLSSkipVerify(c)
		interactiveActionsHandlers["config/clients/"+c+"/tls/show"] = newActionConfigClientTLSShow(c)
	}
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
	"cryptos":           actionListCryptos,
	"config/save":       actionSaveConfig,
	"config/load":       actionLoadConfig,
	"config/show":       actionShowConfig,
	"trade/new":         actionNewTrade,
	"trade/list":        actionListTrades,
	"trade/rename":      actionRenameTrade,
	"trade/delete":      actionDeleteTrade,
	"trade/export":      actionExportTrades,
	"trade/import":      actionImportTrades,
	"proposal/list":     actionListProposals,
	"proposal/export":   actionExportProposal,
	"proposal/accept":   actionAcceptProposal,
	"lockset/list":      nil,
	"lockset/export":    nil,
	"lockset/accept":    nil,
	"lockset/info":      nil,
	"watch/list":        nil,
	"watch/own":         nil,
	"watch/trader":      nil,
	"watch/secret":      nil,
	"redeem/list":       nil,
	"redeem/toaddress":  nil,
	"recover/list":      nil,
	"recover/toaddress": nil,
}

func init() {
	configClientsCommand := &commandCompleter{
		parent: configCommand,
		suggestion: &prompt.Suggest{
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
		configClientsCommand.sub = append(configClientsCommand.sub, newClientConfigMenu(c, configClientsCommand))
	}
	configClientsCommand.sub = append(configClientsCommand.sub, tailCommands...)
	configCommand.sub = append(configCommand.sub,
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

func actionListCryptos(cmd *cobra.Command) {
	tpl, err := template.New("main").Parse("{{ .Name }} ({{ .Short }})\n")
	if err != nil {
		fmt.Printf("can't list cryptos: %s\n", err)
		return
	}
	fmt.Printf("\navailable cryptocurrencies:\n\n")
	if err = listCryptos(os.Stdout, tpl); err != nil {
		fmt.Printf("can't list cryptos: %s\n", err)
	}
	fmt.Println()
}

func actionNewTrade(cmd *cobra.Command) {
	var (
		tradeName    string
		err          error
		ownAmount    types.Amount
		ownCrypto    *cryptos.Crypto
		traderAmount types.Amount
		traderCrypto *cryptos.Crypto
		dur          time.Duration
	)
	tradeName, err = inputTradeName(cmd, "new trade name: ", false)
	if err != nil {
		fmt.Printf("can't create new trade: %s\n", err)
		return
	}
	tradeName = trimPath(tradeName, tradesDir(cmd))
	if tradeName == "" {
		fmt.Printf("trade creation aborted\n")
		return
	}
	for {
		if v := inputText("own amount"); v != "" {
			if ownAmount = types.Amount(v); !ownAmount.Valid() {
				fmt.Printf("invalid amount\n")
				continue
			}
			break
		}
	}
	cryptosNames := sortedCryptos()
	cryptosSuggestions := make([]prompt.Suggest, 0, len(cryptosNames))
	for _, i := range cryptosNames {
		cryptosSuggestions = append(cryptosSuggestions, prompt.Suggest{Text: i})
	}
	helpFunc := func(c []prompt.Suggest) {
		fmt.Printf("\navailable cryptocurrencies:\n\n")
		for _, i := range c {
			fmt.Printf("  %s\n", i.Text)
		}
		fmt.Printf("\n  .. to abort\n\n")
	}
	for {
		c, ok := inputMultiChoice("own crypto", cryptosNames[0], cryptosSuggestions, helpFunc)
		if !ok {
			fmt.Printf("trade creation aborted\n")
			return
		}
		own, err := parseCrypto(c)
		if err != nil {
			fmt.Printf("invalid crypto: %s\n", err)
			continue
		}
		ownCrypto = own
		break
	}
	for {
		if v := inputText("trader amount"); v != "" {
			if traderAmount = types.Amount(v); !traderAmount.Valid() {
				fmt.Printf("invalid amount\n")
				continue
			}
			break
		}
	}
	for {
		c, ok := inputMultiChoice("trader crypto", cryptosNames[0], cryptosSuggestions, helpFunc)
		if !ok {
			fmt.Printf("trade creation aborted\n")
			return
		}
		trader, err := parseCrypto(c)
		if err != nil {
			fmt.Printf("invalid crypto: %s\n", err)
			continue
		}
		traderCrypto = trader
		break
	}
	for {
		v := inputText("duration")
		if v == "" {
			fmt.Printf("trade creation aborted\n")
			return
		}
		var err error
		dur, err = time.ParseDuration(v)
		if err != nil {
			fmt.Printf("invalid duration: %s\n", err)
			continue
		}
		break
	}
	tr, err := newTrade(ownAmount, ownCrypto, traderAmount, traderCrypto, dur)
	if err != nil {
		fmt.Printf("can't create a new trade: %s\n", err)
		return
	}
	if err = saveTrade(tradePath(cmd, tradeName), tr); err != nil {
		fmt.Printf("can't save trade: %s\n", err)
	}
}

func actionListTrades(cmd *cobra.Command) {
	tpl, err := template.New("main").Parse(tradeListTemplates[len(tradeListTemplates)-1])
	if err != nil {
		fmt.Printf("can't parse template: %s\n", err)
		return
	}
	fmt.Printf("\nexisting trades:\n\n")
	if err := listTrades(tradesDir(cmd), os.Stdout, tpl); err != nil {
		fmt.Printf("can't list trades: %s\n", err)
		return
	}
	fmt.Println()
}

func actionRenameTrade(cmd *cobra.Command) {
	tradeName, err := inputTradeName(cmd, "trade name: ", true)
	if err != nil {
		fmt.Printf("can't open trade: %s\n", err)
		return
	}
	if tradeName == "" {
		fmt.Printf("aborted\n")
		return
	}
	newName, err := inputTradeName(cmd, "new trade name: ", false)
	if err != nil {
		fmt.Printf("can't open trade: %s\n", err)
		return
	}
	if newName == "" {
		fmt.Printf("aborted\n")
		return
	}
	if err = renameFile(tradeName, newName); err != nil {
		fmt.Printf("can't rename trade: %s\n", err)
	}
}

func actionDeleteTrade(cmd *cobra.Command) {
	tradeName, err := inputTradeName(cmd, "trade name: ", true)
	if err != nil {
		fmt.Printf("can't open trade: %s\n", err)
		return
	}
	if tradeName == "" {
		fmt.Printf("aborted\n")
		return
	}
	if err = os.Remove(tradeName); err != nil {
		fmt.Printf("can't delete trade: %s\n", err)
	}
}

func actionExportTrades(cmd *cobra.Command) {
	tradesNames := make(map[string]struct{}, 4)
	td := tradesDir(cmd)
	for {
		if len(tradesNames) > 0 {
			fmt.Printf("\nselected trades:\n\n")
			for i := range tradesNames {
				fmt.Printf("  %s\n", i)
			}
			fmt.Println()
		}
		tn, err := inputTradeName(cmd, "add/remove trade: ", true)
		if err != nil {
			fmt.Printf("can't add/remove trade: %s\n", err)
			continue
		}
		if tn == "" {
			break
		}
		tn = trimPath(tn, td)
		if _, ok := tradesNames[tn]; ok {
			delete(tradesNames, tn)
		} else {
			tradesNames[tn] = struct{}{}
		}
	}
	if len(tradesNames) == 0 {
		fmt.Printf("no trades selected. aborting\n")
		return
	}
	trades, err := exportTrades(td, func(name string, tr trade.Trade) bool {
		_, ok := tradesNames[name]
		return ok
	})
	if err != nil {
		fmt.Printf("can't export trades: %s\n", err)
		return
	}
	outFn, err := inputFilename("output file (blank to stdout): ", ".", false)
	if err != nil {
		fmt.Printf("can't open output file: %s\n", err)
		return
	}
	var fout io.Writer
	if outFn == "" {
		fout = os.Stdout
	} else {
		f, err := createFile(outFn)
		if err != nil {
			fmt.Printf("can't open output file: %s\n", err)
			return
		}
		defer f.Close()
		fout = f
	}
	if err = yaml.NewEncoder(fout).Encode(trades); err != nil {
		fmt.Printf("can't encode trades: %s\n", err)
	}
}

func actionImportTrades(cmd *cobra.Command) {
	inFn, err := inputFilename("input file: ", ".", true)
	if err != nil {
		fmt.Printf("can't open input file: %s\n", err)
		return
	}
	f, err := os.Open(inFn)
	if err != nil {
		fmt.Printf("can't open input file: %s\n", err)
		return
	}
	defer f.Close()
	trades := make(map[string]*trade.OnChainTrade, 16)
	if err = yaml.NewDecoder(f).Decode(trades); err != nil {
		fmt.Printf("can't decode trades file: %s\n", err)
		return
	}
	td := tradesDir(cmd)
	for n, tr := range trades {
		f, err := createFile(filepath.Join(td, filepath.FromSlash(n)))
		if err != nil {
			fmt.Printf("can't create trade: %s\n", err)
			return
		}
		defer f.Close()
		if err = yaml.NewEncoder(f).Encode(tr); err != nil {
			fmt.Printf("can't encode trade: %s\n", err)
			return
		}
	}
}

type clientConfig struct {
	Address  string
	Username string
	Password string
	TLS      cryptocore.TLSConfig
}

type consoleConfig map[string]*clientConfig

func (cc consoleConfig) client(name string) *clientConfig {
	r, ok := mainConfig[name]
	if !ok {
		return &clientConfig{}
	}
	return r
}

var mainConfig = consoleConfig{}

type consoleCommand = func(*cobra.Command)

func newActionConfigClientAddress(name string) consoleCommand {
	return func(cmd *cobra.Command) {
		cfg := mainConfig.client(name)
		cfg.Address = inputTextWithDefault("new address", cfg.Address)
		mainConfig[name] = cfg
	}
}

func newActionConfigClientUsername(name string) consoleCommand {
	return func(cmd *cobra.Command) {
		cfg := mainConfig.client(name)
		cfg.Username = inputTextWithDefault("new username", cfg.Username)
		mainConfig[name] = cfg
	}
}

func newActionConfigClientPassword(name string) consoleCommand {
	return func(cmd *cobra.Command) {
		cfg := mainConfig.client(name)
		cfg.Password = inputTextWithDefault("new password", cfg.Password)
		mainConfig[name] = cfg
	}
}

func newActionConfigClientShow(name string) consoleCommand {
	return func(cmd *cobra.Command) {
		b := bytes.NewBuffer(make([]byte, 0, 1024))
		err := yaml.NewEncoder(b).Encode(mainConfig.client(name))
		if err != nil {
			fmt.Printf("can't encode config:% s\n", err)
			return
		}
		fmt.Printf("\n%s client configuration:\n\n%s\n\n", name, b.String())
	}
}

func newActionConfigClientTLSCaCert(name string) consoleCommand {
	return func(cmd *cobra.Command) {
		cfg := mainConfig.client(name)
		fn, err := inputFilename("CA certificate file: ", ".", true)
		if err != nil {
			fmt.Printf("can't find file: %s\n", err)
			return
		}
		cfg.TLS.CA = fn
		mainConfig[name] = cfg
	}
}

func newActionConfigClientTLSCert(name string) consoleCommand {
	return func(cmd *cobra.Command) {
		cfg := mainConfig.client(name)
		fn, err := inputFilename("Client certificate file: ", ".", true)
		if err != nil {
			fmt.Printf("can't find file: %s\n", err)
			return
		}
		cfg.TLS.ClientCertificate = fn
		mainConfig[name] = cfg
	}
}

func newActionConfigClientTLSKey(name string) consoleCommand {
	return func(cmd *cobra.Command) {
		cfg := mainConfig.client(name)
		fn, err := inputFilename("Client key file: ", ".", true)
		if err != nil {
			fmt.Printf("can't find file: %s\n", err)
			return
		}
		cfg.TLS.ClientKey = fn
		mainConfig[name] = cfg
	}
}

func newActionConfigClientTLSSkipVerify(name string) consoleCommand {
	return func(cmd *cobra.Command) {
		cfg := mainConfig.client(name)
		skip, ok := inputYesNo("skip certificate verification", false)
		if !ok {
			return
		}
		cfg.TLS.SkipVerify = skip
		mainConfig[name] = cfg
	}
}

func newActionConfigClientTLSShow(name string) consoleCommand {
	return func(cmd *cobra.Command) {
		cfg := mainConfig.client(name)
		if err := yaml.NewEncoder(os.Stdout).Encode(cfg.TLS); err != nil {
			fmt.Printf("can't encode tls config: %s\n", err)
		}
	}
}

func loadConfigFile(cmd *cobra.Command, name string) error {
	var def bool
	if strings.TrimSpace(name) == "" {
		name = consoleConfigPath(cmd, DEFAULT_CONSOLE_CONFIG_NAME)
		def = true
	}
	f, err := os.Open(name)
	if err != nil {
		if def {
			if e, ok := err.(*os.PathError); !ok || e.Err != syscall.ENOENT {
				return err
			}
			mainConfig = consoleConfig{}
			return nil
		}
	}
	defer f.Close()
	return yaml.NewDecoder(f).Decode(mainConfig)
}

func actionLoadConfig(cmd *cobra.Command) {
	name, err := inputSandboxedFilename("file to load: ", consoleConfigDir(cmd), true)
	if err != nil {
		fmt.Printf("can't load config file: %s\n", err)
	}
	if err := loadConfigFile(cmd, name); err != nil {
		fmt.Printf("can't load config file: %s\n", err)
	}
}

func saveConfigFile(cmd *cobra.Command, name string) error {
	if strings.TrimSpace(name) == "" {
		name = consoleConfigPath(cmd, DEFAULT_CONSOLE_CONFIG_NAME)
	}
	f, err := createFile(name)
	if err != nil {
		return err
	}
	defer f.Close()
	return yaml.NewEncoder(f).Encode(mainConfig)
}

func actionSaveConfig(cmd *cobra.Command) {
	name, err := inputSandboxedFilename("save to file: ", consoleConfigDir(cmd), false)
	if err != nil {
		fmt.Printf("can't save config file: %s\n", err)
	}
	if err := saveConfigFile(cmd, name); err != nil {
		fmt.Printf("can't save config file: %s\n", err)
	}
}

func actionShowConfig(cmd *cobra.Command) {
	b := bytes.NewBuffer(make([]byte, 0, 1024))
	if err := yaml.NewEncoder(b).Encode(mainConfig); err != nil {
		fmt.Printf("\ncan't encode main config: %s\n\n", err)
		return
	}
	fmt.Printf("\ncurrent configuration:\n\n%s\n\n", b.String())
}

func actionListProposals(cmd *cobra.Command) {
	tpl, err := template.New("main").Parse(tradeListTemplates[len(tradeListTemplates)-1])
	if err != nil {
		fmt.Printf("can't parse template: %s\n", err)
		return
	}
	if err = listProposals(tradesDir(cmd), os.Stdout, tpl); err != nil {
		fmt.Printf("can't list proposals: %s\n", err)
	}
}

func actionExportProposal(cmd *cobra.Command) {
	tn, err := inputTradeName(cmd, "trade: ", true)
	if err != nil {
		fmt.Printf("can't open trade: %s\n", err)
		return
	}
	if tn == "" {
		fmt.Printf("aborted\n")
		return
	}
	tr, err := openTrade(tn)
	if err != nil {
		fmt.Printf("can't open trade: %s\n", err)
		return
	}
	buyerTrade, err := tr.Buyer()
	if err != nil {
		fmt.Printf("can't open trade: %s\n", err)
		return
	}
	outFn, err := inputFilename("output file (blank for stdout): ", ".", false)
	if err != nil {
		fmt.Printf("can't open output file: %s\n", err)
		return
	}
	var fout io.Writer
	if outFn == "" {
		fout = os.Stdout
	} else {
		f, err := createFile(outFn)
		if err != nil {
			fmt.Printf("can't open output file: %s\n", err)
			return
		}
		defer f.Close()
		fout = f
	}
	prop, err := buyerTrade.GenerateBuyProposal()
	if err != nil {
		fmt.Printf("can't generate buy proposal: %s\n", err)
		return
	}
	if err = yaml.NewEncoder(fout).Encode(prop); err != nil {
		fmt.Printf("can't encode proposal: %s\n", err)
	}
}

func actionAcceptProposal(cmd *cobra.Command) {
	tn, err := inputTradeName(cmd, "new trade: ", false)
	if err != nil {
		fmt.Printf("can't open trade: %s\n", err)
		return
	}
	if tn == "" {
		fmt.Printf("aborted\n")
		return
	}
	propFn, err := inputFilename("proposal: ", ".", true)
	if err != nil {
		fmt.Printf("can't open proposal: %s\n", err)
		return
	}
	if propFn == "" {
		fmt.Printf("aborted\n")
		return
	}
	b, err := ioutil.ReadFile(propFn)
	if err != nil {
		fmt.Printf("can't read proposal: %s\n", err)
		return
	}
	prop, err := trade.UnamrshalBuyProposal(b)
	if err != nil {
		fmt.Printf("can't decode proposal: %s\n", err)
		return
	}
	if err = acceptProposal("", tn, prop); err != nil {
		fmt.Printf("can't accept proposal: %s\n", err)
	}
}
