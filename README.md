
# V I P E R  -  N E T W O R K


## The Problem:

Most node infrastructure providers in the Web3 space are completely centralized, contradicting Web3's core value of Decentralization and introducing potential central points of failure. This compromises reliability, security, and data integrity.

## Viper Network: An RPC relay protocol for Web3

Viper Network is building a Decentralised Physical Infrastructure Network (DePIN) protocol that provides a trustless RPC layer for Web3 applications to interact with blockchains in a decentralized way.

Imagine us as a decentralized counterpart to Alchemy, where we leverage a network of individual nodes to offer a more secure, cost-effective, and reliable RPC solution for web3 applications.

## Installation

### Build from source

**Step 1. Install Golang**

Go version [1.18](https://go.dev/doc/go1.18) or higher is required.

If you haven't already, install Go by following the installation guide in [the official docs](https://golang.org/doc/install). Make sure that your `GOPATH` and `GOBIN` environment variables are properly set up.

**Step 2. Get source code**

Use `git` to retrieve Viper Network from [the official repository](https://github.com/vipernet-xyz/viper-network) and checkout latest release, which will install the `viper` binary.

## Get source code
```bash
git clone https://github.com/vipernet-xyz/viper-network.git
cd viper-network
# build locally
go build cmd/main.go
# copy binary to a standard path
sudo cp main /usr/local/bin/viper
```

**Step 3. Verify your installation**

Verify the version to see if you have installed `viper` correctly.

```bash
viper version
```

### CLI:

`viper` is the all-in-one command for operating and interacting with a running Viper network. To view various subcommands and their expected arguments, use the `$ viper --help` command:

```bash
    // // // // // // // // // // // // // // // // // 
                           V I P E R  N E T W O R K
                // // // // // // // // // // // // // // // // //

Usage:
  viper [command]

Available Commands:
  accounts     account management
  clients      client management
  completion   Generate the autocompletion script for the specified shell
  governance   governance management
  help         Help about any command
  ibc-transfer IBC-Transfer
  query        query the blockchain
  reset        Reset viper-network
  servicers    servicer management
  start        starts viper-network daemon
  stop         Stop viper-network
  util         utility functions
  version      Get current version

Flags:
      --datadir string            data directory (default is $HOME/.github.com/vipernet-xyz/viper-network/
  -h, --help                      help for viper
      --persistent_peers string   a comma separated list of PeerURLs: '<ID>@<IP>:<PORT>,<ID2>@<IP2>:<PORT>...<IDn>@<IPn>:<PORT>'
      --remoteCLIURL string       takes a remote endpoint in the form of <protocol>://<host> (uses RPC Port)
      --seeds string              a comma separated list of PeerURLs: '<ID>@<IP>:<PORT>,<ID2>@<IP2>:<PORT>...<IDn>@<IPn>:<PORT>'
      --servicer string           takes a remote endpoint in the form <protocol>://<host>:<port>

Use "viper [command] --help" for more information about a command.
```
## How to Run a Node?

**Step 1. Create account:**
```bash
viper accounts create
```
**Step 2. Set the account as a Validator using the address:**
```bash 
viper accounts set-validator <address>
```
**Step 3. Generate Chains:**
```bash
viper util generate-chains
```
**Step 4. Generate Geozone:**
```bash
viper utils generate-geozone
```
**Step 5. Create genesis.json:**
```bash
cd ~/.viper/config

#use the below sample genesis code with address and public key replaced by the once generated locally

{
    "genesis_time": "2023-09-19T12:19:57.265745Z",
    "chain_id": "viper-test",
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
        "transfer": {
            "port_id": "transfer",
            "denom_traces": [],
            "params": {
                "send_enabled": true,
                "receive_enabled": true
            }
        },
        "vipernet": {
            "params": {
                "proof_waiting_period": "3",
                "supported_blockchains": [
                    "0001"
                ],
                "claim_expiration": "24",
                "replay_attack_burn_multiplier": "3",
                "minimum_number_of_proofs": "1000",
                "supported_geo_zones": null,
                "minimum_sample_relays": "100"
            },
            "claims": null
        },
        "provider": {
            "params": {
                "unstaking_time": "1814400000000000",
                "max_providers": "9223372036854775807",
                "minimum_provider_stake": "0",
                "base_relays_per_vip": "200000",
                "stability_modulation": "0",
                "participation_rate_on": false,
                "maximum_chains": "15",
                "minimum_number_servicers": "3",
                "maximum_number_servicers": "25",
                "maximum_free_tier_relays_per_session": "5000"
            },
            "providers": [
                {
                    "address": "a583e80d51ab63f41a46fa4d233057905044a7c2",
                    "public_key": "d10a6c474275c0e591723f57366e349620b1bc899808d8bfd0be1f5d3e4baf07",
                    "jailed": false,
                    "chains": [
                        "0001"
                    ],
                    "max_relays": "10000000000000",
                    "status": 2,
                    "staked_tokens": "10000000000000",
                    "geo_zones": null,
                    "num_servicers": 0,
                    "unstaking_time": "0001-01-01T00:00:00Z"
                }
            ],
            "exported": false
        },
        "authentication": {
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
                        "address": "a583e80d51ab63f41a46fa4d233057905044a7c2",
                        "coins": [
                            {
                                "denom": "uvipr",
                                "amount": "100000000000"
                            }
                        ],
                        "public_key": {
                            "type": "crypto/ed25519_public_key",
                            "value": "d10a6c474275c0e591723f57366e349620b1bc899808d8bfd0be1f5d3e4baf07"
                        }
                    }
                }
            ],
            "supply": []
        },
        "capability": {
            "index": "1",
            "owners": []
        },
        "ibc": {
            "client_genesis": {
                "clients": [],
                "clients_consensus": [],
                "clients_metadata": null,
                "params": {
                    "allowed_clients": [
                        "06-solomachine",
                        "07-tendermint"
                    ]
                }
            },
            "connection_genesis": {
                "connections": [],
                "client_connection_paths": [],
                "params": {
                    "max_expected_time_per_block": "30000000000"
                }
            },
            "channel_genesis": {
                "channels": [],
                "acknowledgements": [],
                "commitments": [],
                "receipts": [],
                "send_sequences": [],
                "recv_sequences": [],
                "ack_sequences": []
            }
        },
        "governance": {
            "params": {
                "acl": [
                    {
                        "acl_key": "provider/MinimumProviderStake",
                        "address": "a583e80d51ab63f41a46fa4d233057905044a7c2"
                    },
                    {
                        "acl_key": "provider/ProviderUnstakingTime",
                        "address": "a583e80d51ab63f41a46fa4d233057905044a7c2"
                    },
                    {
                        "acl_key": "provider/BaseRelaysPerVIPR",
                        "address": "a583e80d51ab63f41a46fa4d233057905044a7c2"
                    },
                    {
                        "acl_key": "provider/MaxProviders",
                        "address": "a583e80d51ab63f41a46fa4d233057905044a7c2"
                    },
                    {
                        "acl_key": "provider/MaximumChains",
                        "address": "a583e80d51ab63f41a46fa4d233057905044a7c2"
                    },
                    {
                        "acl_key": "provider/ParticipationRate",
                        "address": "a583e80d51ab63f41a46fa4d233057905044a7c2"
                    },
                    {
                        "acl_key": "provider/StabilityModulation",
                        "address": "a583e80d51ab63f41a46fa4d233057905044a7c2"
                    },
                    {
                        "acl_key": "provider/MinNumServicers",
                        "address": "a583e80d51ab63f41a46fa4d233057905044a7c2"
                    },
                    {
                        "acl_key": "provider/MaxNumServicers",
                        "address": "a583e80d51ab63f41a46fa4d233057905044a7c2"
                    },
                    {
                        "acl_key": "provider/MaxFreeTierRelaysPerSession",
                        "address": "a583e80d51ab63f41a46fa4d233057905044a7c2"
                    },
                    {
                        "acl_key": "authentication/MaxMemoCharacters",
                        "address": "a583e80d51ab63f41a46fa4d233057905044a7c2"
                    },
                    {
                        "acl_key": "authentication/TxSigLimit",
                        "address": "a583e80d51ab63f41a46fa4d233057905044a7c2"
                    },
                    {
                        "acl_key": "authentication/FeeMultipliers",
                        "address": "a583e80d51ab63f41a46fa4d233057905044a7c2"
                    },
                    {
                        "acl_key": "governance/acl",
                        "address": "a583e80d51ab63f41a46fa4d233057905044a7c2"
                    },
                    {
                        "acl_key": "governance/daoOwner",
                        "address": "a583e80d51ab63f41a46fa4d233057905044a7c2"
                    },
                    {
                        "acl_key": "governance/upgrade",
                        "address": "a583e80d51ab63f41a46fa4d233057905044a7c2"
                    },
                    {
                        "acl_key": "vipernet/ClaimExpiration",
                        "address": "a583e80d51ab63f41a46fa4d233057905044a7c2"
                    },
                    {
                        "acl_key": "vipernet/ReplayAttackBurnMultiplier",
                        "address": "a583e80d51ab63f41a46fa4d233057905044a7c2"
                    },
                    {
                        "acl_key": "vipernet/ClaimSubmissionWindow",
                        "address": "a583e80d51ab63f41a46fa4d233057905044a7c2"
                    },
                    {
                        "acl_key": "vipernet/MinimumNumberOfProofs",
                        "address": "a583e80d51ab63f41a46fa4d233057905044a7c2"
                    },
                    {
                        "acl_key": "vipernet/SupportedBlockchains",
                        "address": "a583e80d51ab63f41a46fa4d233057905044a7c2"
                    },
                    {
                        "acl_key": "vipernet/SupportedGeoZones",
                        "address": "a583e80d51ab63f41a46fa4d233057905044a7c2"
                    },
                    {
                        "acl_key": "vipernet/MinimumSampleRelays",
                        "address": "a583e80d51ab63f41a46fa4d233057905044a7c2"
                    },
                    {
                        "acl_key": "pos/BlocksPerSession",
                        "address": "a583e80d51ab63f41a46fa4d233057905044a7c2"
                    },
                    {
                        "acl_key": "pos/DAOAllocation",
                        "address": "a583e80d51ab63f41a46fa4d233057905044a7c2"
                    },
                    {
                        "acl_key": "pos/ProviderAllocation",
                        "address": "a583e80d51ab63f41a46fa4d233057905044a7c2"
                    },
                    {
                        "acl_key": "pos/DowntimeJailDuration",
                        "address": "a583e80d51ab63f41a46fa4d233057905044a7c2"
                    },
                    {
                        "acl_key": "pos/MaxEvidenceAge",
                        "address": "a583e80d51ab63f41a46fa4d233057905044a7c2"
                    },
                    {
                        "acl_key": "pos/MaximumChains",
                        "address": "a583e80d51ab63f41a46fa4d233057905044a7c2"
                    },
                    {
                        "acl_key": "pos/MaxJailedBlocks",
                        "address": "a583e80d51ab63f41a46fa4d233057905044a7c2"
                    },
                    {
                        "acl_key": "pos/MaxValidators",
                        "address": "a583e80d51ab63f41a46fa4d233057905044a7c2"
                    },
                    {
                        "acl_key": "pos/MinSignedPerWindow",
                        "address": "a583e80d51ab63f41a46fa4d233057905044a7c2"
                    },
                    {
                        "acl_key": "pos/TokenRewardFactor",
                        "address": "a583e80d51ab63f41a46fa4d233057905044a7c2"
                    },
                    {
                        "acl_key": "pos/SignedBlocksWindow",
                        "address": "a583e80d51ab63f41a46fa4d233057905044a7c2"
                    },
                    {
                        "acl_key": "pos/SlashFractionDoubleSign",
                        "address": "a583e80d51ab63f41a46fa4d233057905044a7c2"
                    },
                    {
                        "acl_key": "pos/SlashFractionDowntime",
                        "address": "a583e80d51ab63f41a46fa4d233057905044a7c2"
                    },
                    {
                        "acl_key": "pos/StakeDenom",
                        "address": "a583e80d51ab63f41a46fa4d233057905044a7c2"
                    },
                    {
                        "acl_key": "pos/StakeMinimum",
                        "address": "a583e80d51ab63f41a46fa4d233057905044a7c2"
                    },
                    {
                        "acl_key": "pos/UnstakingTime",
                        "address": "a583e80d51ab63f41a46fa4d233057905044a7c2"
                    },
                    {
                        "acl_key": "pos/ServicerCountLock",
                        "address": "a583e80d51ab63f41a46fa4d233057905044a7c2"
                    },
                    {
                        "acl_key": "pos/BurnActive",
                        "address": "a583e80d51ab63f41a46fa4d233057905044a7c2"
                    },
                    {
                        "acl_key": "pos/MinPauseTime",
                        "address": "a583e80d51ab63f41a46fa4d233057905044a7c2"
                    },
                    {
                        "acl_key": "pos/MaxFishermen",
                        "address": "a583e80d51ab63f41a46fa4d233057905044a7c2"
                    },
                    {
                        "acl_key": "pos/FishermenCount",
                        "address": "a583e80d51ab63f41a46fa4d233057905044a7c2"
                    },
                    {
                        "acl_key": "pos/SlashFractionNoActivity",
                        "address": "a583e80d51ab63f41a46fa4d233057905044a7c2"
                    },
                    {
                        "acl_key": "pos/ProposerPercentage",
                        "address": "a583e80d51ab63f41a46fa4d233057905044a7c2"
                    }
                ],
                "dao_owner": "a583e80d51ab63f41a46fa4d233057905044a7c2",
                "upgrade": {
                    "Height": "0",
                    "Version": "0"
                }
            },
            "DAO_Tokens": "100000000000"
        },
        "pos": {
            "params": {
                "relays_to_tokens_multiplier": "1000",
                "unstaking_time": "1814400000000000",
                "max_validators": "100",
                "stake_denom": "uvipr",
                "stake_minimum": "10000000000",
                "session_block_frequency": "4",
                "dao_allocation": "10",
                "provider_allocation": "5",
                "proposer_allocation": "5",
                "maximum_chains": "15",
                "max_jailed_blocks": "2000",
                "max_evidence_age": "120000000000",
                "signed_blocks_window": "10",
                "min_signed_per_window": "0.500000000000000000",
                "downtime_jail_duration": "600000000000",
                "slash_fraction_double_sign": "0.000001000000000000",
                "slash_fraction_downtime": "0.000001000000000000",
                "servicer_count_lock": false,
                "burn_active": false,
                "min_pause_time": "600000000000",
                "max_fishermen": "50",
                "fishermen_count": "1",
                "slash_fraction_noactivity": "0.000001000000000000"
            },
            "prevState_total_power": "0",
            "prevState_validator_powers": null,
            "validators": [
                {
                    "address": "a583e80d51ab63f41a46fa4d233057905044a7c2",
                    "public_key": "d10a6c474275c0e591723f57366e349620b1bc899808d8bfd0be1f5d3e4baf07",
                    "jailed": false,
                    "paused": false,
                    "status": 2,
                    "chains": [
                        "0001"
                    ],
                    "service_url": "http://127.0.0.1:8081",
                    "tokens": "100000000000",
                    "geo_zone": null,
                    "unstaking_time": "0001-01-01T00:00:00Z",
                    "output_address": "",
                    "ReportCard": {
                        "total_sessions": 0,
                        "total_latency_score": "0",
                        "total_availability_score": "0",
                        "total_reliability_score": "0"
                    }
                }
            ],
            "exported": false,
            "signing_infos": {},
            "missed_blocks": {},
            "previous_proposer": ""
        }
    }
}
```
**Step 6. Start the Node:**
```bash
viper start
```