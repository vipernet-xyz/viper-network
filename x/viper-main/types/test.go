package types

import (
	"encoding/hex"
	"encoding/json"
	"fmt"
	"log"

	sdk "github.com/vipernet-xyz/viper-network/types"
)

type Test interface {
	Hash() []byte                                             // returns cryptographic hash of bz
	Bytes() []byte                                            // returns bytes representation
	HashString() string                                       // returns the hex string representation of the merkleHash
	ValidateBasic() sdk.Error                                 // storeless validation check for the object
	GetSigner() sdk.Address                                   // returns the main signer(s) for the proof (used in messages)
	Store(sessionHeader SessionHeader, storage *CacheStorage) // handle the proof after validation
	ToProto() TestI                                           // convert to protobuf
}

type Tests []Test

type TestIs []TestI

func (ts Tests) ToTestI() (res []TestI) {
	for _, test := range ts {
		res = append(res, test.ToProto())
	}
	return
}

func (ti TestI) FromProto() Test {
	switch x := ti.Test.(type) {
	case *TestI_TestResult:
		return x.TestResult
	default:
		fmt.Println(fmt.Sprintf("invalid type assertion of testI: %T", x))
		return TestResult{}
	}
}

func (ts TestIs) FromTestI() (res Tests) {
	for _, test := range ts {
		res = append(res, test.FromProto())
	}
	return
}

var _ Test = TestResult{}

// "ValidateLocal" - Validates the proof object, where the owner of the proof is the local node
func (tr TestResult) ValidateLocal(verifyAddr sdk.Address) sdk.Error {
	//Basic Validations
	err := tr.ValidateBasic()
	if err != nil {
		return err
	}
	// validate the public key correctness
	if !(tr.ServicerAddress).Equals(verifyAddr) {
		return NewInvalidNodePubKeyError(ModuleName) // the public key is not this nodes, so they would not get paid
	}
	return nil
}

func (tr TestResult) ValidateBasic() sdk.Error {
	// Validate the ServicerAddress
	if tr.ServicerAddress.String() == "" {
		return NewEmptyAddressError(ModuleName)
	}

	// Validate the Timestamp. You can decide the range of acceptable timestamps if needed.
	if tr.IsAvailable && tr.Timestamp.IsZero() {
		return NewZeroTimeError(ModuleName)
	}

	// If the minimum latency is a non-zero duration, validate that the Latency is positive and within acceptable range.
	if tr.IsAvailable && tr.Latency <= 0 {
		return NewNegativeLatency(ModuleName)
	}

	return nil
}

func (tr TestResult) ToProto() TestI {
	return TestI{Test: &TestI_TestResult{TestResult: &tr}}
}

func (tr TestResult) Bytes() []byte {
	res, err := json.Marshal(TestResult{
		ServicerAddress: tr.ServicerAddress,
		Timestamp:       tr.Timestamp,
		Latency:         tr.Latency,
		IsAvailable:     tr.IsAvailable,
		IsReliable:      tr.IsReliable,
	})
	if err != nil {
		log.Fatal(fmt.Errorf("an error occured converting the test result to bytes:\n%v", err).Error())
	}
	return res
}

// "Hash" - Returns the cryptographic merkleHash of the rp bytes
func (tr TestResult) Hash() []byte {
	res := tr.Bytes()
	return Hash(res)
}

// "HashString" - Returns the hex encoded string of the rp merkleHash
func (tr TestResult) HashString() string {
	return hex.EncodeToString(tr.Hash())
}

// "Store" - Handles the test result object by adding it to the cache
func (tr TestResult) Store(sessionHeader SessionHeader, testStore *CacheStorage) {
	SetTestResult(sessionHeader, FishermanTestEvidence, tr, tr.ServicerAddress, testStore)
}

func (tr TestResult) GetSigner() sdk.Address {
	return tr.ServicerAddress
}
