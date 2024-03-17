package types

import (
	"testing"

	sdk "github.com/vipernet-xyz/viper-network/types"

	"github.com/stretchr/testify/assert"
)

func TestNewExpiredProofsSubmissionError(t *testing.T) {
	assert.Equal(t, NewExpiredProofsSubmissionError(ModuleName), sdk.NewError(ModuleName, CodeExpiredProofsSubmissionError, ExpiredProofsSubmissionError.Error()))
}

func TestNewMerkleNodeNotFoundError(t *testing.T) {
	assert.Equal(t, NewMerkleNodeNotFoundError(ModuleName), sdk.NewError(ModuleName, CodeMerkleNodeNotFoundError, MerkleNodeNotFoundError.Error()))
}

func TestNewEmptyMerkleTreeError(t *testing.T) {
	assert.Equal(t, NewEmptyMerkleTreeError(ModuleName), sdk.NewError(ModuleName, CodeEmptyMerkleTreeError, EmptyMerkleTreeError.Error()))
}

func TestNewInvalidMerkleVerifyError(t *testing.T) {
	assert.Equal(t, NewInvalidClaimMerkleVerifyError(ModuleName), sdk.NewError(ModuleName, CodeInvalidClaimMerkleVerifyError, InvalidClaimMerkleVerifyError.Error()))
}

func TestClaimNotFoundError(t *testing.T) {
	assert.Equal(t, NewClaimNotFoundError(ModuleName), sdk.NewError(ModuleName, CodeClaimNotFoundError, ClaimNotFoundError.Error()))
}

func TestCousinLeafEquivalentError(t *testing.T) {
	assert.Equal(t, NewCousinLeafEquivalentError(ModuleName), sdk.NewError(ModuleName, CodeCousinLeafEquivalentError, CousinLeafEquivalentError.Error()))
}

func TestInvalidLeafCousinProofsComboError(t *testing.T) {
	assert.Equal(t, NewInvalidLeafCousinProofsComboError(ModuleName), sdk.NewError(ModuleName, CodeInvalidLeafCousinProofsCombo, InvalidLeafCousinProofsCombo.Error()))
}

func TestInvalidRootError(t *testing.T) {
	assert.Equal(t, NewInvalidRootError(ModuleName), sdk.NewError(ModuleName, CodeInvalidRootError, InvalidRootError.Error()))
}

func TestInvalidHashLengthError(t *testing.T) {
	assert.Equal(t, NewInvalidHashLengthError(ModuleName), sdk.NewError(ModuleName, CodeInvalidHashLengthError, InvalidHashLengthError.Error()))
}

func TestInvalidRequestorPubKeyError(t *testing.T) {
	assert.Equal(t, NewInvalidRequestorPubKeyError(ModuleName), sdk.NewError(ModuleName, CodeInvalidRequestorPubKeyError, InvalidRequestorPubKeyError.Error()))
}
