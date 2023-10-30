package keeper

import (
	"reflect"
	"testing"

	"github.com/vipernet-xyz/viper-network/x/providers/types"
)

func TestProviderUtil_Provider(t *testing.T) {
	stakedProvider := getStakedProvider()

	type args struct {
		provider types.Provider
	}
	type want struct {
		provider types.Provider
	}
	tests := []struct {
		name string
		find bool
		args
		want
	}{
		{
			name: "gets provider",
			find: false,
			args: args{provider: stakedProvider},
			want: want{provider: stakedProvider},
		},
		{
			name: "errors if no provider",
			find: true,
			args: args{provider: stakedProvider},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			context, _, keeper := createTestInput(t, true)
			switch tt.find {
			case true:
				if got := keeper.Provider(context, tt.args.provider.Address); got != nil {
					t.Errorf("keeperProviderUtil.Provider()= %v, want nil", got)
				}
			default:
				keeper.SetProvider(context, tt.args.provider)
				keeper.SetStakedProvider(context, tt.args.provider)
				if got := keeper.Provider(context, tt.args.provider.Address); !reflect.DeepEqual(got, tt.want.provider) {
					t.Errorf("keeperProviderUtil.Provider()= %v, want %v", got, tt.want.provider)
				}
			}
		})
	}

}

//func TestNewProviderCaching(t *testing.T) { todo
//	stakedProvider := getStakedProvider()
//
//	type args struct {
//		bz        []byte
//		provider types.Provider
//	}
//	type expected struct {
//		provider types.Provider
//		message   string
//		length    int
//	}
//	tests := []struct {
//		name   string
//		errors bool
//		args
//		expected
//	}{
//		{
//			name:     "getPrevStatePowerMap",
//			errors:   false,
//			args:     args{provider: stakedProvider},
//			expected: expected{provider: stakedProvider, length: 1},
//		},
//	}
//
//	for _, test := range tests {
//		t.Run(test.name, func(t *testing.T) {
//			context, _, keeper := createTestInput(t, true)
//			keeper.SetProvider(context, test.args.provider)
//			keeper.SetStakedProvider(context, test.args.provider)
//			store := context.KVStore(keeper.storeKey)
//			key := types.KeyForProviderPrevStateStateByPower(test.args.provider.Address)
//			store.Set(key, test.args.provider.Address)
//			powermap := keeper.getPrevStatePowerMap(context)
//			assert.Len(t, powermap, test.expected.length, "does not have correct length")
//			var valAddr [sdk.AddrLen]byte
//			copy(valAddr[:], key[1:])
//
//			for mapKey, value := range powermap {
//				assert.Equal(t, valAddr, mapKey, "key is not correct")
//				bz := make([]byte, len(test.args.provider.Address))
//				copy(bz, test.args.provider.Address)
//				assert.Equal(t, bz, value, "key is not correct")
//			}
//		})
//	}
//}

func TestProviderUtil_ProviderCaching(t *testing.T) {
	stakedProvider := getStakedProvider()

	type args struct {
		provider types.Provider
	}
	tests := []struct {
		name   string
		panics bool
		args
		want types.Provider
	}{
		{
			name: "gets provider",
			args: args{provider: stakedProvider},
			want: stakedProvider,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			context, _, keeper := createTestInput(t, true)
			keeper.SetProvider(context, tt.args.provider)
			keeper.SetStakedProvider(context, tt.args.provider)
			if got, _ := keeper.ProviderCache.Get(tt.args.provider.Address.String()); !got.(types.Provider).Equals(tt.want) {
				t.Errorf("keeperProviderUtil.ProviderCaching()= %v, want %v", got, tt.want)
			}
		})
	}
}
