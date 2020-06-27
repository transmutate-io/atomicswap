package {{ .Values.package }}

func NewEngine{{ .Values.short }}() *Engine { return newEngine(NewGenerator{{ .Values.short }}()) }

type generator{{ .Values.short }} struct{ generatorBTC }

func NewGenerator{{ .Values.short }}() Generator { return &generator{{ .Values.short }}{generatorBTC: generatorBTC{}} }

type disassembler{{ .Values.short }} struct{ disassemblerBTC }

func NewDisassembler{{ .Values.short }}() Disassembler { return &disassembler{{ .Values.short }}{disassemblerBTC: disassemblerBTC{}} }

type intParser{{ .Values.short }} struct{ intParserBTC }

func NewIntParser{{ .Values.short }}() IntParser { return &intParser{{ .Values.short }}{intParserBTC: intParserBTC{}} }

