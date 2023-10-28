package keeper

import (
	sdk "github.com/vipernet-xyz/viper-network/types"
	"github.com/vipernet-xyz/viper-network/x/vipernet/types"
)

// "ParamKeyTable" - Registers the paramset types in a keytable and returns the table
func ParamKeyTable() sdk.KeyTable {
	return sdk.NewKeyTable().RegisterParamSet(&types.Params{})
}

// "ClaimExpiration" - Returns the claim expiration parameter from the paramstore
// Number of sessions pass before claim is expired
func (k Keeper) ClaimExpiration(ctx sdk.Ctx) (res int64) {
	k.Paramstore.Get(ctx, types.KeyClaimExpiration, &res)
	return
}

// "ReplayAttackBurn" - Returns the replay attack burn parameter from the paramstore
// The multiplier for how heavily servicers are burned for replay attacks
func (k Keeper) ReplayAttackBurnMultiplier(ctx sdk.Ctx) (res int64) {
	k.Paramstore.Get(ctx, types.KeyReplayAttackBurnMultiplier, &res)
	return
}

// "BlocksPerSession" - Returns blocksPerSession parameter from the paramstore
// How many blocks per session
func (k Keeper) BlocksPerSession(ctx sdk.Ctx) int64 {
	frequency := k.posKeeper.BlocksPerSession(ctx)
	return frequency
}

// "ClaimSubmissionWindow" - Returns claimSubmissionWindow parameter from the paramstore
// How long do you have to submit a claim before the secret is revealed and it's invalid
func (k Keeper) ClaimSubmissionWindow(ctx sdk.Ctx) (res int64) {
	k.Paramstore.Get(ctx, types.KeyClaimSubmissionWindow, &res)
	return
}

// "SupportedBlockchains" - Returns a supported blockchain parameter from the paramstore
// What blockchains are supported in viper network (list of network identifier hashes)
func (k Keeper) SupportedBlockchains(ctx sdk.Ctx) (res []string) {
	k.Paramstore.Get(ctx, types.KeySupportedBlockchains, &res)
	return
}

// "SupportedGeoZones" - Returns a supported geozones parameter from the paramstore
// What geozones are supported in viper network (list of network identifier hashes)
func (k Keeper) SupportedGeoZones(ctx sdk.Ctx) (res []string) {
	k.Paramstore.Get(ctx, types.KeySupportedGeoZones, &res)
	return
}

// "MinimumNumberOfProofs" - Returns a minimun number of proofs parameter from the paramstore
// What blockchains are supported in viper network (list of network identifier hashes)
func (k Keeper) MinimumNumberOfProofs(ctx sdk.Ctx) (res int64) {
	k.Paramstore.Get(ctx, types.KeyMinimumNumberOfProofs, &res)
	return
}

func (k Keeper) BlockByteSize(ctx sdk.Ctx) (res int64) {
	k.Paramstore.Get(ctx, types.KeyBlockByteSize, &res)
	return
}

func (k Keeper) MinimumSampleRelays(ctx sdk.Ctx) (res int64) {
	k.Paramstore.Get(ctx, types.KeyMinimumSampleRelays, &res)
	return
}

func (k Keeper) ReportCardSubmissionWindow(ctx sdk.Ctx) (res int64) {
	k.Paramstore.Get(ctx, types.KeyReportCardSubmissionWindow, &res)
	return
}

// "GetParams" - Returns all module parameters in a `Params` struct
func (k Keeper) GetParams(ctx sdk.Ctx) types.Params {
	return types.Params{
		ClaimSubmissionWindow:      k.ClaimSubmissionWindow(ctx),
		SupportedBlockchains:       k.SupportedBlockchains(ctx),
		ClaimExpiration:            k.ClaimExpiration(ctx),
		ReplayAttackBurnMultiplier: k.ReplayAttackBurnMultiplier(ctx),
		MinimumNumberOfProofs:      k.MinimumNumberOfProofs(ctx),
		BlockByteSize:              k.BlockByteSize(ctx),
		SupportedGeoZones:          k.SupportedGeoZones(ctx),
		MinimumSampleRelays:        k.MinimumSampleRelays(ctx),
		ReportCardSubmissionWindow: k.ReportCardSubmissionWindow(ctx),
	}
}

// "SetParams" - Sets all of the parameters in the paramstore using the params structure
func (k Keeper) SetParams(ctx sdk.Ctx, params types.Params) {
	k.Paramstore.SetParamSet(ctx, &params)
}
