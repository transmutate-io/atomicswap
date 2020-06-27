package script

func NewEngineLTC() *Engine { return newEngine(NewGeneratorLTC()) }

type generatorLTC struct{ generatorBTC }

func NewGeneratorLTC() Generator { return &generatorLTC{generatorBTC: generatorBTC{}} }

type disassemblerLTC struct{ disassemblerBTC }

func NewDisassemblerLTC() Disassembler { return &disassemblerLTC{disassemblerBTC: disassemblerBTC{}} }

type intParserLTC struct{ intParserBTC }

func NewIntParserLTC() IntParser { return &intParserLTC{intParserBTC: intParserBTC{}} }
