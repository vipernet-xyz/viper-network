package keeper

import (
	"encoding/binary"
	"math/rand"
	"testing"

	"time"

	"github.com/vipernet-xyz/viper-network/x/vipernet/types"

	"github.com/stretchr/testify/assert"
)

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
