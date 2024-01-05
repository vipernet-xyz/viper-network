package types

// pos module event types
const (
	EventTypeCompleteUnstaking = "complete_unstaking"
	EventTypeCreateRequestor   = "create_requestor"
	EventTypeStake             = "stake"
	EventTypeBeginUnstake      = "begin_unstake"
	EventTypeUnstake           = "unstake"
	AttributeKeyRequestor      = "requestor"
	AttributeValueCategory     = ModuleName
)
