package types

const (
	ClaimFee = 10000 // fee for claim message (in uvip)
	ProofFee = 10000 // fee for proof message (in uvip)
)

var (
	// map of message name to fee value
	ViperFeeMap = map[string]int64{
		MsgClaimName: ClaimFee,
		MsgProofName: ProofFee,
	}
)
