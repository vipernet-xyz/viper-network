
# V I P E R  -  N E T W O R K


## The Problem:

Most node infrastructure providers in the Web3 space are completely centralized, contradicting Web3's core value of Decentralization and introducing potential central points of failure. This compromises reliability, security, and data integrity.

## Viper Network: An RPC relay protocol for Web3

Viper Network is building a Decentralised Physical Infrastructure Network (DePIN) protocol that provides a trustless RPC layer for Web3 applications to interact with blockchains in a decentralized way.

Imagine us as a decentralized counterpart to [Alchemy](https://www.alchemy.com/), where we leverage a network of individual nodes to offer a more secure, cost-effective, fast and reliable RPC solution for web3 applications.

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
## How to Run a Test-Node?

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
**Step 5. Start the Node:**
```bash
viper start
```

## Contact

<div>
  <a  href="https://twitter.com/viper_network_" ><img src="https://img.shields.io/twitter/url/http/shields.io.svg?style=social"></a>
  <a href="https://t.me/vishruthsk"><img src="https://img.shields.io/badge/Telegram-blue.svg"></a>
</div>


Engage and collaborate with our community on [Discord](https://discord.gg/eBDYH4Zxek)

Dive deeper into our vision on our [Website](https://vipernet.xyz/) or explore our detailed [Deck](https://www.dropbox.com/scl/fi/nuhh9ag7idaekxelf6rl9/Viper.pdf?rlkey=g6h3lpegyzp24rna4flhni43i&dl=0)
