package exitcodes

const (
	OK = iota * -1
	BadInput
	BadOutput
	BadTemplate
	CantCreateFile
	CantGetFlag
	CantLoadConfig
	CantOpenLockSet
	CantOpenTrade
	CantOpenWatchData
	CantSaveTrade
	CantSaveWatchData
	ExecutionError
	InvalidDuration
	InvalidLockData
	NotABuyer
	UnknownCrypto
	UnknownShell

	// CantAcceptLockSet
	// CantCalculateAddress
	// CantCreateTrade
	// CantDeleteTrade
	// CantExportLockSet
	// CantExportProposal
	// CantExportTrades
	// CantImportTrades
	// CantListLockSets
	// CantListProposals
	// CantListTrades
	// CantRecover
	// CantRedeem
	// CantRenameTrade
	// CantShowLockSetInfo
	// FailedToWatch
	// OnlyOneNetwork
)

var Messages = map[int]string{
	BadInput:          "can't open input: %s\n",
	BadOutput:         "can't create output: %s\n",
	BadTemplate:       "bad template: %s\n",
	CantCreateFile:    "can't create file: %s\n",
	CantGetFlag:       "can't get flag: %s\n",
	CantLoadConfig:    "can't load config: %s\n",
	CantOpenLockSet:   "can't open lock set: %s\n",
	CantOpenTrade:     "can't open trade \"%s\": %s\n",
	CantOpenWatchData: "can't open watch data: %s\n",
	CantSaveTrade:     "can't save trade: %s\n",
	CantSaveWatchData: "can't save watch data: %s\n",
	ExecutionError:    "error: %s\n",
	InvalidDuration:   "invalid duration: \"%s\"\n",
	InvalidLockData:   "invalid lock data: %s\n",
	NotABuyer:         "not a buyer\n",
	UnknownCrypto:     "unknown crypto: \"%s\"\n",
	UnknownShell:      "can't generate completion file: %s\n",

	// CantAcceptLockSet:    "can't accept lock set: %s\n",
	// CantCalculateAddress: "can't calculate address: %s\n",
	// CantCreateTrade:      "can't create trade: %s\n",
	// CantDeleteTrade:      "can't delete trade: %s\n",
	// CantExportLockSet:    "can't export lock set: %s\n",
	// CantExportProposal:   "can't export proposal: %s\n",
	// CantExportTrades:     "can't export trades: %s\n",
	// CantImportTrades:     "can't import trades: %s\n",
	// CantListLockSets:     "can't list locksets: %s\n",
	// CantListProposals:    "can't list proposals: %s\n",
	// CantListTrades:       "can't list trades: %s\n",
	// CantRecover:          "can't recover funds: %s\n",
	// CantRedeem:           "can't redeem: %s\n",
	// CantRenameTrade:      "can't rename trade: %s\n",
	// CantShowLockSetInfo:  "can't show lockset info: %s\n",
	// FailedToWatch:        "failed to watch blockchain: %s\n",
	// OnlyOneNetwork:       "pick only one network\n",
}
