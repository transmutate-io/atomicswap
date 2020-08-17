package script

// NewEngineLTC returns a new *Engine for litecoin
func NewEngineLTC() *Engine { return newEngine(NewGeneratorLTC()) }

type generatorLTC struct{ generatorBTC }

// NewGeneratorLTC returns a new litecoin generator
func NewGeneratorLTC() Generator { return &generatorLTC{generatorBTC: generatorBTC{}} }

type disassemblerLTC struct{ disassemblerBTC }

// NewDisassemblerLTC returns a new Disassembler for litecoin
func NewDisassemblerLTC() Disassembler { return &disassemblerLTC{disassemblerBTC: disassemblerBTC{}} }

type intParserLTC struct{ intParserBTC }

// NewIntParserLTC returns a new int64 parser for litecoin
func NewIntParserLTC() IntParser { return &intParserLTC{intParserBTC: intParserBTC{}} }
