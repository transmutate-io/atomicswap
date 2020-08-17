package {{ .Values.package }}

// NewEngine{{ .Values.short }} returns a new *Engine for {{ .Values.name }}
func NewEngine{{ .Values.short }}() *Engine { return newEngine(NewGenerator{{ .Values.short }}()) }

type generator{{ .Values.short }} struct{ generatorBTC }

// NewGenerator{{ .Values.short }} returns a new {{ .Values.name }} generator
func NewGenerator{{ .Values.short }}() Generator { return &generator{{ .Values.short }}{generatorBTC: generatorBTC{}} }

type disassembler{{ .Values.short }} struct{ disassemblerBTC }

// NewDisassembler{{ .Values.short }} returns a new Disassembler for {{ .Values.name }}
func NewDisassembler{{ .Values.short }}() Disassembler { return &disassembler{{ .Values.short }}{disassemblerBTC: disassemblerBTC{}} }

type intParser{{ .Values.short }} struct{ intParserBTC }

// NewIntParser{{ .Values.short }} returns a new int64 parser for {{ .Values.name }}
func NewIntParser{{ .Values.short }}() IntParser { return &intParser{{ .Values.short }}{intParserBTC: intParserBTC{}} }

