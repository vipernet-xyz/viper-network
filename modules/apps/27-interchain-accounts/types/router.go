package types

import (
	baseapp "github.com/vipernet-xyz/viper-network/baseapp"
	sdk "github.com/vipernet-xyz/viper-network/types"
)

// MessageRouter ADR 031 request type routing
// https://github.com/vipernet-xyz/viper-network/blob/main/docs/architecture/adr-031-msg-service.md
type MessageRouter interface {
	Handler(msg sdk.Msg1) baseapp.MsgServiceHandler
}
