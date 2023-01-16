package keeper

import (
	"reflect"
	"testing"

	"github.com/vipernet-xyz/viper-network/x/platforms/exported"
	"github.com/vipernet-xyz/viper-network/x/platforms/types"
)

func TestPlatformUtil_Platform(t *testing.T) {
	stakedPlatform := getStakedPlatform()

	type args struct {
		platform types.Platform
	}
	type want struct {
		platform types.Platform
	}
	tests := []struct {
		name string
		find bool
		args
		want
	}{
		{
			name: "gets platform",
			find: false,
			args: args{platform: stakedPlatform},
			want: want{platform: stakedPlatform},
		},
		{
			name: "errors if no platform",
			find: true,
			args: args{platform: stakedPlatform},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			context, _, keeper := createTestInput(t, true)
			switch tt.find {
			case true:
				if got := keeper.Platform(context, tt.args.platform.Address); got != nil {
					t.Errorf("keeperPlatformUtil.Platform()= %v, want nil", got)
				}
			default:
				keeper.SetPlatform(context, tt.args.platform)
				keeper.SetStakedPlatform(context, tt.args.platform)
				if got := keeper.Platform(context, tt.args.platform.Address); !reflect.DeepEqual(got, tt.want.platform) {
					t.Errorf("keeperPlatformUtil.Platform()= %v, want %v", got, tt.want.platform)
				}
			}
		})
	}

}

func TestPlatformUtil_AllPlatforms(t *testing.T) {
	stakedPlatform := getStakedPlatform()

	type args struct {
		platform types.Platform
	}
	tests := []struct {
		name   string
		panics bool
		args
		expected []exported.PlatformI
	}{
		{
			name:     "gets platform",
			panics:   false,
			args:     args{platform: stakedPlatform},
			expected: []exported.PlatformI{stakedPlatform},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			context, _, keeper := createTestInput(t, true)
			keeper.SetPlatform(context, tt.args.platform)
			keeper.SetStakedPlatform(context, tt.args.platform)
			if got := keeper.AllPlatforms(context); !reflect.DeepEqual(got, tt.expected) {
				t.Errorf("keeperPlatformUtil.AllPlatforms()= %v, want %v", got, tt.expected)
			}
		})
	}
}

//func TestNewPlatformCaching(t *testing.T) { todo
//	stakedPlatform := getStakedPlatform()
//
//	type args struct {
//		bz        []byte
//		platform types.Platform
//	}
//	type expected struct {
//		platform types.Platform
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
//			args:     args{platform: stakedPlatform},
//			expected: expected{platform: stakedPlatform, length: 1},
//		},
//	}
//
//	for _, test := range tests {
//		t.Run(test.name, func(t *testing.T) {
//			context, _, keeper := createTestInput(t, true)
//			keeper.SetPlatform(context, test.args.platform)
//			keeper.SetStakedPlatform(context, test.args.platform)
//			store := context.KVStore(keeper.storeKey)
//			key := types.KeyForPlatformPrevStateStateByPower(test.args.platform.Address)
//			store.Set(key, test.args.platform.Address)
//			powermap := keeper.getPrevStatePowerMap(context)
//			assert.Len(t, powermap, test.expected.length, "does not have correct length")
//			var valAddr [sdk.AddrLen]byte
//			copy(valAddr[:], key[1:])
//
//			for mapKey, value := range powermap {
//				assert.Equal(t, valAddr, mapKey, "key is not correct")
//				bz := make([]byte, len(test.args.platform.Address))
//				copy(bz, test.args.platform.Address)
//				assert.Equal(t, bz, value, "key is not correct")
//			}
//		})
//	}
//}

func TestPlatformUtil_PlatformCaching(t *testing.T) {
	stakedPlatform := getStakedPlatform()

	type args struct {
		platform types.Platform
	}
	tests := []struct {
		name   string
		panics bool
		args
		want types.Platform
	}{
		{
			name: "gets platform",
			args: args{platform: stakedPlatform},
			want: stakedPlatform,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			context, _, keeper := createTestInput(t, true)
			keeper.SetPlatform(context, tt.args.platform)
			keeper.SetStakedPlatform(context, tt.args.platform)
			if got, _ := keeper.PlatformCache.Get(tt.args.platform.Address.String()); !got.(types.Platform).Equals(tt.want) {
				t.Errorf("keeperPlatformUtil.PlatformCaching()= %v, want %v", got, tt.want)
			}
		})
	}
}
