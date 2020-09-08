package cmds

import (
	"bytes"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
	"syscall"
	"text/template"
	"time"

	"github.com/c-bata/go-prompt"
	"github.com/spf13/cobra"
	"github.com/transmutate-io/atomicswap/cryptos"
	"github.com/transmutate-io/atomicswap/internal/cmdutil"
	"github.com/transmutate-io/atomicswap/internal/flagutil"
	"github.com/transmutate-io/atomicswap/internal/flagutil/exitcodes"
	"github.com/transmutate-io/atomicswap/internal/uiutil"
	"github.com/transmutate-io/atomicswap/trade"
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

func init() {
	network := &_network
	flagutil.AddFlags(flagutil.FlagFuncMap{
		InteractiveConsoleCmd.Flags(): []flagutil.FlagFunc{
			network.AddFlag,
		},
	})
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
		cmdutil.ErrorExit(exitcodes.CantLoadConfig, err)
	}
	entries := append(make([]*uiutil.MenuCompleter, 0, 8), &uiutil.MenuCompleter{
		Suggestion: &prompt.Suggest{
			Text:        "cryptos",
			Description: "list available cryptos",
		},
	})
	for _, i := range []*cobra.Command{
		TradeCmd,
		ProposalCmd,
		LockSetCmd,
		WatchCmd,
		RedeemCmd,
		RecoverCmd,
	} {
		entries = append(entries, uiutil.NewMenuCompleter(i, nil))
	}
	rootNode := uiutil.NewRootNode(entries)
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
		command := prompt.Input(node.Prompt(">"), node.Completer, inputOpts)
		switch c := strings.TrimSpace(command); c {
		case uiutil.ExitCommand.Suggestion.Text:
			break Outer
		case uiutil.HelpCommand.Suggestion.Text:
			var menuName string
			if node == rootNode {
				menuName = "main"
			} else {
				menuName = node.Name()
			}
			fmt.Printf("\navailable commands (%s menu):\n", menuName)
			var maxNameLen int
			for _, i := range node.Sub {
				if sz := len(i.Suggestion.Text); sz > maxNameLen {
					maxNameLen = sz
				}
			}
			for _, i := range node.Sub {
				name := i.Suggestion.Text + strings.Repeat(" ", maxNameLen-len(i.Suggestion.Text))
				fmt.Printf("  %s    %s\n", name, i.Suggestion.Description)
			}
			fmt.Println()
		case uiutil.UpCommand.Suggestion.Text:
			if node == rootNode {
				fmt.Println("already at the top")
			} else {
				node = node.Parent
			}
		case "":
		default:
			for _, i := range node.Sub {
				if i.Suggestion != nil && i.Suggestion.Text == c {
					iaFunc, ok := interactiveActionsHandlers[i.Name()]
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
	"lockset/list":      actionListLockSets,
	"lockset/export":    actionExportLockSet,
	"lockset/accept":    actionAcceptLockSet,
	"lockset/info":      actionLockSetInfo,
	"watch/list":        actionListWatchable,
	"watch/own":         actionWatchOwn,
	"watch/trader":      actionWatchTrader,
	"watch/secret":      actionWatchSecret,
	"redeem/list":       actionListRedeemable,
	"redeem/toaddress":  actionRedeem,
	"recover/list":      actionListRecoverable,
	"recover/toaddress": actionRecover,
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
		ownCrypto    *cryptos.Crypto
		traderCrypto *cryptos.Crypto
		dur          time.Duration
	)
	tradeName, err = inputTradeName(cmd, "new trade name: ", false)
	if err != nil {
		fmt.Printf("can't create new trade: %s\n", err)
		return
	}
	if tradeName == "" {
		return
	}
	tradeName = cmdutil.TrimPath(tradeName, tradesDir(cmd))
	ownAmount, ok := uiutil.InputAmount("own amount")
	if !ok {
		return
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
		c, ok := uiutil.InputMultiChoice("own crypto", cryptosSuggestions[0].Text, cryptosSuggestions, helpFunc)
		if !ok {
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
	traderAmount, ok := uiutil.InputAmount("trader amount")
	if !ok {
		return
	}
	for i, s := range cryptosSuggestions {
		if s.Text == ownCrypto.Name {
			cryptosSuggestions = append(cryptosSuggestions[:i], cryptosSuggestions[i+1:]...)
			break
		}
	}
	for {
		c, ok := uiutil.InputMultiChoice("trader crypto", cryptosSuggestions[0].Text, cryptosSuggestions, helpFunc)
		if !ok {
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
		v := uiutil.InputText("duration")
		if v == "" {
			fmt.Println("aborted")
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
	tr, err := trade.NewOnChainTrade(ownAmount, ownCrypto, traderAmount, traderCrypto, dur)
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
		return
	}
	newName, err := inputTradeName(cmd, "new trade name: ", false)
	if err != nil {
		fmt.Printf("can't open trade: %s\n", err)
		return
	}
	if newName == "" {
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
		tn = cmdutil.TrimPath(tn, td)
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
	outFn, err := uiutil.InputFilename("output file (blank to stdout): ", ".", false)
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
	inFn, err := uiutil.InputFilename("input file: ", ".", true)
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
	TLS      *cryptocore.TLSConfig
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
		cfg.Address = uiutil.InputTextWithDefault("new address", cfg.Address)
		mainConfig[name] = cfg
	}
}

func newActionConfigClientUsername(name string) consoleCommand {
	return func(cmd *cobra.Command) {
		cfg := mainConfig.client(name)
		cfg.Username = uiutil.InputTextWithDefault("new username", cfg.Username)
		mainConfig[name] = cfg
	}
}

func newActionConfigClientPassword(name string) consoleCommand {
	return func(cmd *cobra.Command) {
		cfg := mainConfig.client(name)
		cfg.Password = uiutil.InputTextWithDefault("new password", cfg.Password)
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
		fn, err := uiutil.InputFilename("CA certificate file: ", ".", true)
		if err != nil {
			fmt.Printf("can't find file: %s\n", err)
			return
		}
		if cfg.TLS == nil {
			cfg.TLS = &cryptocore.TLSConfig{}
		}
		cfg.TLS.CA = fn
		mainConfig[name] = cfg
	}
}

func newActionConfigClientTLSCert(name string) consoleCommand {
	return func(cmd *cobra.Command) {
		cfg := mainConfig.client(name)
		fn, err := uiutil.InputFilename("Client certificate file: ", ".", true)
		if err != nil {
			fmt.Printf("can't find file: %s\n", err)
			return
		}
		if cfg.TLS == nil {
			cfg.TLS = &cryptocore.TLSConfig{}
		}
		cfg.TLS.ClientCertificate = fn
		mainConfig[name] = cfg
	}
}

func newActionConfigClientTLSKey(name string) consoleCommand {
	return func(cmd *cobra.Command) {
		cfg := mainConfig.client(name)
		fn, err := uiutil.InputFilename("Client key file: ", ".", true)
		if err != nil {
			fmt.Printf("can't find file: %s\n", err)
			return
		}
		if cfg.TLS == nil {
			cfg.TLS = &cryptocore.TLSConfig{}
		}
		cfg.TLS.ClientKey = fn
		mainConfig[name] = cfg
	}
}

func newActionConfigClientTLSSkipVerify(name string) consoleCommand {
	return func(cmd *cobra.Command) {
		cfg := mainConfig.client(name)
		skip, ok := uiutil.InputYesNo("skip certificate verification", false)
		if !ok {
			return
		}
		if cfg.TLS == nil {
			cfg.TLS = &cryptocore.TLSConfig{}
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
	name, err := uiutil.InputSandboxedFilename("file to load: ", consoleConfigDir(cmd), true)
	if err != nil {
		fmt.Printf("can't load config file: %s\n", err)
	}
	if name == "" {
		return
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
	name, err := uiutil.InputSandboxedFilename("save to file: ", consoleConfigDir(cmd), false)
	if err != nil {
		fmt.Printf("can't save config file: %s\n", err)
	}
	if name == "" {
		return
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
	tn, tr, err := openTradeFromInput(cmd, "trade to export: ")
	if err != nil {
		fmt.Printf("can't open trade: %s\n", err)
		return
	}
	if tn == "" && tr == nil {
		return
	}
	buyerTrade, err := tr.Buyer()
	if err != nil {
		fmt.Printf("can't open trade: %s\n", err)
		return
	}
	outFn, err := uiutil.InputFilename("output file (blank for stdout): ", ".", false)
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
		return
	}
	propFn, err := uiutil.InputFilename("proposal: ", ".", true)
	if err != nil {
		fmt.Printf("can't open proposal: %s\n", err)
		return
	}
	if propFn == "" {
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

func actionListLockSets(cmd *cobra.Command) {
	tpl, err := template.New("main").Parse(tradeListTemplates[len(tradeListTemplates)-1])
	if err != nil {
		fmt.Printf("can't parse template: %s\n", err)
		return
	}
	if err := listLockSets(tradesDir(cmd), os.Stdout, tpl); err != nil {
		fmt.Printf("can't list locksets: %s\n", err)
	}
}

func actionExportLockSet(cmd *cobra.Command) {
	tn, err := inputTradeName(cmd, "lockset to export: ", true)
	if err != nil {
		fmt.Printf("can't read lockset name: %s\n", err)
		return
	}
	if tn == "" {
		return
	}
	outFn, err := uiutil.InputFilename("export to: ", ".", false)
	if err != nil {
		fmt.Printf("can't read filename: %s\n", err)
		return
	}
	if outFn == "" {
		return
	}
	f, err := createFile(outFn)
	if err != nil {
		fmt.Printf("can't create output file: %s\n", err)
		return
	}
	defer f.Close()
	if err = exportLockSet(tn, f); err != nil {
		fmt.Printf("can't export lockset: %s\n", err)
		return
	}
}

func actionAcceptLockSet(cmd *cobra.Command) {
	tp, tr, err := openTradeFromInput(cmd, "trade: ")
	if err != nil {
		fmt.Printf("can't open trade: %s\n", err)
		return
	}
	if tp == "" && tr == nil {
		return
	}
	inFn, err := uiutil.InputFilename("lockset file: ", ".", true)
	if err != nil {
		fmt.Printf("can't find lockset file: %s\n", err)
		return
	}
	if inFn == "" {
		return
	}
	lsBytes, err := ioutil.ReadFile(inFn)
	if err != nil {
		fmt.Printf("can't read lockset file: %s\n", err)
		return
	}
	tpl, err := template.New("main").
		Funcs(template.FuncMap{"now": time.Now}).
		Parse(lockSetInfoTemplates[len(lockSetInfoTemplates)-1])
	if err != nil {
		fmt.Printf("can't parse template: %s\n", err)
		return
	}
	fmt.Printf("\nlockset info:\n\n")
	if err = showLockSetInfo(tp, bytes.NewReader(lsBytes), os.Stdout, tpl); err != nil {
		fmt.Printf("can't show lockset info: %s\n", err)
		return
	}
	fmt.Printf("\n")
	accepted, ok := uiutil.InputYesNo("accept locks", false)
	if !ok {
		return
	}
	if !accepted {
		fmt.Printf("not accepted\n")
		return
	}
	if err := acceptLockSet(tr, bytes.NewReader(lsBytes)); err != nil {
		fmt.Printf("can't accept trade: %s\n", err)
		return
	}
	if err = saveTrade(tp, tr); err != nil {
		fmt.Printf("can't save trade: %s\n", err)
	}
}

func actionLockSetInfo(cmd *cobra.Command) {
	tp, err := inputTradeName(cmd, "trade: ", true)
	if err != nil {
		fmt.Printf("can't find trade: %s\n", err)
		return
	}
	if tp == "" {
		return
	}
	inFn, err := uiutil.InputFilename("lockset file: ", ".", true)
	if err != nil {
		fmt.Printf("can't find lockset file: %s\n", err)
		return
	}
	if inFn == "" {
		return
	}
	fin, err := os.Open(inFn)
	if err != nil {
		fmt.Printf("can't open lockset file: %s\n", err)
		return
	}
	defer fin.Close()
	tpl, err := newLockSetTemplate().
		Parse(lockSetInfoTemplates[len(lockSetInfoTemplates)-1])
	if err != nil {
		fmt.Printf("can't parse template: %s\n", err)
		return
	}
	if err = showLockSetInfo(tp, fin, os.Stdout, tpl); err != nil {
		fmt.Printf("can't show lockset info: %s\n", err)
	}
}

func actionListWatchable(cmd *cobra.Command) {
	tpl, err := template.New("main").Parse(watchableTradesTemplates[len(watchableTradesTemplates)-1])
	if err != nil {
		fmt.Printf("can't parse template: %s\n", err)
		return
	}
	if err := listWatchable(tradesDir(cmd), os.Stdout, tpl); err != nil {
		fmt.Printf("can't list watchable trades: %s\n", err)
	}
}

func tradePathToWatchData(cmd *cobra.Command, tp string) string {
	return watchDataPath(cmd, cmdutil.TrimPath(tp, tradesDir(cmd)))
}

func actionWatch(
	cmd *cobra.Command,
	selectCryptoInfo func(trade.Trade) *trade.TraderInfo,
	selectWatchData func(*watchData) *blockWatchData,
	selectFunds func(trade.Trade) trade.FundsData,
) error {
	tn, tr, err := openTradeFromInput(cmd, "trade to watch: ")
	if err != nil {
		return err
	}
	if tn == "" && tr == nil {
		return nil
	}
	wd, err := openWatchData(tradePathToWatchData(cmd, tn))
	if err != nil {
		return err
	}
	depositTpl, err := template.New("main").
		Parse(depositChunkLogTemplates[len(depositChunkLogTemplates)-1])
	if err != nil {
		return err
	}
	blockTpl, err := template.New("main").
		Parse(blockInspectionTemplates[len(blockInspectionTemplates)-1])
	if err != nil {
		return err
	}
	cryptoInfo := selectCryptoInfo(tr)
	clientCfg := mainConfig.client(cryptoInfo.Crypto.Name)
	cl, err := newClient(
		cryptoInfo.Crypto,
		clientCfg.Address,
		clientCfg.Username,
		clientCfg.Password,
		clientCfg.TLS,
	)
	if err != nil {
		return err
	}
	firstBlock, ok := uiutil.InputIntWithDefault("lower height", 1)
	if !ok {
		return nil
	}
	confirmations, ok := uiutil.InputIntWithDefault("confirmations", 1)
	if !ok {
		return nil
	}
	return watchDeposit(
		tr,
		wd,
		os.Stdout,
		depositTpl,
		blockTpl,
		cl,
		uint64(firstBlock),
		false,
		uint64(confirmations),
		cryptoInfo,
		selectWatchData(wd),
		selectFunds(tr),
		func(tr trade.Trade) {
			if err := saveTrade(tn, tr); err != nil {
				fmt.Printf("error saving trade: %s\n", err)
			}
		},
		func(wd *watchData) {
			if err := saveWatchData(tradePathToWatchData(cmd, tn), wd); err != nil {
				fmt.Printf("error saving watch data: %s\n", err)
			}
		},
	)
}

func actionWatchOwn(cmd *cobra.Command) {
	err := actionWatch(
		cmd,
		func(tr trade.Trade) *trade.TraderInfo { return tr.OwnInfo() },
		func(wd *watchData) *blockWatchData { return wd.Own },
		func(tr trade.Trade) trade.FundsData { return tr.RecoverableFunds() },
	)
	if err != nil {
		fmt.Printf("can't watch deposit: %s\n", err)
	}
}

func actionWatchTrader(cmd *cobra.Command) {
	err := actionWatch(
		cmd,
		func(tr trade.Trade) *trade.TraderInfo { return tr.TraderInfo() },
		func(wd *watchData) *blockWatchData { return wd.Trader },
		func(tr trade.Trade) trade.FundsData { return tr.RedeemableFunds() },
	)
	if err != nil {
		fmt.Printf("can't watch deposit: %s\n", err)
	}
}

func actionWatchSecret(cmd *cobra.Command) {
	tn, tr, err := openTradeFromInput(cmd, "trade to watch: ")
	if err != nil {
		fmt.Printf("can't open trade: %s\n", err)
		return
	}
	if tn == "" && tr == nil {
		return
	}
	wd, err := openWatchData(tradePathToWatchData(cmd, tn))
	if err != nil {
		fmt.Printf("can't open watch data: %s\n", err)
		return
	}
	clientCfg := mainConfig.client(tr.OwnInfo().Crypto.Name)
	cl, err := newClient(
		tr.OwnInfo().Crypto,
		clientCfg.Address,
		clientCfg.Username,
		clientCfg.Password,
		clientCfg.TLS,
	)
	if err != nil {
		fmt.Printf("can't create client: %s\n", err)
		return
	}
	firstBlock, ok := uiutil.InputIntWithDefault("lower height", 1)
	if !ok {
		return
	}
	blockTpl, err := template.New("main").
		Parse(blockInspectionTemplates[len(blockInspectionTemplates)-1])
	if err != nil {
		fmt.Printf("can't parse template: %s\n", err)
		return
	}
	foundTpl, err := template.New("main").Parse("found token: {{ .Hex }}\n")
	if err != nil {
		fmt.Printf("can't parse template: %s\n", err)
		return
	}
	err = watchSecretToken(tr, wd, cl, uint64(firstBlock), os.Stdout, blockTpl, foundTpl)
	if err != nil {
		fmt.Printf("error watching for the secret token: %s\n", err)
		return
	}
	if err := saveTrade(tn, tr); err != nil {
		fmt.Printf("can't save trade: %s\n", err)
	}
}

func actionListRedeemable(cmd *cobra.Command) {
	tpl, err := template.New("main").Parse(tradeListTemplates[len(tradeListTemplates)-1])
	if err != nil {
		fmt.Printf("can't parse template: %s\n", err)
		return
	}
	fmt.Printf("\nredeemable trades:\n\n")
	if err := listRedeemable(tradesDir(cmd), os.Stdout, tpl); err != nil {
		fmt.Printf("can't list redeemable trades: %s\n", err)
		return
	}
	fmt.Println()
}

func inputRedeemRecoverData(cmd *cobra.Command, pr string) (trade.Trade, string, uint64, bool, bool) {
	tn, tr, err := openTradeFromInput(cmd, pr)
	if err != nil {
		fmt.Printf("can't open trade: %s\n", err)
		return nil, "", 0, false, false
	}
	if tn == "" && tr == nil {
		return nil, "", 0, false, false
	}
	destAddr := uiutil.InputText(fmt.Sprintf("destination address (%s)", tr.TraderInfo().Crypto.Name))
	if destAddr == "" {
		fmt.Println("aborted")
		return nil, "", 0, false, false
	}
	choices := []prompt.Suggest{
		{Text: "byte", Description: "per byte fee"},
		{Text: "fixed", Description: "fixed fee"},
	}
	ft, ok := uiutil.InputMultiChoice("fee type", choices[0].Text, choices, func(c []prompt.Suggest) {
		fmt.Printf("\nfee types:\n\n")
		for _, i := range c {
			fmt.Printf("  %s - %s\n", i.Text, i.Description)
		}
		fmt.Printf("\n  .. abort\n\n")
	})
	if !ok {
		return nil, "", 0, false, false
	}
	var intPr string
	if ft == "fixed" {
		intPr = "fee"
	} else {
		intPr = "fee per byte"
	}
	fee, ok := uiutil.InputIntWithDefault(intPr, 1)
	if !ok {
		return nil, "", 0, false, false
	}
	var fixedFee bool
	if ft == "fixed" {
		fixedFee = true
	}
	return tr, destAddr, uint64(fee), fixedFee, true
}

func actionRedeem(cmd *cobra.Command) {
	tr, destAddr, fee, fixedFee, ok := inputRedeemRecoverData(cmd, "trade to redeem: ")
	if !ok {
		return
	}
	cfg := mainConfig.client(tr.TraderInfo().Crypto.Name)
	err := redeemToAddress(
		tr,
		os.Stdout,
		destAddr,
		cfg.Address,
		cfg.Username,
		cfg.Password,
		cfg.TLS,
		uint64(fee),
		fixedFee,
		true,
	)
	if err != nil {
		fmt.Printf("can't redeem funds: %s\n", err)
	}
}

func actionListRecoverable(cmd *cobra.Command) {
	tpl, err := template.New("main").Parse(tradeListTemplates[len(tradeListTemplates)-1])
	if err != nil {
		fmt.Printf("can't parse template: %s\n", err)
		return
	}
	fmt.Printf("\nrecoverable trades:\n\n")
	if err := listRecoverable(tradesDir(cmd), os.Stdout, tpl); err != nil {
		fmt.Printf("can't list recoverable trades: %s\n", err)
		return
	}
	fmt.Println()
}

func actionRecover(cmd *cobra.Command) {
	tr, destAddr, fee, fixedFee, ok := inputRedeemRecoverData(cmd, "trade to recover: ")
	if !ok {
		return
	}
	cfg := mainConfig.client(tr.TraderInfo().Crypto.Name)
	cl, err := newClient(
		tr.OwnInfo().Crypto,
		cfg.Address,
		cfg.Username,
		cfg.Password,
		cfg.TLS,
	)
	if err != nil {
		fmt.Printf("can't create client: %s\n", err)
		return
	}
	err = recoverFunds(
		tr,
		cl,
		tr.OwnInfo(),
		destAddr,
		fixedFee,
		fee,
		os.Stdout,
		true,
	)
	if err != nil {
		fmt.Printf("can't recover funds: %s\n", err)
	}
}
