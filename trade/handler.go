package trade

import (
	"errors"
	"fmt"
	"strings"

	"github.com/transmutate-io/atomicswap/stages"
)

type (
	// StageHandlerFunc is a trade stage handler function
	StageHandlerFunc = func(trade Trade) error
	// StageHandlerMap is a map of stage handlers
	StageHandlerMap = map[stages.Stage]StageHandlerFunc
	// Handler is a trade handler
	Handler struct{ handlers StageHandlerMap }
)

// NewHandler returns a new *Handler
func NewHandler(hm StageHandlerMap) *Handler {
	r := &Handler{handlers: make(StageHandlerMap, 16)}
	if hm != nil {
		r.InstallStageHandlers(hm)
	}
	return r
}

// NewHandlerDefaults returns a new *Handler with the default handlers installed
func NewHandlerDefaults(hm StageHandlerMap) *Handler {
	r := NewHandler(DefaultStageHandlers)
	r.InstallStageHandlers(hm)
	return r
}

// InstallStageHandlers installs multiple handlers from a stage handler map
func (sh *Handler) InstallStageHandlers(hm StageHandlerMap) {
	for k, v := range hm {
		sh.handlers[k] = v
	}
}

// InstallStageHandler installs a stage handler
func (sh *Handler) InstallStageHandler(s stages.Stage, h StageHandlerFunc) {
	sh.handlers[s] = h
}

// Unhandled returns the unhandler stages of the trade
func (sh *Handler) Unhandled(s ...stages.Stage) []stages.Stage {
	r := make([]stages.Stage, 0, len(s))
	for _, i := range s {
		if _, ok := sh.handlers[i]; !ok {
			r = append(r, i)
		}
	}
	return r
}

// HandleStage handles a single trade stage
func (sh *Handler) HandleStage(s stages.Stage, t Trade) error {
	h, ok := sh.handlers[s]
	if !ok {
		return StagesNotHandledError([]stages.Stage{s})
	}
	return h(t)
}

// HandleTrade handles a trade
func (sh *Handler) HandleTrade(t Trade) error {
	stager := t.Stager()
	h := sh.Unhandled(stager.Stages()...)
	if len(h) > 0 {
		return StagesNotHandledError(h)
	}
	for {
		if err := sh.HandleStage(stager.Stage(), t); err != nil {
			if err == ErrInterruptTrade {
				return nil
			}
			return err
		}
		if s := stager.NextStage(); s == stages.Done {
			return nil
		}
	}
}

// StagesNotHandledError represents an error related to unhandled stages
type StagesNotHandledError []stages.Stage

// Error implement error
func (e StagesNotHandledError) Error() string {
	s := []stages.Stage(e)
	ss := make([]string, 0, len(s))
	for _, i := range s {
		ss = append(ss, i.String())
	}
	return fmt.Sprintf("stages not handled: %s", strings.Join(ss, ", "))
}

var (
	// NoOpHandler no operation handler
	NoOpHandler = func(_ Trade) error { return nil }
	// InterruptHandler interrupt trade handler
	InterruptHandler = func(_ Trade) error { return ErrInterruptTrade }
	// DefaultStageHandlers default handlers
	DefaultStageHandlers = StageHandlerMap{
		stages.Done:         func(_ Trade) error { return nil },
		stages.GenerateKeys: func(tr Trade) error { return tr.GenerateKeys() },
		stages.GenerateToken: func(tr Trade) error {
			btr, err := tr.Buyer()
			if err != nil {
				return err
			}
			_, err = btr.GenerateToken()
			return err
		},
	}
	// ErrInterruptTrade is returned when a trade interruption happens
	ErrInterruptTrade = errors.New("trade interrupted")
)
