package keeper

import (
	"bytes"
	"fmt"
	"sort"

	sdk "github.com/vipernet-xyz/viper-network/types"
	"github.com/vipernet-xyz/viper-network/x/servicers/exported"
	"github.com/vipernet-xyz/viper-network/x/servicers/types"
)

func (k Keeper) MarshalValidator(ctx sdk.Ctx, validator types.Validator) ([]byte, error) {
	bz, err := k.Cdc.MarshalBinaryLengthPrefixed(&validator, ctx.BlockHeight())
	if err != nil {
		ctx.Logger().Error("could not marshal validator: " + err.Error())
	}
	return bz, err
}

func (k Keeper) UnmarshalValidator(ctx sdk.Ctx, valBytes []byte) (val types.Validator, err error) {
	err = k.Cdc.UnmarshalBinaryLengthPrefixed(valBytes, &val, ctx.BlockHeight())
	if err != nil {
		ctx.Logger().Error("could not unmarshal validator: " + err.Error())
	}
	return val, err

}

// GetValidator - Retrieve validator with address from the main store
func (k Keeper) GetValidator(ctx sdk.Ctx, addr sdk.Address) (validator types.Validator, found bool) {
	val, found := k.validatorCache.GetWithCtx(ctx, addr.String())
	if found {
		return val.(types.Validator), found
	}
	store := ctx.KVStore(k.storeKey)
	value, _ := store.Get(types.KeyForValByAllVals(addr))
	if value == nil {
		return validator, false
	}
	validator, err := k.UnmarshalValidator(ctx, value)
	if err != nil {
		ctx.Logger().Error("can't get validator: " + err.Error())
		return validator, false
	}
	_ = k.validatorCache.AddWithCtx(ctx, addr.String(), validator)
	return validator, true
}

// SetValidator - Store validator in the main store and state stores (stakingset/ unstakingset)
func (k Keeper) SetValidator(ctx sdk.Ctx, validator types.Validator) {
	store := ctx.KVStore(k.storeKey)
	bz, err := k.MarshalValidator(ctx, validator)
	if err != nil {
		ctx.Logger().Error("could not marshal validator: " + err.Error())
	}
	err = store.Set(types.KeyForValByAllVals(validator.Address), bz)
	if err != nil {
		ctx.Logger().Error("could not set validator: " + err.Error())
	}
	if validator.IsUnstaking() {
		// Adds to unstaking validator queue
		k.SetUnstakingValidator(ctx, validator)
	}
	if validator.IsStaked() {
		if !validator.IsJailed() {
			// save in the staked store
			k.SetStakedValidator(ctx, validator)
		}
	}
	_ = k.validatorCache.AddWithCtx(ctx, validator.Address.String(), validator)
}

func (k Keeper) SetValidators(ctx sdk.Ctx, validators types.Validators) {
	for _, val := range validators {
		k.SetValidator(ctx, val)
	}
}

func (k Keeper) GetValidatorOutputAddress(ctx sdk.Ctx, operatorAddress sdk.Address) (sdk.Address, bool) {
	val, found := k.GetValidator(ctx, operatorAddress)
	if val.OutputAddress == nil {
		return val.Address, found
	}
	return val.OutputAddress, found
}

func (k Keeper) DeleteValidator(ctx sdk.Ctx, addr sdk.Address) {
	store := ctx.KVStore(k.storeKey)
	_ = store.Delete(types.KeyForValByAllVals(addr))
	k.DeleteValidatorSigningInfo(ctx, addr)
	k.validatorCache.RemoveWithCtx(ctx, addr.String())
}

// GetAllValidators - Retrieve set of all validators with no limits from the main store
func (k Keeper) GetAllValidators(ctx sdk.Ctx) (validators []types.Validator) {
	validators = make([]types.Validator, 0)
	store := ctx.KVStore(k.storeKey)
	iterator, _ := sdk.KVStorePrefixIterator(store, types.AllValidatorsKey)
	defer iterator.Close()
	for ; iterator.Valid(); iterator.Next() {
		validator, err := k.UnmarshalValidator(ctx, iterator.Value())
		if err != nil {
			ctx.Logger().Error("can't get validator in GetAllValidators: " + err.Error())
			continue
		}
		validators = append(validators, validator)
	}
	return validators
}

// GetAllValidators - Retrieve set of all validators with no limits from the main store
func (k Keeper) GetAllValidatorsAddrs(ctx sdk.Ctx) (validators []sdk.Address) {
	validators = make([]sdk.Address, 0)
	store := ctx.KVStore(k.storeKey)
	iterator, _ := sdk.KVStorePrefixIterator(store, types.AllValidatorsKey)
	defer iterator.Close()
	for ; iterator.Valid(); iterator.Next() {
		validators = append(validators, iterator.Key())
	}
	return validators
}

// GetAllValidators - - Retrieve the set of all validators with no limits from the main store
func (k Keeper) GetAllValidatorsWithOpts(ctx sdk.Ctx, opts types.QueryValidatorsParams) (validators []types.Validator) {
	validators = make([]types.Validator, 0)
	store := ctx.KVStore(k.storeKey)
	iterator, _ := sdk.KVStorePrefixIterator(store, types.AllValidatorsKey)
	defer iterator.Close()
	for ; iterator.Valid(); iterator.Next() {
		validator, err := k.UnmarshalValidator(ctx, iterator.Value())
		if err != nil {
			ctx.Logger().Error("could not unmarshal validator in GetAllValidatorsWithOpts: ", err.Error())
			continue
		}
		if opts.IsValid(validator) {
			validators = append(validators, validator)
		}
	}
	return validators
}

// GetValidators - Retrieve a given amount of all the validators
func (k Keeper) GetValidators(ctx sdk.Ctx, maxRetrieve uint16) (validators []types.Validator) {
	store := ctx.KVStore(k.storeKey)
	validators = make([]types.Validator, maxRetrieve)
	iterator, _ := sdk.KVStorePrefixIterator(store, types.AllValidatorsKey)
	defer iterator.Close()
	i := 0
	for ; iterator.Valid() && i < int(maxRetrieve); iterator.Next() {
		validator, err := k.UnmarshalValidator(ctx, iterator.Value())
		if err != nil {
			ctx.Logger().Error("could not unmarshal validator in GetValidators: ", err.Error())
			continue
		}
		validators[i] = validator
		i++
	}
	return validators[:i] // trim if the array length < maxRetrieve
}

func (k Keeper) ClearSessionCache() {
	if k.ViperKeeper != nil {
		k.ViperKeeper.ClearSessionCache()
	}
}

// IterateAndExecuteOverVals - Goes through the validator set and executes handler
func (k Keeper) IterateAndExecuteOverVals(
	ctx sdk.Ctx, fn func(index int64, validator exported.ValidatorI) (stop bool)) {
	store := ctx.KVStore(k.storeKey)
	iterator, _ := sdk.KVStorePrefixIterator(store, types.AllValidatorsKey)
	defer iterator.Close()
	i := int64(0)
	for ; iterator.Valid(); iterator.Next() {
		validator, err := k.UnmarshalValidator(ctx, iterator.Value())
		if err != nil {
			ctx.Logger().Error("could not unmarshal validator in IterateAndExecuteOverVals: ", err.Error())
			continue
		}
		stop := fn(i, validator) // XXX is this safe will the validator unexposed fields be able to get written to?
		if stop {
			break
		}
		i++
	}
}

// Validator - wrproviderer for GetValidator call
func (k Keeper) Validator(ctx sdk.Ctx, address sdk.Address) exported.ValidatorI {
	val, found := k.GetValidator(ctx, address)
	if !found {
		return nil
	}
	return val
}

// AllValidators - Retrieve a list of all validators
func (k Keeper) AllValidators(ctx sdk.Ctx) (validators []exported.ValidatorI) {
	store := ctx.KVStore(k.storeKey)
	iterator, _ := sdk.KVStorePrefixIterator(store, types.AllValidatorsKey)
	defer iterator.Close()

	for ; iterator.Valid(); iterator.Next() {
		validator, err := k.UnmarshalValidator(ctx, iterator.Value())
		if err != nil {
			ctx.Logger().Error("could not unmarshal validator in AllValidators: ", err.Error())
			continue
		}
		validators = append(validators, validator)
	}
	return validators
}

// map of validator addresses to serialized power
type valPowerMap map[[sdk.AddrLen]byte][]byte

// getPrevStatePowerMap - Retrieve the prevState validator set
func (k Keeper) getPrevStatePowerMap(ctx sdk.Ctx) valPowerMap {
	prevState := make(valPowerMap)
	store := ctx.KVStore(k.storeKey)
	iterator, _ := sdk.KVStorePrefixIterator(store, types.PrevStateValidatorsPowerKey)
	defer iterator.Close()
	// iterate over the prevState validator set index
	for ; iterator.Valid(); iterator.Next() {
		var valAddr [sdk.AddrLen]byte
		// extract the validator address from the key (prefix is 1-byte)
		copy(valAddr[:], iterator.Key()[1:])
		// power bytes is just the value
		powerBytes := iterator.Value()
		prevState[valAddr] = make([]byte, len(powerBytes))
		copy(prevState[valAddr], powerBytes)
	}
	return prevState
}

// sortNoLongerStakedValidators - Given a map of remaining validators to previous staked power
// returns the list of validators to be unbstaked, sorted by operator address
func sortNoLongerStakedValidators(prevState valPowerMap) [][]byte {
	// sort the map keys for determinism
	noLongerStaked := make([][]byte, len(prevState))
	index := 0
	for valAddrBytes := range prevState {
		valAddr := make([]byte, sdk.AddrLen)
		copy(valAddr, valAddrBytes[:])
		noLongerStaked[index] = valAddr
		index++
	}
	// sorted by address - order doesn't matter
	sort.SliceStable(noLongerStaked, func(i, j int) bool {
		// -1 means strictly less than
		return bytes.Compare(noLongerStaked[i], noLongerStaked[j]) == -1
	})
	return noLongerStaked
}

// get the group of the bonded validators
func (k Keeper) GetLastValidators(ctx sdk.Ctx) (validators []types.Validator) {
	store := ctx.KVStore(k.storeKey)

	// add the actual validator power sorted store
	maxValidators := k.MaxValidators(ctx)
	validators = make([]types.Validator, maxValidators)

	iterator, _ := sdk.KVStorePrefixIterator(store, types.LastValidatorPowerKey)
	defer iterator.Close()

	i := 0
	for ; iterator.Valid(); iterator.Next() {
		// sanity check
		if i >= int(maxValidators) {
			panic("more validators than maxValidators found")
		}

		address := types.AddressFromLastValidatorPowerKey(iterator.Key())
		validator := k.mustGetValidator(ctx, address)

		validators[i] = validator
		i++
	}

	return validators[:i] // trim
}

func (k Keeper) mustGetValidator(ctx sdk.Ctx, addr sdk.Address) types.Validator {
	validator, found := k.GetValidator(ctx, addr)
	if !found {
		panic(fmt.Sprintf("validator record not found for address: %X\n", addr))
	}

	return validator
}

func (k Keeper) InitializeReportCardForValidator(ctx sdk.Ctx, validator *types.Validator) {
	validator.ReportCard = types.ReportCard{
		TotalSessions:          0,
		TotalLatencyScore:      sdk.BigDec{},
		TotalAvailabilityScore: sdk.BigDec{},
	}
}

func (k Keeper) SetValidatorReportCard(ctx sdk.Ctx, validator types.Validator) {
	store := ctx.KVStore(k.storeKey)
	bz, err := k.MarshalValidator(ctx, validator) // If MarshalValidator includes the ReportCard field
	if err != nil {
		ctx.Logger().Error("could not marshal validator report card: " + err.Error())
	}
	err = store.Set(types.KeyForValidatorInReportCardSet(validator), bz)
	if err != nil {
		ctx.Logger().Error("could not set validator report card: " + err.Error())
	}
}

func (k Keeper) GetValidatorReportCard(ctx sdk.Ctx, validator types.Validator) (reportCard types.ReportCard, found bool) {
	store := ctx.KVStore(k.storeKey)
	value, _ := store.Get(types.KeyForValidatorInReportCardSet(validator))
	if value == nil {
		return reportCard, false
	}
	validator, err := k.UnmarshalValidator(ctx, value) // If UnmarshalValidator includes the ReportCard field
	if err != nil {
		ctx.Logger().Error("can't get validator report card: " + err.Error())
		return reportCard, false
	}
	return validator.ReportCard, true
}

func (k Keeper) UpdateValidatorReportCard(ctx sdk.Ctx, addr sdk.Address, sessionReport types.ReportCard) {
	validator, found := k.GetValidator(ctx, addr)
	if !found {
		ctx.Logger().Error(fmt.Sprintf("validator not found for address: %X\n", addr))
		return
	}

	// Increase the total sessions count
	validator.ReportCard.TotalSessions += sessionReport.TotalSessions

	newLatencyScore := validator.ReportCard.TotalLatencyScore.Add(sessionReport.TotalLatencyScore).Quo(sdk.NewDec(int64(validator.ReportCard.TotalSessions)))
	if newLatencyScore.GT(sdk.NewDec(1)) {
		newLatencyScore = sdk.NewDec(1)
	}
	validator.ReportCard.TotalLatencyScore = newLatencyScore

	newAvailabilityScore := validator.ReportCard.TotalAvailabilityScore.Add(sessionReport.TotalAvailabilityScore).Quo(sdk.NewDec(int64(validator.ReportCard.TotalSessions)))
	if newAvailabilityScore.GT(sdk.NewDec(1)) {
		newAvailabilityScore = sdk.NewDec(1)
	}
	validator.ReportCard.TotalAvailabilityScore = newAvailabilityScore

	// Update the reliability score similar to the above scores
	newReliabilityScore := validator.ReportCard.TotalReliabilityScore.Add(sessionReport.TotalReliabilityScore).Quo(sdk.NewDec(int64(validator.ReportCard.TotalSessions)))
	if newReliabilityScore.GT(sdk.NewDec(1)) {
		newReliabilityScore = sdk.NewDec(1)
	}
	validator.ReportCard.TotalReliabilityScore = newReliabilityScore

	// Save the updated validator data
	k.SetValidator(ctx, validator)

	// Set the new report card
	k.SetValidatorReportCard(ctx, validator)
}

// DeleteReportCard deletes the report card of a servicer when they are unstaked
func (k Keeper) deleteValidatorReportCard(ctx sdk.Ctx, validator types.Validator) error {
	store := ctx.KVStore(k.storeKey)

	// Delete the report card from the store
	store.Delete(types.KeyForValidatorInReportCardSet(validator))

	return nil
}
