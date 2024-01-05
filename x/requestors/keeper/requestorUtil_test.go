package keeper

import (
	"reflect"
	"testing"

	"github.com/vipernet-xyz/viper-network/x/requestors/types"
)

func TestRequestorUtil_Requestor(t *testing.T) {
	stakedRequestor := getStakedRequestor()

	type args struct {
		requestor types.Requestor
	}
	type want struct {
		requestor types.Requestor
	}
	tests := []struct {
		name string
		find bool
		args
		want
	}{
		{
			name: "gets requestor",
			find: false,
			args: args{requestor: stakedRequestor},
			want: want{requestor: stakedRequestor},
		},
		{
			name: "errors if no requestor",
			find: true,
			args: args{requestor: stakedRequestor},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			context, _, keeper := createTestInput(t, true)
			switch tt.find {
			case true:
				if got := keeper.Requestor(context, tt.args.requestor.Address); got != nil {
					t.Errorf("keeperRequestorUtil.Requestor()= %v, want nil", got)
				}
			default:
				keeper.SetRequestor(context, tt.args.requestor)
				keeper.SetStakedRequestor(context, tt.args.requestor)
				if got := keeper.Requestor(context, tt.args.requestor.Address); !reflect.DeepEqual(got, tt.want.requestor) {
					t.Errorf("keeperRequestorUtil.Requestor()= %v, want %v", got, tt.want.requestor)
				}
			}
		})
	}

}

//func TestNewRequestorCaching(t *testing.T) { todo
//	stakedRequestor := getStakedRequestor()
//
//	type args struct {
//		bz        []byte
//		requestor types.Requestor
//	}
//	type expected struct {
//		requestor types.Requestor
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
//			args:     args{requestor: stakedRequestor},
//			expected: expected{requestor: stakedRequestor, length: 1},
//		},
//	}
//
//	for _, test := range tests {
//		t.Run(test.name, func(t *testing.T) {
//			context, _, keeper := createTestInput(t, true)
//			keeper.SetRequestor(context, test.args.requestor)
//			keeper.SetStakedRequestor(context, test.args.requestor)
//			store := context.KVStore(keeper.storeKey)
//			key := types.KeyForRequestorPrevStateStateByPower(test.args.requestor.Address)
//			store.Set(key, test.args.requestor.Address)
//			powermap := keeper.getPrevStatePowerMap(context)
//			assert.Len(t, powermap, test.expected.length, "does not have correct length")
//			var valAddr [sdk.AddrLen]byte
//			copy(valAddr[:], key[1:])
//
//			for mapKey, value := range powermap {
//				assert.Equal(t, valAddr, mapKey, "key is not correct")
//				bz := make([]byte, len(test.args.requestor.Address))
//				copy(bz, test.args.requestor.Address)
//				assert.Equal(t, bz, value, "key is not correct")
//			}
//		})
//	}
//}

func TestRequestorUtil_RequestorCaching(t *testing.T) {
	stakedRequestor := getStakedRequestor()

	type args struct {
		requestor types.Requestor
	}
	tests := []struct {
		name   string
		panics bool
		args
		want types.Requestor
	}{
		{
			name: "gets requestor",
			args: args{requestor: stakedRequestor},
			want: stakedRequestor,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			context, _, keeper := createTestInput(t, true)
			keeper.SetRequestor(context, tt.args.requestor)
			keeper.SetStakedRequestor(context, tt.args.requestor)
			if got, _ := keeper.RequestorCache.Get(tt.args.requestor.Address.String()); !got.(types.Requestor).Equals(tt.want) {
				t.Errorf("keeperRequestorUtil.RequestorCaching()= %v, want %v", got, tt.want)
			}
		})
	}
}
