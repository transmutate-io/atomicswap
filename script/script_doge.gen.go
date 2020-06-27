package script

func NewEngineDOGE() *Engine { return newEngine(NewGeneratorDOGE()) }

type generatorDOGE struct{ generatorBTC }

func NewGeneratorDOGE() Generator { return &generatorDOGE{generatorBTC: generatorBTC{}} }

type disassemblerDOGE struct{ disassemblerBTC }

func NewDisassemblerDOGE() Disassembler { return &disassemblerDOGE{disassemblerBTC: disassemblerBTC{}} }

type intParserDOGE struct{ intParserBTC }

func NewIntParserDOGE() IntParser { return &intParserDOGE{intParserBTC: intParserBTC{}} }
