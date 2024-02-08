package types

import (
	"encoding/binary"
	"strconv"
	"time"

	sdk "github.com/vipernet-xyz/viper-network/types"
	"github.com/vipernet-xyz/viper-network/types/kv"
)

const (
	ModuleName   = "pos"                     // servicers module is called 'pos' for proof of stake
	StoreKey     = ModuleName                // StoreKey is the string store representation
	TStoreKey    = "transient_" + ModuleName // TStoreKey is the string transient store representation
	QuerierRoute = ModuleName                // QuerierRoute is the querier route for the staking module
	RouterKey    = ModuleName                // RouterKey is the msg router key for the staking module
	MemStoreKey  = "memory_" + ModuleName
)

var ( // Keys for store prefixes
	ProposerKey                     = []byte{0x01} // key for the proposer address used for rewards
	ValidatorSigningInfoKey         = []byte{0x11} // Prefix for signing info used in slashing
	ValidatorMissedBlockBitArrayKey = []byte{0x12} // Prefix for missed block bit array used in slashing
	AllValidatorsKey                = []byte{0x21} // prefix for each key to a validator
	StakedValidatorsByNetIDKey      = []byte{0x22} // prefix for validators staked by networkID
	StakedValidatorsByGeoZoneKey    = []byte{0x23} // prefix for validators staked by geoZone
	StakedValidatorsKey             = []byte{0x24} // prefix for each key to a staked validator index, sorted by power
	PrevStateValidatorsPowerKey     = []byte{0x31} // prefix for the key to the validators of the prevState state
	PrevStateTotalPowerKey          = []byte{0x32} // prefix for the total power of the prevState state
	UnstakingValidatorsKey          = []byte{0x41} // prefix for unstaking validator
	AwardValidatorKey               = []byte{0x51} // prefix for awarding validators
	BurnValidatorKey                = []byte{0x52} // prefix for burning validators
	WaitingToBeginUnstakingKey      = []byte{0x43} // prefix for waiting validators
	HistoricalInfoKey               = []byte{0x50} // prefix for the historical info
	LastValidatorPowerKey           = []byte{0x11} // prefix for each key to a validator index, for bonded validators
	ReportCardKey                   = []byte{0x70}
)

func KeyForValidatorByNetworkID(addr sdk.Address, networkID []byte) []byte {
	return append(append(StakedValidatorsByNetIDKey, networkID...), addr.Bytes()...)
}

func KeyForValidatorsByNetworkID(networkID []byte) []byte {
	return append(StakedValidatorsByNetIDKey, networkID...)
}

func KeyForValidatorByGeoZone(addr sdk.Address, geoZone []byte) []byte {
	return append(append(StakedValidatorsByGeoZoneKey, geoZone...), addr.Bytes()...)
}

func KeyForValidatorsByGeoZone(geoZone []byte) []byte {
	return append(StakedValidatorsByGeoZoneKey, geoZone...)
}

func AddressForValidatorByNetworkIDKey(key, networkID []byte) sdk.Address {
	i := len(StakedValidatorsByNetIDKey) + len(networkID)
	return key[i:]
}

func AddressForValidatorByGeozoneKey(key, geozone []byte) sdk.Address {
	i := len(StakedValidatorsByGeoZoneKey) + len(geozone)
	return key[i:]
}

func KeyForValWaitingToBeginUnstaking(addr sdk.Address) []byte {
	return append(WaitingToBeginUnstakingKey, addr.Bytes()...)
}

// generates the key for the validator with address
func KeyForValByAllVals(addr sdk.Address) []byte {
	return append(AllValidatorsKey, addr.Bytes()...)
}

// generates the key for unstaking validators by the unstakingtime
func KeyForUnstakingValidators(unstakingTime time.Time) []byte {
	bz := sdk.FormatTimeBytes(unstakingTime)
	return append(UnstakingValidatorsKey, bz...) // use the unstaking time as part of the key
}

// generates the key for a validator in the staking set
func KeyForValidatorInStakingSet(validator Validator) []byte {
	// NOTE the address doesn't need to be stored because counter bytes must always be different
	return getStakedValPowerRankKey(validator)
}

// generates the key for a validator in the prevState state
func KeyForValidatorPrevStateStateByPower(address sdk.Address) []byte {
	return append(PrevStateValidatorsPowerKey, address...)
}

// generates the award key for a validator in the current state
func KeyForValidatorAward(address sdk.Address) []byte {
	return append(AwardValidatorKey, address...)
}

func KeyForValidatorBurn(address sdk.Address) []byte {
	return append(BurnValidatorKey, address...)
}

// Removes the prefix bytes from a key to expose true address
func AddressFromKey(key []byte) []byte {
	return key[1:] // remove prefix bytes
}

// get the power ranking key of a validator
// NOTE the larger values are of higher value
func getStakedValPowerRankKey(validator Validator) []byte {
	// get the consensus power
	consensusPower := sdk.TokensToConsensusPower(validator.StakedTokens)
	consensusPowerBytes := make([]byte, 8)
	binary.BigEndian.PutUint64(consensusPowerBytes, uint64(consensusPower))

	powerBytes := consensusPowerBytes
	powerBytesLen := len(powerBytes) // 8

	// key is of format prefix || powerbytes || addrBytes
	key := make([]byte, 1+powerBytesLen+sdk.AddrLen)

	// generate the key for this validator by deriving it from the main key
	key[0] = StakedValidatorsKey[0]
	copy(key[1:powerBytesLen+1], powerBytes)
	operAddrInvr := sdk.CopyBytes(validator.Address)
	for i, b := range operAddrInvr {
		operAddrInvr[i] = ^b
	}
	copy(key[powerBytesLen+1:], operAddrInvr)

	return key
}

// generates the key for validator signing information by consensus addr
func KeyForValidatorSigningInfo(v sdk.Address) []byte {
	return append(ValidatorSigningInfoKey, v.Bytes()...)
}

// extract the address from a validator signing info key
func GetValidatorSigningInfoAddress(key []byte) (addr sdk.Address, err error) {
	addr = key[1:]
	if len(addr) != sdk.AddrLen {
		err = sdk.ErrInternal("unexpected key length for GetValidatorSigningInfoAddress")
	}
	return
}

// generates the prefix key for missing val who missed block through consensus addr
func GetValMissedBlockPrefixKey(v sdk.Address) []byte {
	return append(ValidatorMissedBlockBitArrayKey, v.Bytes()...)
}

// generates the key for missing val who missed block through consensus addr
func GetValMissedBlockKey(v sdk.Address, i int64) []byte {
	b := make([]byte, 8)
	binary.LittleEndian.PutUint64(b, uint64(i))
	return append(GetValMissedBlockPrefixKey(v), b...)
}

// GetHistoricalInfoKey returns a key prefix for indexing HistoricalInfo objects.
func GetHistoricalInfoKey(height int64) []byte {
	return append(HistoricalInfoKey, []byte(strconv.FormatInt(height, 10))...)
}

// AddressFromLastValidatorPowerKey creates the validator operator address from LastValidatorPowerKey
func AddressFromLastValidatorPowerKey(key []byte) []byte {
	kv.AssertKeyAtLeastLength(key, 3)
	return key[1:]
}

func ScoresToPower(reportCard ReportCard) int64 {
	if reportCard.TotalSessions == 0 {
		return sdk.NewIntWithDecimal(1, 1).Int64()
	}

	LatencyScore := reportCard.TotalLatencyScore
	AvailabilityScore := reportCard.TotalAvailabilityScore
	ReliabilityScore := reportCard.TotalReliabilityScore

	// Adjust the weights based on your preference
	latencyWeight := sdk.NewDecWithPrec(5, 1)
	availabilityWeight := sdk.NewDecWithPrec(2, 1)
	reliabilityWeight := sdk.NewDecWithPrec(3, 1)

	totalScore := LatencyScore.Mul(latencyWeight).Add(
		AvailabilityScore.Mul(availabilityWeight).Add(
			ReliabilityScore.Mul(reliabilityWeight)))

	reducedPower := totalScore.Quo(sdk.NewDecFromInt(sdk.NewInt(1000000000000))).BigInt().Int64()
	return reducedPower
}

// generates the key for a validator in the report card set
func KeyForValidatorInReportCardSet(validator Validator) []byte {
	return getReportCardPowerRankKey(validator)
}

// get the power ranking key of a validator based on the report card score
func getReportCardPowerRankKey(validator Validator) []byte {
	reportPower := ScoresToPower(validator.ReportCard)
	reportPowerBytes := make([]byte, 8)
	binary.BigEndian.PutUint64(reportPowerBytes, uint64(reportPower))

	powerBytes := reportPowerBytes
	powerBytesLen := len(powerBytes) // 8

	// key is of format prefix || powerbytes || addrBytes
	key := make([]byte, 1+powerBytesLen+sdk.AddrLen)

	key[0] = ReportCardKey[0] // Make sure you have a unique prefix for the report card set
	copy(key[1:powerBytesLen+1], powerBytes)
	operAddrInvr := sdk.CopyBytes(validator.Address)
	for i, b := range operAddrInvr {
		operAddrInvr[i] = ^b
	}
	copy(key[powerBytesLen+1:], operAddrInvr)

	return key
}
