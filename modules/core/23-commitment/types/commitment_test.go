package types_test

/*
	"testing"

	dbm "github.com/cometbft/cometbft-db"
	"github.com/cometbft/cometbft/libs/log"
	storetypes "github.com/cosmos/cosmos-sdk/store/types"
	"github.com/stretchr/testify/suite"
	"github.com/vipernet-xyz/viper-network/store/iavl"
	"github.com/vipernet-xyz/viper-network/store/rootmulti"*/

/*type MerkleTestSuite struct {
	suite.Suite

	store     *rootmulti.Store
	storeKey  *storetypes.KVStoreKey
	iavlStore *iavl.Store
}

func (suite *MerkleTestSuite) SetupTest() {
	db := dbm.NewMemDB()
	dblog := log.TestingLogger()
	suite.store = rootmulti.NewStore(db, dblog)

	suite.storeKey = storetypes.NewKVStoreKey("iavlStoreKey")

	suite.store.MountStoreWithDB(suite.storeKey, storetypes.StoreTypeIAVL, nil)
	err := suite.store.LoadVersion(0)
	suite.Require().NoError(err)

	suite.iavlStore = suite.store.GetCommitStore(suite.storeKey).(*iavl.Store)
}

func TestMerkleTestSuite(t *testing.T) {
	suite.Run(t, new(MerkleTestSuite))
}
*/
