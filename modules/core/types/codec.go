package types

import (
	codectypes "github.com/cosmos/cosmos-sdk/codec/types"

	clienttypes "github.com/vipernet-xyz/ibc-go/v7/modules/core/02-client/types"
	connectiontypes "github.com/vipernet-xyz/ibc-go/v7/modules/core/03-connection/types"
	channeltypes "github.com/vipernet-xyz/ibc-go/v7/modules/core/04-channel/types"
	commitmenttypes "github.com/vipernet-xyz/ibc-go/v7/modules/core/23-commitment/types"
)

// RegisterInterfaces registers x/ibc interfaces into protobuf Any.
func RegisterInterfaces(registry codectypes.InterfaceRegistry) {
	clienttypes.RegisterInterfaces(registry)
	connectiontypes.RegisterInterfaces(registry)
	channeltypes.RegisterInterfaces(registry)
	commitmenttypes.RegisterInterfaces(registry)
}
