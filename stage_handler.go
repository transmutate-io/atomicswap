package atomicswap

import (
	"fmt"
	"strings"

	"transmutate.io/pkg/atomicswap/stages"
)

type (
	StageHandlerFunc = func(trade *Trade) error
	StageHandlerMap  = map[stages.Stage]StageHandlerFunc
	StageHandler     struct{ handlers StageHandlerMap }
)

func NewStageHandler(hm StageHandlerMap) *StageHandler {
	r := &StageHandler{handlers: make(StageHandlerMap, 16)}
	r.InstallHandlers(hm)
	return r
}

func NewStageHandlerDefaults(hm StageHandlerMap) *StageHandler {
	r := NewStageHandler(DefaultStageHandlers)
	r.InstallHandlers(hm)
	return r
}

func (sh *StageHandler) InstallHandlers(hm StageHandlerMap) {
	for k, v := range hm {
		sh.handlers[k] = v
	}
}

func (sh *StageHandler) InstallHandler(s stages.Stage, h StageHandlerFunc) {
	sh.handlers[s] = h
}

func (sh *StageHandler) Unhandled(s ...stages.Stage) []stages.Stage {
	r := make([]stages.Stage, 0, len(s))
	for _, i := range s {
		if _, ok := sh.handlers[i]; !ok {
			r = append(r, i)
		}
	}
	return r
}

func (sh *StageHandler) HandleStage(s stages.Stage, t *Trade) error {
	h, ok := sh.handlers[s]
	if !ok {
		return StagesNotHandlerError([]stages.Stage{s})
	}
	return h(t)
}

func (sh *StageHandler) HandleTrade(t *Trade) error {
	h := sh.Unhandled(t.Stages.Stages()...)
	if len(h) > 0 {
		return StagesNotHandlerError(h)
	}
	for {
		if err := sh.HandleStage(t.Stages.Stage(), t); err != nil {
			return err
		}
		if s := t.Stages.NextStage(); s == stages.Done {
			return nil
		}
	}
}

type StagesNotHandlerError []stages.Stage

func (e StagesNotHandlerError) Error() string {
	s := []stages.Stage(e)
	ss := make([]string, 0, len(s))
	for _, i := range s {
		ss = append(ss, i.String())
	}
	return fmt.Sprintf("stages not handled: %s", strings.Join(ss, ", "))
}

var (
	DefaultStageHandlers = StageHandlerMap{
		stages.Done:         func(_ *Trade) error { return nil },
		stages.GenerateKeys: func(tr *Trade) error { return tr.GenerateKeys() },
		stages.GenerateToken: func(tr *Trade) error {
			_, err := tr.GenerateToken()
			return err
		},
	}
)