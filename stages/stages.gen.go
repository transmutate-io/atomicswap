package stages

import "fmt"

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
 	GenerateToken Stage = iota
 	SendProposal
 	ReceiveProposal
 	LockFunds
 	WaitLockedFunds
 	WaitFundsRedeemed
 	RedeemFunds
 	GenerateKeys
 	SendProposalResponse
 	ReceiveProposalResponse
 	Done
)

var (
	_Stage = map[Stage]string{
		GenerateToken:           "generate-token",
		SendProposal:            "send-proposal",
		ReceiveProposal:         "receive-proposal",
		LockFunds:               "lock-funds",
		WaitLockedFunds:         "wait-locked-funds",
		WaitFundsRedeemed:       "wait-funds-redeemed",
		RedeemFunds:             "redeem",
		GenerateKeys:            "generate-keys",
		SendProposalResponse:    "send-proposal-response",
		ReceiveProposalResponse: "receive-proposal-response",
		Done:                    "done",
	}
	_StageNames map[string]Stage
)

func init() {
	_StageNames = make(map[string]Stage, len(_Stage))
	for k, v := range _Stage {
		_StageNames[v] = k
	}
}
