package app

import (
	"fmt"
	"log"
	"os"
	"time"

	tmType "github.com/tendermint/tendermint/types"

	"github.com/vipernet-xyz/viper-network/crypto"
	sdk "github.com/vipernet-xyz/viper-network/types"
	"github.com/vipernet-xyz/viper-network/types/module"
	apps "github.com/vipernet-xyz/viper-network/x/apps"
	appsTypes "github.com/vipernet-xyz/viper-network/x/apps/types"
	"github.com/vipernet-xyz/viper-network/x/auth"
	"github.com/vipernet-xyz/viper-network/x/gov"
	govTypes "github.com/vipernet-xyz/viper-network/x/gov/types"
	"github.com/vipernet-xyz/viper-network/x/nodes"
	nodesTypes "github.com/vipernet-xyz/viper-network/x/nodes/types"
	viper "github.com/vipernet-xyz/viper-network/x/vipercore"
	"github.com/vipernet-xyz/viper-network/x/vipercore/types"
)

var mainnetGenesis = `{
    "genesis_time": "2022-09-23T12:39:45.880637Z",
    "chain_id": "viper-mainnet",
    "consensus_params": {
        "block": {
            "max_bytes": "15000",
            "max_gas": "-1",
            "time_iota_ms": "1"
        },
        "evidence": {
            "max_age": "1000000"
        },
        "validator": {
            "pub_key_types": [
                "ed25519"
            ]
        }
    },
    "app_hash": "",
    "app_state": {
        "pos": {
            "params": {
                "relays_to_tokens_multiplier": "1000",
                "unstaking_time": "1814400000000000",
                "max_validators": "5000",
                "stake_denom": "uvipr",
                "stake_minimum": "15000000000",
                "session_block_frequency": "4",
                "dao_allocation": "8",
                "app_allocation": "2",
                "proposer_allocation": "1",
                "maximum_chains": "15",
                "max_jailed_blocks": "1000",
                "max_evidence_age": "120000000000",
                "signed_blocks_window": "100",
                "min_signed_per_window": "0.500000000000000000",
                "downtime_jail_duration": "3600000000000",
                "slash_fraction_double_sign": "0.050000000000000000",
                "slash_fraction_downtime": "0.010000000000000000",
                "servicer_stake_floor_multipler": "0",
                "servicer_stake_weight_multipler": "0.000000000000000000",
                "servicer_stake_weight_ceiling": "0",
                "servicer_stake_floor_multiplier_exponent": "0.000000000000000000"
            },
            "prevState_total_power": "0",
            "prevState_validator_powers": null,
            "validators": [
                {
                    "address": "52264967f262a7c55a2b570d3d2de409161521b8",
                    "public_key": "41a7bd126a282a5ccaa5e060c81d41c64912f2a76b94dc29043b3f580655d805",
                    "jailed": false,
                    "status": 2,
                    "chains": [
                        "0001"
                    ],
                    "service_url": "http://127.0.0.1:8081",
                    "tokens": "10000000",
                    "unstaking_time": "0001-01-01T00:00:00Z",
                    "output_address": ""
                }
            ],
            "exported": false,
            "signing_infos": {},
            "missed_blocks": {},
            "previous_proposer": ""
        },
        "vipercore": {
            "params": {
                "session_node_count": "1",
                "proof_waiting_period": "3",
                "supported_blockchains": [
                    "0001"
                ],
                "claim_expiration": "100",
                "replay_attack_burn_multiplier": "3",
                "minimum_number_of_proofs": "5"
            },
            "claims": null
        },
        "application": {
            "params": {
                "unstaking_time": "1814400000000000",
                "max_applications": "9223372036854775807",
                "minimum_app_stake": "1000000",
                "base_relays_per_vip": "100",
                "stability_modulation": "0",
                "participation_rate_on": false,
                "maximum_chains": "15"
            },
            "applications": [
                {
                    "address": "52264967f262a7c55a2b570d3d2de409161521b8",
                    "public_key": "41a7bd126a282a5ccaa5e060c81d41c64912f2a76b94dc29043b3f580655d805",
                    "jailed": false,
                    "chains": [
                        "0001"
                    ],
                    "max_relays": "10000000000000",
                    "status": 2,
                    "staked_tokens": "10000000000000",
                    "unstaking_time": "0001-01-01T00:00:00Z"
                }
            ],
            "exported": false
        },
        "auth": {
            "params": {
                "max_memo_characters": "256",
                "tx_sig_limit": "7",
                "fee_multipliers": {
                    "fee_multiplier": null,
                    "default": "1"
                }
            },
            "accounts": [
                {
                    "type": "posmint/Account",
                    "value": {
                        "address": "52264967f262a7c55a2b570d3d2de409161521b8",
                        "coins": [
                            {
                                "denom": "uvipr",
                                "amount": "6000000000000"
                            }
                        ],
                        "public_key": {
                            "type": "crypto/ed25519_public_key",
                            "value": "41a7bd126a282a5ccaa5e060c81d41c64912f2a76b94dc29043b3f580655d805"
                        }
                    }
                },
                {
                    "type": "posmint/Account",
                    "value": {
                        "address": "cb0c04268ef1acb93ac2143879ec619dcb3f3fbe",
                        "coins": [
                            {
                                "denom": "uvipr",
                                "amount": "15000000000"
                            }
                        ],
                        "public_key": {
                            "type": "crypto/ed25519_public_key",
                            "value": "033e9ccb58fa1794c4eeaad7fe7b445b4794260f6941249e19fed2838b4b027a"
                        }
                    }
                }
            ],
            "supply": []
        },
        "gov": {
            "params": {
                "acl": [
                    {
                        "acl_key": "application/MinimumApplicationStake",
                        "address": "52264967f262a7c55a2b570d3d2de409161521b8"
                    },
                    {
                        "acl_key": "application/AppUnstakingTime",
                        "address": "52264967f262a7c55a2b570d3d2de409161521b8"
                    },
                    {
                        "acl_key": "application/BaseRelaysPerVIPR",
                        "address": "52264967f262a7c55a2b570d3d2de409161521b8"
                    },
                    {
                        "acl_key": "application/MaxApplications",
                        "address": "52264967f262a7c55a2b570d3d2de409161521b8"
                    },
                    {
                        "acl_key": "application/MaximumChains",
                        "address": "52264967f262a7c55a2b570d3d2de409161521b8"
                    },
                    {
                        "acl_key": "application/ParticipationRate",
                        "address": "52264967f262a7c55a2b570d3d2de409161521b8"
                    },
                    {
                        "acl_key": "application/StabilityModulation",
                        "address": "52264967f262a7c55a2b570d3d2de409161521b8"
                    },
                    {
                        "acl_key": "auth/MaxMemoCharacters",
                        "address": "52264967f262a7c55a2b570d3d2de409161521b8"
                    },
                    {
                        "acl_key": "auth/TxSigLimit",
                        "address": "52264967f262a7c55a2b570d3d2de409161521b8"
                    },
                    {
                        "acl_key": "gov/acl",
                        "address": "52264967f262a7c55a2b570d3d2de409161521b8"
                    },
                    {
                        "acl_key": "gov/daoOwner",
                        "address": "52264967f262a7c55a2b570d3d2de409161521b8"
                    },
                    {
                        "acl_key": "gov/upgrade",
                        "address": "52264967f262a7c55a2b570d3d2de409161521b8"
                    },
                    {
                        "acl_key": "vipercore/ClaimExpiration",
                        "address": "52264967f262a7c55a2b570d3d2de409161521b8"
                    },
                    {
                        "acl_key": "auth/FeeMultipliers",
                        "address": "52264967f262a7c55a2b570d3d2de409161521b8"
                    },
                    {
                        "acl_key": "vipercore/ReplayAttackBurnMultiplier",
                        "address": "52264967f262a7c55a2b570d3d2de409161521b8"
                    },
                    {
                        "acl_key": "pos/ProposerPercentage",
                        "address": "52264967f262a7c55a2b570d3d2de409161521b8"
                    },
                    {
                        "acl_key": "pos/AppAllocation",
                        "address": "52264967f262a7c55a2b570d3d2de409161521b8"
                    },
                    {
                        "acl_key": "vipercore/ClaimSubmissionWindow",
                        "address": "52264967f262a7c55a2b570d3d2de409161521b8"
                    },
                    {
                        "acl_key": "vipercore/MinimumNumberOfProofs",
                        "address": "52264967f262a7c55a2b570d3d2de409161521b8"
                    },
                    {
                        "acl_key": "vipercore/SessionNodeCount",
                        "address": "52264967f262a7c55a2b570d3d2de409161521b8"
                    },
                    {
                        "acl_key": "vipercore/SupportedBlockchains",
                        "address": "52264967f262a7c55a2b570d3d2de409161521b8"
                    },
                    {
                        "acl_key": "pos/BlocksPerSession",
                        "address": "52264967f262a7c55a2b570d3d2de409161521b8"
                    },
                    {
                        "acl_key": "pos/DAOAllocation",
                        "address": "52264967f262a7c55a2b570d3d2de409161521b8"
                    },
                    {
                        "acl_key": "pos/DowntimeJailDuration",
                        "address": "52264967f262a7c55a2b570d3d2de409161521b8"
                    },
                    {
                        "acl_key": "pos/MaxEvidenceAge",
                        "address": "52264967f262a7c55a2b570d3d2de409161521b8"
                    },
                    {
                        "acl_key": "pos/MaximumChains",
                        "address": "52264967f262a7c55a2b570d3d2de409161521b8"
                    },
                    {
                        "acl_key": "pos/MaxJailedBlocks",
                        "address": "52264967f262a7c55a2b570d3d2de409161521b8"
                    },
                    {
                        "acl_key": "pos/MaxValidators",
                        "address": "52264967f262a7c55a2b570d3d2de409161521b8"
                    },
                    {
                        "acl_key": "pos/MinSignedPerWindow",
                        "address": "52264967f262a7c55a2b570d3d2de409161521b8"
                    },
                    {
                        "acl_key": "pos/TokenRewardFactor",
                        "address": "52264967f262a7c55a2b570d3d2de409161521b8"
                    },
                    {
                        "acl_key": "pos/SignedBlocksWindow",
                        "address": "52264967f262a7c55a2b570d3d2de409161521b8"
                    },
                    {
                        "acl_key": "pos/SlashFractionDoubleSign",
                        "address": "52264967f262a7c55a2b570d3d2de409161521b8"
                    },
                    {
                        "acl_key": "pos/SlashFractionDowntime",
                        "address": "52264967f262a7c55a2b570d3d2de409161521b8"
                    },
                    {
                        "acl_key": "pos/StakeDenom",
                        "address": "52264967f262a7c55a2b570d3d2de409161521b8"
                    },
                    {
                        "acl_key": "pos/StakeMinimum",
                        "address": "52264967f262a7c55a2b570d3d2de409161521b8"
                    },
                    {
                        "acl_key": "pos/UnstakingTime",
                        "address": "52264967f262a7c55a2b570d3d2de409161521b8"
                    }
                ],
                "dao_owner": "52264967f262a7c55a2b570d3d2de409161521b8",
                "upgrade": {
                    "Height": "0",
                    "Version": "0"
                }
            },
            "DAO_Tokens": "6000000000000"
        }
    }
}`

var testnetGenesis = `{
		
}`

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
		apps.AppModuleBasic{},
		auth.AppModuleBasic{},
		gov.AppModuleBasic{},
		nodes.AppModuleBasic{},
		viper.AppModuleBasic{},
	).DefaultGenesis()
	// setup account genesis
	rawAuth := defaultGenesis[auth.ModuleName]
	var accountGenesis auth.GenesisState
	types.ModuleCdc.MustUnmarshalJSON(rawAuth, &accountGenesis)
	accountGenesis.Accounts = append(accountGenesis.Accounts, &auth.BaseAccount{
		Address: cb.GetAddress(),
		Coins:   sdk.NewCoins(sdk.NewCoin(sdk.DefaultStakeDenom, sdk.NewInt(1000000))),
		PubKey:  pubKey,
	})
	res := Codec().MustMarshalJSON(accountGenesis)
	defaultGenesis[auth.ModuleName] = res
	// set address as application too
	rawApps := defaultGenesis[appsTypes.ModuleName]
	var appsGenesis appsTypes.GenesisState
	types.ModuleCdc.MustUnmarshalJSON(rawApps, &appsGenesis)
	appsGenesis.Applications = append(appsGenesis.Applications, appsTypes.Application{
		Address:                 cb.GetAddress(),
		PublicKey:               cb.PublicKey,
		Jailed:                  false,
		Status:                  2,
		Chains:                  []string{sdk.PlaceholderHash},
		StakedTokens:            sdk.NewInt(10000000000000),
		MaxRelays:               sdk.NewInt(10000000000000),
		UnstakingCompletionTime: time.Time{},
	})
	res = Codec().MustMarshalJSON(appsGenesis)
	defaultGenesis[appsTypes.ModuleName] = res
	// set default governance in genesis
	rawViper := defaultGenesis[types.ModuleName]
	var viperGenesis types.GenesisState
	types.ModuleCdc.MustUnmarshalJSON(rawViper, &viperGenesis)
	viperGenesis.Params.SessionNodeCount = 1
	res = Codec().MustMarshalJSON(viperGenesis)
	defaultGenesis[types.ModuleName] = res
	// setup pos genesis
	rawPOS := defaultGenesis[nodesTypes.ModuleName]
	var posGenesisState nodesTypes.GenesisState
	types.ModuleCdc.MustUnmarshalJSON(rawPOS, &posGenesisState)
	posGenesisState.Validators = append(posGenesisState.Validators,
		nodesTypes.Validator{Address: sdk.Address(pubKey.Address()),
			PublicKey:    pubKey,
			Status:       sdk.Staked,
			Chains:       []string{sdk.PlaceholderHash},
			ServiceURL:   sdk.PlaceholderServiceURL,
			StakedTokens: sdk.NewInt(10000000)})
	res = types.ModuleCdc.MustMarshalJSON(posGenesisState)
	defaultGenesis[nodesTypes.ModuleName] = res
	// set default governance in genesis
	var govGenesisState govTypes.GenesisState
	rawGov := defaultGenesis[govTypes.ModuleName]
	Codec().MustUnmarshalJSON(rawGov, &govGenesisState)
	mACL := createDummyACL(pubKey)
	govGenesisState.Params.ACL = mACL
	govGenesisState.Params.DAOOwner = sdk.Address(pubKey.Address())
	govGenesisState.Params.Upgrade = govTypes.NewUpgrade(0, "0")
	res4 := Codec().MustMarshalJSON(govGenesisState)
	defaultGenesis[govTypes.ModuleName] = res4
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

func createDummyACL(kp crypto.PublicKey) govTypes.ACL {
	addr := sdk.Address(kp.Address())
	acl := govTypes.ACL{}
	acl = make([]govTypes.ACLPair, 0)
	acl.SetOwner("application/MinimumApplicationStake", addr)
	acl.SetOwner("application/AppUnstakingTime", addr)
	acl.SetOwner("application/BaseRelaysPerVIPR", addr)
	acl.SetOwner("application/MaxApplications", addr)
	acl.SetOwner("application/MaximumChains", addr)
	acl.SetOwner("application/ParticipationRate", addr)
	acl.SetOwner("application/StabilityModulation", addr)
	acl.SetOwner("auth/MaxMemoCharacters", addr)
	acl.SetOwner("auth/TxSigLimit", addr)
	acl.SetOwner("gov/acl", addr)
	acl.SetOwner("gov/daoOwner", addr)
	acl.SetOwner("gov/upgrade", addr)
	acl.SetOwner("vipercore/ClaimExpiration", addr)
	acl.SetOwner("auth/FeeMultipliers", addr)
	acl.SetOwner("vipercore/ReplayAttackBurnMultiplier", addr)
	acl.SetOwner("pos/ProposerPercentage", addr)
	acl.SetOwner("pos/AppAllocation", addr)
	acl.SetOwner("vipercore/ClaimSubmissionWindow", addr)
	acl.SetOwner("vipercore/MinimumNumberOfProofs", addr)
	acl.SetOwner("vipercore/SessionNodeCount", addr)
	acl.SetOwner("vipercore/SupportedBlockchains", addr)
	acl.SetOwner("pos/BlocksPerSession", addr)
	acl.SetOwner("pos/DAOAllocation", addr)
	acl.SetOwner("pos/AppAllocation", addr)
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
	return acl
}
