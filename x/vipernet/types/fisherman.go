package types

import (
	"time"

	sdk "github.com/vipernet-xyz/viper-network/types"
)

type FishermanRelay struct {
	ServicerAddr sdk.Address
	Latency      time.Duration
	IsSigned     bool
}

//function to validate the relay
//function to populate the data from fishermen into relay struct and execute
