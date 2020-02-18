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

func (s Stage) String() string { return _stages[s] }

func (s *Stage) Set(st string) error {
	ns, ok := _stageNames[st]
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
	Initialized Stage = iota
	GenerateSecrets
	SendPublicKey
	ReceivePublicKey
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
	SellerStages = map[Stage]Stage{
		Initialized:         GenerateSecrets,
		GenerateSecrets:     ReceivePublicKey,
		ReceivePublicKey:    SendPublicKey,
		SendPublicKey:       GenerateLockScript,
		GenerateLockScript:  SendLockScript,
		SendLockScript:      ReceiveLockScript,
		ReceiveLockScript:   LockFunds,
		LockFunds:           WaitLockTransaction,
		WaitLockTransaction: RedeemFunds,
		RedeemFunds:         Done,
	}
	BuyerStages = map[Stage]Stage{
		Initialized:           GenerateSecrets,
		GenerateSecrets:       SendPublicKey,
		SendPublicKey:         ReceivePublicKey,
		ReceivePublicKey:      ReceiveLockScript,
		ReceiveLockScript:     GenerateLockScript,
		GenerateLockScript:    SendLockScript,
		SendLockScript:        WaitLockTransaction,
		WaitLockTransaction:   LockFunds,
		LockFunds:             WaitRedeemTransaction,
		WaitRedeemTransaction: RedeemFunds,
		RedeemFunds:           Done,
	}
	_stages = map[Stage]string{
		GenerateSecrets:       "generate",
		SendPublicKey:         "send-key",
		ReceivePublicKey:      "receive-key",
		ReceiveLockScript:     "receive-lock",
		GenerateLockScript:    "generate-lock",
		SendLockScript:        "send-lock",
		WaitLockTransaction:   "wait-locked-funds",
		LockFunds:             "lock-funds",
		WaitRedeemTransaction: "wait-redeem-funds",
		RedeemFunds:           "redeem",
	}
	_stageNames map[string]Stage
)

func init() {
	_stageNames = make(map[string]Stage, len(_stages))
	for k, v := range _stages {
		_stageNames[v] = k
	}
}
