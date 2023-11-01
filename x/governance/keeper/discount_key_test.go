package keeper

import (
	"testing"

	"github.com/stretchr/testify/assert"
	sdk "github.com/vipernet-xyz/viper-network/types"
)

// TestHasDiscountKey checks if HasDiscountKey correctly detects the existence of a discount key.
func TestHasDiscountKey(t *testing.T) {
	// Create a context for testing, you may need to set up the necessary dependencies and mocks.
	ctx, keeper := createTestKeeperAndContext(t, false)

	// Define a test address and discount key
	testAddress := sdk.Address([]byte("test_address"))
	discountKey := "test_discount_key"

	// Initially, the discount key should not exist
	exists := keeper.HasDiscountKey(ctx, testAddress)
	assert.False(t, exists)

	// Set the discount key
	err := keeper.SetDiscountKey(ctx, testAddress, discountKey)
	assert.NoError(t, err)

	// Now, the discount key should exist
	exists = keeper.HasDiscountKey(ctx, testAddress)
	assert.True(t, exists)
}

// TestSetDiscountKey checks if SetDiscountKey correctly sets a discount key for an address.
func TestSetDiscountKey(t *testing.T) {
	// Create a context for testing, you may need to set up the necessary dependencies and mocks.
	ctx, keeper := createTestKeeperAndContext(t, false)

	// Define a test address and discount key
	testAddress := sdk.Address([]byte("test_address"))
	discountKey := "test_discount_key"

	// Set the discount key
	err := keeper.SetDiscountKey(ctx, testAddress, discountKey)
	assert.NoError(t, err)

	// Try to set the same discount key again (it should return an error)
	err = keeper.SetDiscountKey(ctx, testAddress, discountKey)
	assert.Error(t, err)
}

// TestGetDiscountKey checks if GetDiscountKey correctly retrieves a discount key for an address.
func TestGetDiscountKey(t *testing.T) {
	// Create a context for testing, you may need to set up the necessary dependencies and mocks.
	ctx, keeper := createTestKeeperAndContext(t, false)

	// Define a test address and discount key
	testAddress := sdk.Address([]byte("test_address"))
	discountKey := "test_discount_key"

	// Initially, the discount key should not exist
	dk := keeper.GetDiscountKey(ctx, testAddress)
	assert.Equal(t, dk, "")

	// Set the discount key
	err := keeper.SetDiscountKey(ctx, testAddress, discountKey)
	assert.NoError(t, err)

	// Retrieve the discount key
	dk = keeper.GetDiscountKey(ctx, testAddress)
	assert.Equal(t, dk, discountKey)
}
