package types

import (
	"encoding/hex"
	"encoding/json"
	"fmt"
	"log"

	sdk "github.com/vipernet-xyz/viper-network/types"
	providerexported "github.com/vipernet-xyz/viper-network/x/providers/exported"
	"github.com/vipernet-xyz/viper-network/x/servicers/exported"
)

// "Session" - The relationship between an provider and the viper network

func (s Session) IsSealed() bool {
	return false
}

func (s Session) Seal() CacheObject {
	return s
}

// "NewSession" - create a new session from seed data
func NewSession(sessionCtx, ctx sdk.Ctx, keeper PosKeeper, sessionHeader SessionHeader, blockHash string, sessionServicersCount int) (Session, sdk.Error) {
	// first generate session key
	sessionKey, err := NewSessionKey(sessionHeader.ProviderPubKey, sessionHeader.Chain, blockHash)
	if err != nil {
		return Session{}, err
	}
	// then generate the service servicers for that session
	sessionServicers, err := NewSessionServicers(sessionCtx, ctx, keeper, sessionHeader.Chain, sessionKey, sessionServicersCount)
	if err != nil {
		return Session{}, err
	}
	// then populate the structure and return
	return Session{
		SessionKey:       sessionKey,
		SessionHeader:    sessionHeader,
		SessionServicers: sessionServicers,
	}, nil
}

// "Validate" - Validates a session object
func (s Session) Validate(servicer sdk.Address, provider providerexported.ProviderI, sessionNodeCount int) sdk.Error {
	// validate chain
	if len(s.SessionHeader.Chain) == 0 {
		return NewEmptyNonNativeChainError(ModuleName)
	}
	// validate sessionBlockHeight
	if s.SessionHeader.SessionBlockHeight < 1 {
		return NewInvalidBlockHeightError(ModuleName)
	}
	// validate the provider public key
	if err := PubKeyVerification(s.SessionHeader.ProviderPubKey); err != nil {
		return err
	}
	// validate provider corresponds to providerPubKey
	if provider.GetPublicKey().RawString() != s.SessionHeader.ProviderPubKey {
		return NewInvalidProviderPubKeyError(ModuleName)
	}
	// validate provider chains
	chains := provider.GetChains()
	found := false
	for _, c := range chains {
		if c == s.SessionHeader.Chain {
			found = true
			break
		}
	}
	if !found {
		return NewUnsupportedBlockchainProviderError(ModuleName)
	}
	// validate sessionServicers
	err := s.SessionServicers.Validate(sessionNodeCount)
	if err != nil {
		return err
	}
	// validate servicer is of the session
	if !s.SessionServicers.Contains(servicer) {
		return NewInvalidSessionError(ModuleName)
	}
	return nil
}

var _ CacheObject = Session{} // satisfies the cache object interface

func (s Session) MarshalObject() ([]byte, error) {
	return ModuleCdc.ProtoMarshalBinaryBare(&s)
}

func (s Session) UnmarshalObject(b []byte) (CacheObject, error) {
	err := ModuleCdc.ProtoUnmarshalBinaryBare(b, &s)
	return s, err
}

func (s Session) Key() ([]byte, error) {
	return s.SessionHeader.Hash(), nil
}

// "SessionServicers" - Service servicers in a session
type SessionServicers []sdk.Address

// "NewSessionServicers" - Generates servicers for the session
func NewSessionServicers(sessionCtx, ctx sdk.Ctx, keeper PosKeeper, chain string, sessionKey SessionKey, sessionServicersCount int) (sessionServicers SessionServicers, err sdk.Error) {
	// all servicersAddrs at session genesis
	servicersAddrs, totalServicers := keeper.GetValidatorsByChain(sessionCtx, chain)
	// validate servicersAddrs
	if totalServicers < sessionServicersCount {
		return nil, NewInsufficientServicersError(ModuleName)
	}
	sessionServicers = make(SessionServicers, sessionServicersCount)
	var servicer exported.ValidatorI
	//unique address map to avoid re-checking a pseudorandomly selected servicer
	m := make(map[string]struct{})
	// only select the servicersAddrs if not jailed
	for i, numOfServicers := 0, 0; ; i++ {
		//if this is true we already checked all servicers we got on getValidatorsBychain
		if len(m) >= totalServicers {
			return nil, NewInsufficientServicersError(ModuleName)
		}
		// generate the random index
		index := PseudorandomSelection(sdk.NewInt(int64(totalServicers)), sessionKey)
		// merkleHash the session key to provide new entropy
		sessionKey = Hash(sessionKey)
		// get the servicer from the array
		n := servicersAddrs[index.Int64()]
		//if we already have seen this address we continue as it's either on the list or discarded
		if _, ok := m[n.String()]; ok {
			continue
		}
		//add the servicer address to the map
		m[n.String()] = struct{}{}

		// cross check the servicer from the `new` or `end` world state
		servicer = keeper.Validator(ctx, n)
		// if not found or jailed, don't add to session and continue
		if servicer == nil || servicer.IsJailed() || !NodeHasChain(chain, servicer) || sessionServicers.Contains(servicer.GetAddress()) {
			continue
		}
		// else add the servicer to the session
		sessionServicers[numOfServicers] = n
		// increment the number of servicersAddrs in the sessionServicers slice
		numOfServicers++
		// if maxing out the session count end loop
		if numOfServicers == sessionServicersCount {
			break
		}
	}
	// return the servicersAddrs
	return sessionServicers, nil
}

// "Validate" - Validates the session servicer object
func (sn SessionServicers) Validate(sessionServicersCount int) sdk.Error {
	if len(sn) < sessionServicersCount {
		return NewInsufficientServicersError(ModuleName)
	}
	for _, n := range sn {
		if n == nil {
			return NewEmptyAddressError(ModuleName)
		}
	}
	return nil
}

// "Contains" - Verifies if the session servicers contains the servicer using the address
func (sn SessionServicers) Contains(addr sdk.Address) bool {
	// if nil return
	if addr == nil {
		return false
	}
	// loop over the servicers
	for _, servicer := range sn {
		if servicer == nil {
			continue
		}
		if servicer.Equals(addr) {
			return true
		}
	}
	return false
}

// "SessionKey" - the merkleHash identifier of the session
type SessionKey []byte

// "sessionKey" - Used for custom json
type sessionKey struct {
	ProviderPublicKey string `json:"provider_pub_key"`
	NonNativeChain    string `json:"chain"`
	BlockHash         string `json:"blockchain"`
}

// "NewSessionKey" - generates the session key from metadata
func NewSessionKey(providerPubKey string, chain string, blockHash string) (SessionKey, sdk.Error) {
	// validate providerPubKey
	if err := PubKeyVerification(providerPubKey); err != nil {
		return nil, err
	}
	// validate chain
	if err := NetworkIdentifierVerification(chain); err != nil {
		return nil, NewEmptyChainError(ModuleName)
	}
	// validate block addr
	if err := HashVerification(blockHash); err != nil {
		return nil, err
	}
	// marshal into json
	seed, err := json.Marshal(sessionKey{
		ProviderPublicKey: providerPubKey,
		NonNativeChain:    chain,
		BlockHash:         blockHash,
	})
	if err != nil {
		return nil, NewJSONMarshalError(ModuleName, err)
	}
	// return the addr of the result
	return Hash(seed), nil
}

// "Validate" - Validates the session key
func (sk SessionKey) Validate() sdk.Error {
	return HashVerification(hex.EncodeToString(sk))
}

// "ValidateHeader" - Validates the header of the session
func (sh SessionHeader) ValidateHeader() sdk.Error {
	// check the provider public key for validity
	if err := PubKeyVerification(sh.ProviderPubKey); err != nil {
		return err
	}
	// verify the chain merkleHash
	if err := NetworkIdentifierVerification(sh.Chain); err != nil {
		return err
	}
	// verify the block height
	if sh.SessionBlockHeight < 1 {
		return NewInvalidBlockHeightError(ModuleName)
	}
	return nil
}

// "Hash" - The cryptographic merkleHash representation of the session header
func (sh SessionHeader) Hash() []byte {
	res := sh.Bytes()
	return Hash(res)
}

// "HashString" - The hex string representation of the merkleHash
func (sh SessionHeader) HashString() string {
	return hex.EncodeToString(sh.Hash())
}

// "Bytes" - The bytes representation of the session header
func (sh SessionHeader) Bytes() []byte {
	res, err := json.Marshal(sh)
	if err != nil {
		log.Fatal(fmt.Errorf("an error occured converting the session header into bytes:\n%v", err))
	}
	return res
}

// "BlockHash" - Returns the merkleHash from the ctx block header
func BlockHash(ctx sdk.Context) string {
	return hex.EncodeToString(ctx.BlockHeader().LastBlockId.Hash)
}

// "MaxPossibleRelays" - Returns the maximum possible amount of relays for an Provider on a sessions
func MaxPossibleRelays(provider providerexported.ProviderI, sessionNodeCount int64) sdk.BigInt {
	//GetMaxRelays Max value is bound to math.MaxUint64,
	//current worse case is 1 chain and 5 servicers per session with a result of 3689348814741910323 which can be used safely as int64
	return provider.GetMaxRelays().ToDec().Quo(sdk.NewDec(int64(len(provider.GetChains())))).Quo(sdk.NewDec(sessionNodeCount)).RoundInt()
}

// "NodeHashChain" - Returns whether or not the servicer has the relayChain
func NodeHasChain(chain string, servicer exported.ValidatorI) bool {
	hasChain := false
	for _, c := range servicer.GetChains() {
		if c == chain {
			hasChain = true
			break
		}
	}
	return hasChain
}
