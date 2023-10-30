package keeper

import (
	"encoding/binary"
	"math/rand"
	"testing"

	"time"

	sdk "github.com/vipernet-xyz/viper-network/types"
	providersTypes "github.com/vipernet-xyz/viper-network/x/providers/types"
	"github.com/vipernet-xyz/viper-network/x/vipernet/types"

	"github.com/stretchr/testify/assert"
)

func TestKeeper_ValidateProof(t *testing.T) {
	relaysDone := 8
	maxRelays := int64(5)
	ctx, _, _, _, keeper, keys, _ := createTestInput(t, false)
	types.ClearEvidence(types.GlobalEvidenceCache)
	npk, header, _ := simulateRelays(t, keeper, &ctx, relaysDone)
	evidence, err := types.GetEvidence(header, types.RelayEvidence, sdk.NewInt(1000), types.GlobalEvidenceCache)
	if err != nil {
		t.Fatalf("Set evidence not found")
	}
	root := evidence.GenerateMerkleRoot(0, maxRelays, types.GlobalEvidenceCache)
	_, totalRelays := types.GetTotalProofs(header, types.RelayEvidence, sdk.NewInt(1000), types.GlobalEvidenceCache)
	assert.Equal(t, totalRelays, int64(relaysDone))

	claimMsg := types.MsgClaim{
		SessionHeader: header,
		MerkleRoot:    root,
		TotalProofs:   maxRelays,
		FromAddress:   sdk.Address(npk.Address()),
		EvidenceType:  types.RelayEvidence,
	}

	mockCtx := &Ctx{}
	mockCtx.On("KVStore", keeper.storeKey).Return(ctx.KVStore(keeper.storeKey))
	mockCtx.On("KVStore", keys[sdk.ParamsKey.Name()]).Return(ctx.KVStore(keys[sdk.ParamsKey.Name()]))
	mockCtx.On("KVStore", keys[providersTypes.StoreKey]).Return(ctx.KVStore(keys[providersTypes.StoreKey]))
	mockCtx.On("Logger").Return(ctx.Logger())
	mockCtx.On("BlockHeight").Return(ctx.BlockHeight())
	mockCtx.On("PrevCtx", header.SessionBlockHeight).Return(ctx, nil)

	// Calculate the expected proofHeight for the mock:
	expectedProofHeight := header.SessionBlockHeight + keeper.ClaimSubmissionWindow(ctx)*keeper.BlocksPerSession(ctx)
	mockCtx.On("GetPrevBlockHash", expectedProofHeight).Return(ctx.BlockHeader().LastBlockId.Hash, nil)

	neededLeafIndex, er := keeper.getPseudorandomIndex(mockCtx, maxRelays, header, mockCtx)
	assert.Nil(t, er)
	merkleProofs, _ := evidence.GenerateMerkleProof(0, int(neededLeafIndex), maxRelays)

	leafNode := types.GetProof(header, types.RelayEvidence, neededLeafIndex, types.GlobalEvidenceCache)

	proofMsg := types.MsgProof{
		MerkleProof:  merkleProofs,
		Leaf:         leafNode,
		EvidenceType: types.RelayEvidence,
	}

	err = keeper.SetClaim(mockCtx, claimMsg)
	if err != nil {
		t.Fatal(err)
	}
	_, _, err = keeper.ValidateProof(mockCtx, proofMsg)
	if err != nil {
		t.Fatalf(err.Error())
	}
}

func TestKeeper_GetPsuedorandomIndex(t *testing.T) {
	var totalRelays = []int{10, 100, 10000000}
	for _, relays := range totalRelays {
		ctx, _, _, _, keeper, keys, _ := createTestInput(t, false)
		header := types.SessionHeader{
			ProviderPubKey:     "asdlfj",
			Chain:              "lkajsdf",
			GeoZone:            "asdlfj",
			SessionBlockHeight: 1,
			NumServicers:       5,
		}
		mockCtx := new(Ctx)
		mockCtx.On("KVStore", keeper.storeKey).Return(ctx.KVStore(keeper.storeKey))
		mockCtx.On("KVStore", keys["params"]).Return(ctx.KVStore(keys["params"]))

		expectedBlockHeight := header.SessionBlockHeight + keeper.ClaimSubmissionWindow(ctx)*keeper.BlocksPerSession(ctx)
		mockCtx.On("PrevCtx", expectedBlockHeight).Return(ctx, nil)
		mockCtx.On("GetPrevBlockHash", expectedBlockHeight).Return(ctx.BlockHeader().LastBlockId.Hash, nil)

		// generate the pseudorandom proof
		neededLeafIndex, err := keeper.getPseudorandomIndex(mockCtx, int64(relays), header, mockCtx)
		assert.Nil(t, err)
		assert.LessOrEqual(t, neededLeafIndex, int64(relays))
	}
}

func TestPseudoRandomSelection(t *testing.T) {
	// maximum index selection
	const max = uint64(1000)
	const iterations = 10000
	// an index account array for proof
	dataArr := make([]int64, max)
	// run a for loop for statistics
	for i := 0; i < iterations; i++ {
		// create random seed data
		seed := make([]byte, 8)
		binary.LittleEndian.PutUint64(seed, rand.New(rand.NewSource(time.Now().UnixNano())).Uint64())
		// hash for show and convert back to decimal
		blockHashDecimal := binary.LittleEndian.Uint64(types.Hash(seed))
		// mod the selection
		selection := blockHashDecimal % max
		// increment the data
		dataArr[selection] = dataArr[selection] + 1
	}
	// print the results
	// for i := 0; uint64(i) < max; i++ {
	// 	fmt.Printf("index %d, was selected %d times\n", i, dataArr[i])
	// }
}
