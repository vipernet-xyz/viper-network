package types

const (
	ClaimFee      = 10000 // fee for claim message (in uvipr)
	ProofFee      = 10000 // fee for proof message (in uvipr)
	ReportCardFee = 10000 // fee for report card message (in uvipr)
)

var (
	// map of message name to fee value
	ViperFeeMap = map[string]int64{
		MsgClaimName:            ClaimFee,
		MsgProofName:            ProofFee,
		MsgSubmitReportCardName: ReportCardFee,
	}
)
