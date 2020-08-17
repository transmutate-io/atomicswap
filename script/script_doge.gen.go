package script

// NewEngineDOGE returns a new *Engine for dogecoin
func NewEngineDOGE() *Engine { return newEngine(NewGeneratorDOGE()) }

type generatorDOGE struct{ generatorBTC }

// NewGeneratorDOGE returns a new dogecoin generator
func NewGeneratorDOGE() Generator { return &generatorDOGE{generatorBTC: generatorBTC{}} }

type disassemblerDOGE struct{ disassemblerBTC }

// NewDisassemblerDOGE returns a new Disassembler for dogecoin
func NewDisassemblerDOGE() Disassembler { return &disassemblerDOGE{disassemblerBTC: disassemblerBTC{}} }

type intParserDOGE struct{ intParserBTC }

// NewIntParserDOGE returns a new int64 parser for dogecoin
func NewIntParserDOGE() IntParser { return &intParserDOGE{intParserBTC: intParserBTC{}} }
