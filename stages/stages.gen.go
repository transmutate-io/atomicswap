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
 	GenerateKeys Stage = iota
 	WaitRedeemableFunds
 	RedeemFunds
 	ShareProposalResponse
 	ReceiveKeyData
 	ShareProposal
 	ReceiveProposalResponse
 	ReceiveLock
 	Done
 	GenerateToken
 	ReceiveProposal
 	ShareKeyData
 	GenerateLock
 	ShareLock
 	LockFunds
 	WaitLockedFunds
)

var (
	_Stage = map[Stage]string{
		GenerateKeys:            "generate-keys",
		WaitRedeemableFunds:     "wait-redeemable-funds",
		RedeemFunds:             "redeem",
		ShareProposalResponse:   "share-proposal-response",
		ReceiveKeyData:          "receive-key-data",
		ShareProposal:           "share-proposal",
		ReceiveProposalResponse: "receive-proposal-response",
		ReceiveLock:             "receive-lock",
		LockFunds:               "lock-funds",
		WaitLockedFunds:         "wait-locked-funds",
		Done:                    "done",
		GenerateToken:           "generate-token",
		ReceiveProposal:         "receive-proposal",
		ShareKeyData:            "share-key-data",
		GenerateLock:            "generate-lock",
		ShareLock:               "share-lock",
	}
	_StageNames map[string]Stage
)

func init() {
	_StageNames = make(map[string]Stage, len(_Stage))
	for k, v := range _Stage {
		_StageNames[v] = k
	}
}
