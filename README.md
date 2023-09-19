
# V I P E R  -  N E T W O R K


## The Problem:

Most node infrastructure providers in the Web3 space are completely centralized, contradicting Web3's core value of Decentralization and introducing potential central points of failure. This compromises reliability, security, and data integrity.

## Viper Network: An RPC relay protocol for Web3

Viper Network is building a Decentralised Physical Infrastructure Network (DePIN) protocol that provides a trustless RPC layer for Web3 applications to interact with blockchains in a decentralized way.

Imagine us as a decentralized counterpart to Alchemy, where we leverage a network of individual nodes to offer a more secure, cost-effective, and reliable RPC solution for web3 applications.

```bash
git clone https://github.com/vipernet-xyz/viper-network.git
cd viper-network
#build locally
go build cmd/main.go
```
Create account
```bash
./main accounts create
```
Set the account as a Validator using the address
```bash 
./main accounts set-validator <address>
```
Generate Chains
```bash
./main util generate-chains
```
Generate Geozone
```bash
./main utils generate-geozone
```
Start the node
```bash
./main start
```