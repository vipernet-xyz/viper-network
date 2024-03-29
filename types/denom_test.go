package types

import (
	"testing"

	"github.com/stretchr/testify/require"
)

var (
	atom  = "atom"  // 1 (base denom unit)
	matom = "matom" // 10^-3 (milli)
	uvipr = "uvipr" // 10^-6 (micro)
	natom = "natom" // 10^-9 (nano)
)

func TestRegisterDenom(t *testing.T) {
	atomUnit := OneDec() // 1 (base denom unit)

	require.NoError(t, RegisterDenom(atom, atomUnit))
	require.Error(t, RegisterDenom(atom, atomUnit))

	res, ok := GetDenomUnit(atom)
	require.True(t, ok)
	require.Equal(t, atomUnit, res)

	res, ok = GetDenomUnit(matom)
	require.False(t, ok)
	require.Equal(t, ZeroDec(), res)

	// reset registration
	denomUnits = map[string]BigDec{}
}

func TestConvertCoins(t *testing.T) {
	atomUnit := OneDec() // 1 (base denom unit)
	require.NoError(t, RegisterDenom(atom, atomUnit))

	matomUnit := NewDecWithPrec(1, 3) // 10^-3 (milli)
	require.NoError(t, RegisterDenom(matom, matomUnit))

	uviprUnit := NewDecWithPrec(1, 6) // 10^-6 (micro)
	require.NoError(t, RegisterDenom(uvipr, uviprUnit))

	natomUnit := NewDecWithPrec(1, 9) // 10^-9 (nano)
	require.NoError(t, RegisterDenom(natom, natomUnit))

	testCases := []struct {
		input  Coin
		denom  string
		result Coin
		expErr bool
	}{
		{NewCoin("foo", ZeroInt()), atom, Coin{}, true},
		{NewCoin(atom, ZeroInt()), "foo", Coin{}, true},
		{NewCoin(atom, ZeroInt()), "FOO", Coin{}, true},

		{NewCoin(atom, NewInt(5)), matom, NewCoin(matom, NewInt(5000)), false},       // atom => matom
		{NewCoin(atom, NewInt(5)), uvipr, NewCoin(uvipr, NewInt(5000000)), false},    // atom => uvipr
		{NewCoin(atom, NewInt(5)), natom, NewCoin(natom, NewInt(5000000000)), false}, // atom => natom

		{NewCoin(uvipr, NewInt(5000000)), matom, NewCoin(matom, NewInt(5000)), false},       // uvipr => matom
		{NewCoin(uvipr, NewInt(5000000)), natom, NewCoin(natom, NewInt(5000000000)), false}, // uvipr => natom
		{NewCoin(uvipr, NewInt(5000000)), atom, NewCoin(atom, NewInt(5)), false},            // uvipr => atom

		{NewCoin(matom, NewInt(5000)), natom, NewCoin(natom, NewInt(5000000000)), false}, // matom => natom
		{NewCoin(matom, NewInt(5000)), uvipr, NewCoin(uvipr, NewInt(5000000)), false},    // matom => uvipr
	}

	for i, tc := range testCases {
		res, err := ConvertCoin(tc.input, tc.denom)
		require.Equal(
			t, tc.expErr, err != nil,
			"unexpected error; tc: #%d, input: %s, denom: %s", i+1, tc.input, tc.denom,
		)
		require.Equal(
			t, tc.result, res,
			"invalid result; tc: #%d, input: %s, denom: %s", i+1, tc.input, tc.denom,
		)
	}

	// reset registration
	denomUnits = map[string]BigDec{}
}
