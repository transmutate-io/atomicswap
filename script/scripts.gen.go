package script

var (
	generators = map[string]Generator{
		"bitcoin-cash": NewGeneratorBCH(),
		"bitcoin":      NewGeneratorBTC(),
		"decred":       NewGeneratorDCR(),
		"dogecoin":     NewGeneratorDOGE(),
		"litecoin":     NewGeneratorLTC(),
	}

	disassemblers = map[string]Disassembler{
		"bitcoin-cash": NewDisassemblerBCH(),
		"bitcoin":      NewDisassemblerBTC(),
		"decred":       NewDisassemblerDCR(),
		"dogecoin":     NewDisassemblerDOGE(),
		"litecoin":     NewDisassemblerLTC(),
	}

	intParsers = map[string]IntParser{
		"bitcoin-cash": NewIntParserBCH(),
		"bitcoin":      NewIntParserBTC(),
		"decred":       NewIntParserDCR(),
		"dogecoin":     NewIntParserDOGE(),
		"litecoin":     NewIntParserLTC(),
	}
)
