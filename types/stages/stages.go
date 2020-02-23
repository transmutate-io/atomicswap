package stages

import (
	"fmt"
)

type InvalidStageError string

func (e InvalidStageError) Error() string { return fmt.Sprintf("invalid stage: \"%s\"", string(e)) }

type Stage int

func ParseStage(s string) (Stage, error) {
	var r Stage
	if err := (&r).Set(s); err != nil {
		return 0, err
	}
	return r, nil
}

func (s Stage) String() string { return stages[s] }

func (s *Stage) Set(st string) error {
	ns, ok := stageNames[st]
	if !ok {
		return InvalidStageError(st)
	}
	*s = ns
	return nil
}

func (s Stage) MarshalYAML() (interface{}, error) { return s.String(), nil }

func (s *Stage) UnmarshalYAML(unmarshal func(interface{}) error) error {
	var r string
	if err := unmarshal(&r); err != nil {
		return err
	}
	return s.Set(r)
}

const (
	SendPublicKeyHash Stage = iota
	ReceivePublicKeyHash
	SendTokenHash
	ReceiveTokenHash
	ReceiveLockScript
	GenerateLockScript
	SendLockScript
	WaitLockTransaction
	LockFunds
	WaitRedeemTransaction
	RedeemFunds
	Done
)

var (
	stages = map[Stage]string{
		SendPublicKeyHash:     "send-key-hash",
		ReceivePublicKeyHash:  "receive-key-hash",
		SendTokenHash:         "send-token-hash",
		ReceiveTokenHash:      "receive-token-hash",
		ReceiveLockScript:     "receive-lock",
		GenerateLockScript:    "generate-lock",
		SendLockScript:        "send-lock",
		WaitLockTransaction:   "wait-locked-funds",
		LockFunds:             "lock-funds",
		WaitRedeemTransaction: "wait-redeem-funds",
		RedeemFunds:           "redeem",
		Done:                  "done",
	}
	stageNames map[string]Stage
)

func init() {
	stageNames = make(map[string]Stage, len(stages))
	for k, v := range stages {
		stageNames[v] = k
	}
}
