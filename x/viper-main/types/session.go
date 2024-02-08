package types

import (
	"encoding/hex"
	"encoding/json"
	"fmt"
	"log"

	sdk "github.com/vipernet-xyz/viper-network/types"
	requestorexported "github.com/vipernet-xyz/viper-network/x/requestors/exported"
	"github.com/vipernet-xyz/viper-network/x/servicers/exported"
	servicerTypes "github.com/vipernet-xyz/viper-network/x/servicers/types"
)

// "Session" - The relationship between an application and the viper network

func (s Session) IsSealable() bool {
	return false
}

func (s Session) HashString() string {
	return s.HashString()
}

// "SessionNodes" - Service nodes in a session
type SessionNodes []sdk.Address

// "NewSession" - create a new session from seed data
func NewSession(sessionCtx, ctx sdk.Ctx, keeper PosKeeper, sessionHeader SessionHeader, blockHash string) (Session, sdk.Error) {
	// first generate session key
	sessionKey, err := NewSessionKey(sessionHeader.RequestorPubKey, sessionHeader.Chain, blockHash)
	if err != nil {
		return Session{}, err
	}
	// then generate the service servicers for that session
	sessionServicers, err := NewSessionServicers(sessionCtx, ctx, keeper, sessionHeader.Chain, sessionHeader.GeoZone, sessionKey, sessionHeader.NumServicers)
	if err != nil {
		return Session{}, err
	}
	sessionFishermenCount := keeper.FishermenCount(ctx)
	sessionFishermen, err := NewSessionFishermen(sessionCtx, ctx, keeper, sessionHeader.Chain, sessionKey, sessionFishermenCount)
	if err != nil {
		return Session{}, err
	}
	// then populate the structure and return
	return Session{
		SessionKey:       sessionKey,
		SessionHeader:    sessionHeader,
		SessionServicers: sessionServicers,
		SessionFishermen: sessionFishermen,
	}, nil
}

// "Validate" - Validates a session object
func (s Session) Validate(servicer sdk.Address, requestor requestorexported.RequestorI, sessionNodeCount int) sdk.Error {
	// validate chain
	if len(s.SessionHeader.Chain) == 0 {
		return NewEmptyNonNativeChainError(ModuleName)
	}
	if len(s.SessionHeader.GeoZone) == 0 {
		return NewEmptyGeoZoneError(ModuleName)
	}
	// validate sessionBlockHeight
	if s.SessionHeader.SessionBlockHeight < 1 {
		return NewInvalidBlockHeightError(ModuleName)
	}
	// validate the requestor public key
	if err := PubKeyVerification(s.SessionHeader.RequestorPubKey); err != nil {
		return err
	}
	// validate requestor corresponds to requestorPubKey
	if requestor.GetPublicKey().RawString() != s.SessionHeader.RequestorPubKey {
		return NewInvalidRequestorPubKeyError(ModuleName)
	}
	// validate requestor chains
	chains := requestor.GetChains()
	found := false
	for _, c := range chains {
		if c == s.SessionHeader.Chain {
			found = true
			break
		}
	}
	if !found {
		return NewUnsupportedBlockchainRequestorError(ModuleName)
	}

	geoZones := requestor.GetGeoZones()
	found1 := false
	for _, c := range geoZones {
		if c == s.SessionHeader.GeoZone {
			found1 = true
			break
		}
	}
	if !found1 {
		return NewUnsupportedGeoZoneRequestorError(ModuleName)
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

type SessionFishermen []sdk.Address

// NewSessionServicers - Generates servicers for the session based on both chain and geo zone
func NewSessionServicers(sessionCtx, ctx sdk.Ctx, keeper PosKeeper, chain, geoZone string, sessionKey SessionKey, sessionServicersCount int64) (sessionServicers SessionServicers, err sdk.Error) {
	// all servicersAddrs at session genesis based on the chain
	servicersByChain, _ := keeper.GetValidatorsByChain(sessionCtx, chain)

	// all servicersAddrs at session genesis based on the geo zone
	servicersByGeoZone, _ := keeper.GetValidatorsByGeoZone(sessionCtx, geoZone)

	// Filter validators that are present in both lists (matching chain and geo zone)
	validatorsInBoth := make([]sdk.Address, 0)
	for _, addrByChain := range servicersByChain {
		for _, addrByGeoZone := range servicersByGeoZone {
			if addrByChain.Equals(addrByGeoZone) {
				validatorsInBoth = append(validatorsInBoth, addrByChain)
				break // Break to avoid duplicates
			}
		}
	}
	// Validate that the number of servicers is sufficient
	if len(validatorsInBoth) < int(sessionServicersCount) {
		return nil, NewInsufficientServicersError(ModuleName)
	}

	sessionServicers = make(SessionServicers, sessionServicersCount)
	var servicer exported.ValidatorI

	// Map to store the performance scores of validators
	scoresMap := make(map[string]int64)

	// Populate the scoresMap with scores from the report cards
	for _, validatorAddr := range validatorsInBoth {
		validator, found := keeper.GetValidator(ctx, validatorAddr)
		if found && validator.ReportCard != (servicerTypes.ReportCard{}) {
			score := servicerTypes.ScoresToPower(validator.ReportCard)
			scoresMap[validatorAddr.String()] = score
		}
	}

	// Unique address map to avoid re-checking a pseudorandomly selected servicer
	m := make(map[string]struct{})
	// Only select the servicersAddrs if not jailed and contain both chain and geo zone
	for i, numOfServicers := 0, 0; i < len(validatorsInBoth) && numOfServicers < int(sessionServicersCount); i++ {
		// Generate the random index based on report card scores
		index := PseudorandomSelectionWithWeights(scoresMap, sessionKey)
		// MerkleHash the session key to provide new entropy
		sessionKey = Hash(sessionKey)
		// Get the servicer from the array
		n := validatorsInBoth[index.Int64()]
		// If we already have seen this address we continue as it's either on the list or discarded
		if _, ok := m[n.String()]; ok {
			continue
		}
		// Add the servicer address to the map
		m[n.String()] = struct{}{}

		// Cross check the servicer from the `new` or `end` world state
		servicer = keeper.Validator(ctx, n)
		// If not found or jailed, don't add to session and continue
		if servicer == nil || servicer.IsJailed() || servicer.IsPaused() || !NodeHasChain(chain, servicer) || !NodeHasGeoZone(geoZone, servicer) || sessionServicers.Contains(servicer.GetAddress()) {
			continue
		}
		// Else add the servicer to the session
		sessionServicers[numOfServicers] = n
		// Increment the number of servicers in the sessionServicers slice
		numOfServicers++
	}

	// Return the servicers
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

func (sf SessionFishermen) Contains(addr sdk.Address) bool {
	// if nil return
	if addr == nil {
		return false
	}
	for _, fisherman := range sf {
		if fisherman == nil {
			continue
		}
		if fisherman.Equals(addr) {
			return true
		}
	}
	return false
}

// "SessionKey" - the merkleHash identifier of the session
type SessionKey []byte

// "sessionKey" - Used for custom json
type sessionKey struct {
	RequestorPublicKey string `json:"requestor_pub_key"`
	NonNativeChain     string `json:"chain"`
	BlockHash          string `json:"blockchain"`
}

// "NewSessionKey" - generates the session key from metadata
func NewSessionKey(requestorPubKey string, chain string, blockHash string) (SessionKey, sdk.Error) {
	// validate requestorPubKey
	if err := PubKeyVerification(requestorPubKey); err != nil {
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
		RequestorPublicKey: requestorPubKey,
		NonNativeChain:     chain,
		BlockHash:          blockHash,
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
	// check the requestor public key for validity
	if err := PubKeyVerification(sh.RequestorPubKey); err != nil {
		return err
	}
	// verify the chain merkleHash
	if err := NetworkIdentifierVerification(sh.Chain); err != nil {
		return err
	}
	if err := GeoZoneIdentifierVerification(sh.GeoZone); err != nil {
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
func BlockHash(ctx sdk.Ctx) string {
	return hex.EncodeToString(ctx.BlockHeader().LastBlockId.Hash)
}

// "MaxPossibleRelays" - Returns the maximum possible amount of relays for an App on a sessions
func MaxPossibleRelays(app requestorexported.RequestorI, sessionNodeCount int64) sdk.BigInt {
	//GetMaxRelays Max value is bound to math.MaxUint64,
	//current worse case is 1 chain and 5 nodes per session with a result of 3689348814741910323 which can be used safely as int64
	return app.GetMaxRelays().ToDec().Quo(sdk.NewDec(int64(len(app.GetChains())))).Quo(sdk.NewDec(sessionNodeCount)).RoundInt()
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

// "NodeHashChain" - Returns whether or not the servicer has the relayChain
func NodeHasGeoZone(geoZone string, servicer exported.ValidatorI) bool {
	hasGeoZone := false
	for _, c := range servicer.GetGeoZone() {
		if string(c) == geoZone {
			hasGeoZone = true
			break
		}
	}
	return hasGeoZone
}

// "Contains" - Verifies if the session nodes contains the node using the address
func (sn SessionNodes) Contains(addr sdk.Address) bool {
	// if nil return
	if addr == nil {
		return false
	}
	// loop over the nodes
	for _, node := range sn {
		if node == nil {
			continue
		}
		if node.Equals(addr) {
			return true
		}
	}
	return false
}

func NewSessionFishermen(sessionCtx, ctx sdk.Ctx, keeper PosKeeper, chain string, sessionKey SessionKey, sessionFishermenCount int64) (sessionFishermen SessionFishermen, err sdk.Error) {
	// Get all validators for the specified chain
	servicersByChain, _ := keeper.GetValidatorsByChain(sessionCtx, chain)

	// Validate that the number of fishermen is sufficient
	if len(servicersByChain) < int(sessionFishermenCount) {
		return nil, NewInsufficientServicersError(ModuleName)
	}

	sessionFishermen = make(SessionFishermen, sessionFishermenCount)
	var fisherman exported.ValidatorI

	// Unique address map to avoid re-checking a pseudorandomly selected fisherman
	m := make(map[string]struct{})
	// Only select the fishermen if not jailed and serve the chain
	for i, numOfFishermen := 0, 0; ; i++ {
		// If this is true we already checked all validators we got on GetValidatorsByChain
		if len(m) >= len(servicersByChain) {
			return nil, NewInsufficientServicersError(ModuleName)
		}

		// Generate the random index
		index := PseudorandomSelection(sdk.NewInt(int64(len(servicersByChain))), sessionKey)
		// MerkleHash the session key to provide new entropy
		sessionKey = Hash(sessionKey)
		// Get the fisherman from the array
		n := servicersByChain[index.Int64()]
		// If we already have seen this address we continue as it's either on the list or discarded
		if _, ok := m[n.String()]; ok {
			continue
		}
		// Add the fisherman address to the map
		m[n.String()] = struct{}{}

		// Cross check the fisherman from the `new` or `end` world state
		fisherman = keeper.Validator(ctx, n)
		// If not found or jailed, don't add to session and continue
		if fisherman == nil || fisherman.IsJailed() || fisherman.IsPaused() || sessionFishermen.Contains(fisherman.GetAddress()) {
			continue
		}
		// Else add the fisherman to the session
		sessionFishermen[numOfFishermen] = n
		// Increment the number of fishermen in the sessionFishermen slice
		numOfFishermen++
		// If maxing out the session count, end the loop
		if numOfFishermen == int(sessionFishermenCount) {
			break
		}
	}

	// Return the fishermen
	return sessionFishermen, nil
}

func FishermanInList(fisherman sdk.Address, sessionFishermen SessionFishermen) bool {
	for _, v := range sessionFishermen {
		if v.Equals(fisherman) {
			return true
		}
	}
	return false
}
