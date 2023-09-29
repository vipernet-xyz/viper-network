package app

import (
	"fmt"
	"log"
	"os"
	"time"

	tmType "github.com/tendermint/tendermint/types"

	crypto "github.com/vipernet-xyz/viper-network/crypto/codec"
	ibc "github.com/vipernet-xyz/viper-network/modules/core"
	sdk "github.com/vipernet-xyz/viper-network/types"
	"github.com/vipernet-xyz/viper-network/types/module"
	"github.com/vipernet-xyz/viper-network/x/authentication"
	"github.com/vipernet-xyz/viper-network/x/capability"
	"github.com/vipernet-xyz/viper-network/x/governance"
	governanceTypes "github.com/vipernet-xyz/viper-network/x/governance/types"
	providers "github.com/vipernet-xyz/viper-network/x/providers"
	providersTypes "github.com/vipernet-xyz/viper-network/x/providers/types"
	"github.com/vipernet-xyz/viper-network/x/servicers"
	servicersTypes "github.com/vipernet-xyz/viper-network/x/servicers/types"
	transfer "github.com/vipernet-xyz/viper-network/x/transfer"
	viper "github.com/vipernet-xyz/viper-network/x/vipernet"
	"github.com/vipernet-xyz/viper-network/x/vipernet/types"
)

var mainnetGenesis = `{ }`

var testnetGenesis = `{ }`

func GenesisStateFromJson(json string) GenesisState {
	genDoc, err := tmType.GenesisDocFromJSON([]byte(json))
	if err != nil {
		fmt.Println("unable to read genesis from json (internal)")
		os.Exit(1)
	}
	return GenesisStateFromGenDoc(cdc, *genDoc)
}

func newDefaultGenesisState() []byte {
	keyb, err := GetKeybase()
	if err != nil {
		log.Fatal(err)
	}
	cb, err := keyb.GetCoinbase()
	if err != nil {
		log.Fatal(err)
	}
	pubKey := cb.PublicKey
	defaultGenesis := module.NewBasicManager(
		providers.AppModuleBasic{},
		authentication.AppModuleBasic{},
		capability.AppModuleBasic{},
		ibc.AppModuleBasic{},
		governance.AppModuleBasic{},
		servicers.AppModuleBasic{},
		transfer.AppModuleBasic{},
		viper.AppModuleBasic{},
	).DefaultGenesis()
	// setup account genesis
	rawAuth := defaultGenesis[authentication.ModuleName]
	var accountGenesis authentication.GenesisState
	types.ModuleCdc.MustUnmarshalJSON(rawAuth, &accountGenesis)
	accountGenesis.Accounts = append(accountGenesis.Accounts, &authentication.BaseAccount{
		Address: cb.GetAddress(),
		Coins:   sdk.NewCoins(sdk.NewCoin(sdk.DefaultStakeDenom, sdk.NewInt(10000000000000))),
		PubKey:  pubKey,
	})
	res := Codec().MustMarshalJSON(accountGenesis)
	defaultGenesis[authentication.ModuleName] = res
	// set address as provider too
	rawApps := defaultGenesis[providersTypes.ModuleName]
	var providersGenesis providersTypes.GenesisState
	types.ModuleCdc.MustUnmarshalJSON(rawApps, &providersGenesis)
	providersGenesis.Providers = append(providersGenesis.Providers, providersTypes.Provider{
		Address:                 cb.GetAddress(),
		PublicKey:               cb.PublicKey,
		Jailed:                  false,
		Status:                  2,
		Chains:                  []string{sdk.PlaceholderHash},
		GeoZones:                []string{sdk.PlaceholderGeoZone},
		NumServicers:            5,
		StakedTokens:            sdk.NewInt(10000000000000),
		MaxRelays:               sdk.NewInt(10000000000000),
		UnstakingCompletionTime: time.Time{},
	})
	res = Codec().MustMarshalJSON(providersGenesis)
	defaultGenesis[providersTypes.ModuleName] = res
	rawViper := defaultGenesis[types.ModuleName]
	var viperGenesis types.GenesisState
	types.ModuleCdc.MustUnmarshalJSON(rawViper, &viperGenesis)
	res = Codec().MustMarshalJSON(viperGenesis)
	defaultGenesis[types.ModuleName] = res
	// setup pos genesis
	rawPOS := defaultGenesis[servicersTypes.ModuleName]
	var posGenesisState servicersTypes.GenesisState
	types.ModuleCdc.MustUnmarshalJSON(rawPOS, &posGenesisState)
	posGenesisState.Validators = append(posGenesisState.Validators,
		servicersTypes.Validator{
			Address:                 sdk.Address(pubKey.Address()),
			PublicKey:               pubKey,
			Jailed:                  false,
			Paused:                  false,
			Status:                  sdk.Staked,
			Chains:                  []string{sdk.PlaceholderHash},
			GeoZone:                 []string{sdk.PlaceholderGeoZone},
			ServiceURL:              sdk.PlaceholderServiceURL,
			StakedTokens:            sdk.NewInt(10000000000000),
			UnstakingCompletionTime: time.Time{},
			ReportCard:              servicersTypes.ReportCard{TotalSessions: 0, TotalLatencyScore: sdk.NewDec(0), TotalAvailabilityScore: sdk.NewDec(0)}})
	res = types.ModuleCdc.MustMarshalJSON(posGenesisState)
	defaultGenesis[servicersTypes.ModuleName] = res
	// set default governance in genesis
	var governanceGenesisState governanceTypes.GenesisState
	rawGov := defaultGenesis[governanceTypes.ModuleName]
	Codec().MustUnmarshalJSON(rawGov, &governanceGenesisState)
	mACL := createDummyACL(pubKey)
	governanceGenesisState.Params.ACL = mACL
	governanceGenesisState.Params.DAOOwner = sdk.Address(pubKey.Address())
	governanceGenesisState.Params.Upgrade = governanceTypes.NewUpgrade(0, "0")
	res4 := Codec().MustMarshalJSON(governanceGenesisState)
	defaultGenesis[governanceTypes.ModuleName] = res4
	// end genesis setup
	j, _ := types.ModuleCdc.MarshalJSONIndent(defaultGenesis, "", "    ")
	j, _ = types.ModuleCdc.MarshalJSONIndent(tmType.GenesisDoc{
		GenesisTime: time.Now(),
		ChainID:     "viper-test",
		ConsensusParams: &tmType.ConsensusParams{
			Block: tmType.BlockParams{
				MaxBytes:   15000,
				MaxGas:     -1,
				TimeIotaMs: 1,
			},
			Evidence: tmType.EvidenceParams{
				MaxAge: 1000000,
			},
			Validator: tmType.ValidatorParams{
				PubKeyTypes: []string{"ed25519"},
			},
		},
		Validators: nil,
		AppHash:    nil,
		AppState:   j,
	}, "", "    ")
	return j
}

func createDummyACL(kp crypto.PublicKey) governanceTypes.ACL {
	addr := sdk.Address(kp.Address())
	acl := governanceTypes.ACL{}
	acl = make([]governanceTypes.ACLPair, 0)
	acl.SetOwner("provider/MinimumProviderStake", addr)
	acl.SetOwner("provider/ProviderUnstakingTime", addr)
	acl.SetOwner("provider/BaseRelaysPerVIPR", addr)
	acl.SetOwner("provider/MaxProviders", addr)
	acl.SetOwner("provider/MaximumChains", addr)
	acl.SetOwner("provider/ParticipationRate", addr)
	acl.SetOwner("provider/StabilityModulation", addr)
	acl.SetOwner("provider/MinNumServicers", addr)
	acl.SetOwner("provider/MaxNumServicers", addr)
	acl.SetOwner("provider/MaxFreeTierRelaysPerSession", addr)
	acl.SetOwner("authentication/MaxMemoCharacters", addr)
	acl.SetOwner("authentication/TxSigLimit", addr)
	acl.SetOwner("authentication/FeeMultipliers", addr)
	acl.SetOwner("governance/acl", addr)
	acl.SetOwner("governance/daoOwner", addr)
	acl.SetOwner("governance/upgrade", addr)
	acl.SetOwner("vipernet/ClaimExpiration", addr)
	acl.SetOwner("vipernet/ReplayAttackBurnMultiplier", addr)
	acl.SetOwner("vipernet/ClaimSubmissionWindow", addr)
	acl.SetOwner("vipernet/MinimumNumberOfProofs", addr)
	acl.SetOwner("vipernet/SupportedBlockchains", addr)
	acl.SetOwner("vipernet/SupportedGeoZones", addr)
	acl.SetOwner("vipernet/MinimumSampleRelays", addr)
	acl.SetOwner("pos/BlocksPerSession", addr)
	acl.SetOwner("pos/DAOAllocation", addr)
	acl.SetOwner("pos/ProviderAllocation", addr)
	acl.SetOwner("pos/DowntimeJailDuration", addr)
	acl.SetOwner("pos/MaxEvidenceAge", addr)
	acl.SetOwner("pos/MaximumChains", addr)
	acl.SetOwner("pos/MaxJailedBlocks", addr)
	acl.SetOwner("pos/MaxValidators", addr)
	acl.SetOwner("pos/MinSignedPerWindow", addr)
	acl.SetOwner("pos/TokenRewardFactor", addr)
	acl.SetOwner("pos/SignedBlocksWindow", addr)
	acl.SetOwner("pos/SlashFractionDoubleSign", addr)
	acl.SetOwner("pos/SlashFractionDowntime", addr)
	acl.SetOwner("pos/StakeDenom", addr)
	acl.SetOwner("pos/StakeMinimum", addr)
	acl.SetOwner("pos/UnstakingTime", addr)
	acl.SetOwner("pos/ServicerCountLock", addr)
	acl.SetOwner("pos/BurnActive", addr)
	acl.SetOwner("pos/MinPauseTime", addr)
	acl.SetOwner("pos/MaxFishermen", addr)
	acl.SetOwner("pos/FishermenCount", addr)
	acl.SetOwner("pos/SlashFractionNoActivity", addr)
	acl.SetOwner("pos/ProposerPercentage", addr)
	acl.SetOwner("pos/ProviderAllocation", addr)
	return acl
}
