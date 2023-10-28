package keeper

import (
	"fmt"

	"github.com/tendermint/tendermint/rpc/client"
	crypto "github.com/vipernet-xyz/viper-network/crypto/codec"
	sdk "github.com/vipernet-xyz/viper-network/types"
	auth "github.com/vipernet-xyz/viper-network/x/authentication"
	"github.com/vipernet-xyz/viper-network/x/authentication/util"
	vc "github.com/vipernet-xyz/viper-network/x/vipernet/types"
)

func (k Keeper) SendReportCardTx(ctx sdk.Ctx, keeper Keeper, n client.Client, node *vc.ViperNode, servicerAddr sdk.Address, sessionHeader vc.SessionHeader, evidenceType vc.EvidenceType, qosReport vc.ViperQoSReport, reportCardTx func(pk crypto.PrivateKey, cliCtx util.CLIContext, txBuilder auth.TxBuilder, header vc.SessionHeader, servicerAddr sdk.Address, reportCard vc.ViperQoSReport) (*sdk.TxResponse, error)) {
	// Get the private val key (main) account from the keybase
	fishermanAddr := node.GetAddress()

	// Use GetResult to fetch the result for the given session, servicer address and evidence type
	result, err := vc.GetResult(sessionHeader, evidenceType, servicerAddr, node.TestStore)
	if err != nil {
		ctx.Logger().Error(fmt.Sprintf("An error occurred retrieving the result: \n%s", err.Error()))
		return
	}

	// If the blockchain in the evidence is not supported then return.
	if !k.IsViperSupportedBlockchain(ctx, sessionHeader.Chain) {
		ctx.Logger().Info(fmt.Sprintf("Report card for %s blockchain isn't viper supported, so will not send.", sessionHeader.Chain))
		return
	}

	// Check if the current session is still ongoing
	if ctx.BlockHeight() <= sessionHeader.SessionBlockHeight+k.BlocksPerSession(ctx)-1 {
		ctx.Logger().Info("The session is ongoing, so will not send the report card yet.")
		return
	}

	// Check the current state to see if the report card has already been sent and processed (if so, then return)
	if rc, found := k.GetReportCard(ctx, qosReport.ServicerAddress, fishermanAddr, sessionHeader); found {
		ctx.Logger().Info(fmt.Sprintf("Report card already found for session: %v", rc.SessionHeader))
		return
	}

	// Check if the report card has expired
	if k.ReportCardIsExpired(ctx, result.SessionBlockHeight) {
		// Delete the result since we cannot submit an expired report card
		if err := vc.DeleteResult(sessionHeader, evidenceType, node.TestStore); err != nil {
			ctx.Logger().Debug(err.Error())
		}
		return
	}

	// Check if proof count is above minimum
	if result.NumOfTestResults < k.MinimumSampleRelays(ctx) {
		ctx.Logger().Info("Number of proofs is below the required minimum, will not send report card.")
		return
	}

	// Generate the auto txbuilder and clictx
	txBuilder, cliCtx, err := newTxBuilderAndCliCtx(ctx, &vc.MsgSubmitReportCard{}, n, node.PrivateKey, k)
	if err != nil {
		ctx.Logger().Error(fmt.Sprintf("An error occurred creating the tx builder for the report card tx:\n%s", err.Error()))
		return
	}

	// Send in the report card
	if _, err := reportCardTx(node.PrivateKey, cliCtx, txBuilder, sessionHeader, servicerAddr, qosReport); err != nil {
		ctx.Logger().Error(fmt.Sprintf("An error occurred executing the report card transaction: \n%s", err.Error()))
	}
}

// "SetReportCard" - Sets the report card in the state storage
func (k Keeper) SetReportCard(ctx sdk.Ctx, msg vc.MsgSubmitReportCard) error {
	// retrieve the store
	store := ctx.KVStore(k.storeKey)

	// generate the store key for the report card. Here, I'm assuming a function `KeyForReportCard` similar to `KeyForClaim`.
	key, err := vc.KeyForReportCard(ctx, msg.ServicerAddress, msg.FishermanAddress, msg.SessionHeader)
	if err != nil {
		return err
	}

	// marshal the report card into amino (or the appropriate codec)
	bz, err := k.Cdc.MarshalBinaryBare(&msg)
	if err != nil {
		panic(err)
	}

	// set in the store
	_ = store.Set(key, bz)
	return nil
}

// GetReportCard retrieves the ReportCard message from the store.
func (k Keeper) GetReportCard(ctx sdk.Ctx, servicerAddr sdk.Address, fishermanAddr sdk.Address, header vc.SessionHeader) (msg vc.MsgSubmitReportCard, found bool) {
	// Get the store.
	store := ctx.KVStore(k.storeKey)

	// Generate the key for the ReportCard.
	key, err := vc.KeyForReportCard(ctx, servicerAddr, fishermanAddr, header)
	if err != nil {
		ctx.Logger().Error("Error generating key for report card:", err)
		return vc.MsgSubmitReportCard{}, false
	}

	// Get the report card from the store using the generated key.
	res, _ := store.Get(key)
	if res == nil {
		return vc.MsgSubmitReportCard{}, false
	}

	// Unmarshal the data into the report card object.
	err = k.Cdc.UnmarshalBinaryBare(res, &msg)
	if err != nil {
		panic(err)
	}

	return msg, true
}

func (k Keeper) SetReportCards(ctx sdk.Ctx, reportCards []vc.MsgSubmitReportCard) {
	for _, msg := range reportCards {
		err := k.SetReportCard(ctx, msg)
		if err != nil {
			ctx.Logger().Error("an error occurred setting the report card:\n", msg)
		}
	}
}

func (k Keeper) GetReportCards(ctx sdk.Ctx, servicerAddress sdk.Address, fishermanAddress sdk.Address) (reportCards []vc.MsgSubmitReportCard, err error) {
	// retrieve the store
	store := ctx.KVStore(k.storeKey)

	key, err := vc.KeyForReportCards(servicerAddress, fishermanAddress)
	if err != nil {
		return nil, err
	}
	// iterate through all of the kv pairs and unmarshal into claim objects
	iterator, _ := sdk.KVStorePrefixIterator(store, key)
	defer iterator.Close()
	for ; iterator.Valid(); iterator.Next() {
		var reportCard vc.MsgSubmitReportCard
		err = k.Cdc.UnmarshalBinaryBare(iterator.Value(), &reportCard)
		if err != nil {
			panic(err)
		}
		reportCards = append(reportCards, reportCard)
	}
	return
}

// "GetAllClaims" - Gets all of the submit report card messages held in the state storage.
func (k Keeper) GetAllReportCards(ctx sdk.Ctx) (reportCards []vc.MsgSubmitReportCard) {
	// retrieve the store
	store := ctx.KVStore(k.storeKey)

	iterator, _ := sdk.KVStorePrefixIterator(store, vc.ReportCardKey)
	defer iterator.Close()
	for ; iterator.Valid(); iterator.Next() {
		var reportCard vc.MsgSubmitReportCard
		err := k.Cdc.UnmarshalBinaryBare(iterator.Value(), &reportCard)
		if err != nil {
			panic(err)
		}
		reportCards = append(reportCards, reportCard)
	}
	return
}

func (k Keeper) DeleteReportCard(ctx sdk.Ctx, servicerAddr sdk.Address, fishermanAddr sdk.Address, header vc.SessionHeader) error {
	// retrieve the store
	store := ctx.KVStore(k.storeKey)
	// generate the key for the claim
	key, err := vc.KeyForReportCard(ctx, servicerAddr, fishermanAddr, header)
	if err != nil {
		return err
	}
	// delete it from the state storage
	_ = store.Delete(key)
	return nil
}

func (k Keeper) ReportCardIsExpired(ctx sdk.Ctx, sessionBlockHeight int64) bool {
	expirationWindowInBlocks := k.ReportCardSubmissionWindow(ctx) * k.BlocksPerSession(ctx)
	return ctx.BlockHeight() > expirationWindowInBlocks+sessionBlockHeight
}
