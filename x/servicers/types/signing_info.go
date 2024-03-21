package types

// Signing information of the validator is needed for tracking bad acting within the block signing process
//type ValidatorSigningInfo struct {
//	Address             sdk.Address `json:"address" yaml:"address"`                             // validators address
//	StartHeight         int64       `json:"start_height" yaml:"start_height"`                   // height at which validator was first a candidate
//	Index               int64       `json:"index_offset" yaml:"index_offset"`                   // index offset into signed block bit array
//	JailedUntil         time.Time   `json:"jailed_until" yaml:"jailed_until"`                   // timestamp validator cannot be unjailed until
//	MissedBlocksCounter int64       `json:"missed_blocks_counter" yaml:"missed_blocks_counter"` // missed blocks counter (to avoid scanning the array every time)
//	JailedBlocksCounter int64       `json:"jailed_blocks_counter" yaml:"jailed_blocks_counter"` // jailed blocks counter (to avoid scanning the array every time)
//}

func (i *ValidatorSigningInfo) ResetSigningInfo() {
	i.JailedBlocksCounter = 0
	i.MissedBlocksCounter = 0
	i.Index = 0
}

func (i *ValidatorSigningInfo) ResetMissedReportCard() {
	i.MissedReportCardCounter = 0
}
