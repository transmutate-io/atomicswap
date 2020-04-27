package stages

import "fmt"

type InvalidStageError string

func (e InvalidStageError) Error() string {
	return fmt.Sprintf("invalid stage: \"%s\"", string(e))
}

type Stage int

func ParseStage(s string) (Stage, error) {
	var r Stage
	if err := (&r).Set(s); err != nil {
		return 0, err
	}
	return r, nil
}

func (v Stage) String() string { return _Stage[v] }

func (v *Stage) Set(sv string) error {
	nv, ok := _StageNames[sv]
	if !ok {
		return InvalidStageError(sv)
	}
	*v = nv
	return nil
}

func (v Stage) MarshalYAML() (interface{}, error) { return v.String(), nil }

func (v *Stage) UnmarshalYAML(unmarshal func(interface{}) error) error {
	var r string
	if err := unmarshal(&r); err != nil {
		return err
	}
	return v.Set(r)
}

const (
	GenerateKeys Stage = iota
	GenerateToken
	SendProposal
	ReceiveProposal
	SendProposalResponse
	ReceiveProposalResponse
	LockFunds
	WaitLockedFunds
	WaitFundsRedeemed
	RedeemFunds
	Done
)

var (
	_Stage = map[Stage]string{
		GenerateKeys: "generate-keys",
		GenerateToken: "generate-token",
		SendProposal: "send-proposal",
		ReceiveProposal: "receive-proposal",
		SendProposalResponse: "send-proposal-response",
		ReceiveProposalResponse: "receive-proposal-response",
		LockFunds: "lock-funds",
		WaitLockedFunds: "wait-locked-funds",
		WaitFundsRedeemed: "wait-funds-redeemed",
		RedeemFunds: "redeem",
		Done: "done",
	}
	_StageNames map[string]Stage
)

func init() {
	_StageNames = make(map[string]Stage, len(_Stage))
	for k, v := range _Stage {
		_StageNames[v] = k
	}
}

