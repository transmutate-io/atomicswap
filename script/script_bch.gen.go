package script

// NewEngineBCH returns a new *Engine for bitcoin-cash
func NewEngineBCH() *Engine { return newEngine(NewGeneratorBCH()) }

type generatorBCH struct{ generatorBTC }

// NewGeneratorBCH returns a new bitcoin-cash generator
func NewGeneratorBCH() Generator { return &generatorBCH{generatorBTC: generatorBTC{}} }

type disassemblerBCH struct{ disassemblerBTC }

// NewDisassemblerBCH returns a new Disassembler for bitcoin-cash
func NewDisassemblerBCH() Disassembler { return &disassemblerBCH{disassemblerBTC: disassemblerBTC{}} }

type intParserBCH struct{ intParserBTC }

// NewIntParserBCH returns a new int64 parser for bitcoin-cash
func NewIntParserBCH() IntParser { return &intParserBCH{intParserBTC: intParserBTC{}} }
