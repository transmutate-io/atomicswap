package script

func NewEngineBCH() *Engine { return newEngine(NewGeneratorBCH()) }

type generatorBCH struct{ generatorBTC }

func NewGeneratorBCH() Generator { return &generatorBCH{generatorBTC: generatorBTC{}} }

type disassemblerBCH struct{ disassemblerBTC }

func NewDisassemblerBCH() Disassembler { return &disassemblerBCH{disassemblerBTC: disassemblerBTC{}} }

type intParserBCH struct{ intParserBTC }

func NewIntParserBCH() IntParser { return &intParserBCH{intParserBTC: intParserBTC{}} }
