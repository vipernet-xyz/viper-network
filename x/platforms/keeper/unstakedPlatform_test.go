package keeper

import (
	"testing"

	sdk "github.com/vipernet-xyz/viper-network/types"
	"github.com/vipernet-xyz/viper-network/x/platforms/types"

	"github.com/stretchr/testify/assert"
)

func TestPlatformUnstaked_GetAndSetlUnstaking(t *testing.T) {
	stakedPlatform := getStakedPlatform()
	unstaking := getUnstakingPlatform()

	type want struct {
		platforms       []types.Platform
		stakedPlatforms bool
		length          int
	}
	type args struct {
		platforms      []types.Platform
		stakedPlatform types.Platform
	}
	tests := []struct {
		name      string
		platform  types.Platform
		platforms []types.Platform
		want
		args
	}{
		{
			name: "gets platforms",
			args: args{platforms: []types.Platform{unstaking}},
			want: want{platforms: []types.Platform{unstaking}, length: 1, stakedPlatforms: false},
		},
		{
			name: "gets emtpy slice of platforms",
			want: want{length: 0, stakedPlatforms: true},
			args: args{stakedPlatform: stakedPlatform},
		},
		{
			name:      "only gets unstaking platforms",
			platforms: []types.Platform{stakedPlatform, unstaking},
			want:      want{length: 1, stakedPlatforms: true},
			args:      args{stakedPlatform: stakedPlatform, platforms: []types.Platform{unstaking}},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			context, _, keeper := createTestInput(t, true)
			for _, platform := range tt.args.platforms {
				keeper.SetPlatform(context, platform)
			}
			if tt.want.stakedPlatforms {
				keeper.SetPlatform(context, tt.args.stakedPlatform)
			}
			platforms := keeper.getAllUnstakingPlatforms(context)
			if len(platforms) != tt.want.length {
				t.Errorf("platformUnstaked.GetPlatforms() = %v, want %v", len(platforms), tt.want.length)
			}
		})
	}
}

func TestPlatformUnstaked_DeleteUnstakingPlatform(t *testing.T) {
	stakedPlatform := getStakedPlatform()
	secondStakedPlatform := getStakedPlatform()

	type want struct {
		stakedPlatforms bool
		length          int
	}
	type args struct {
		platforms []types.Platform
	}
	tests := []struct {
		name      string
		platform  types.Platform
		platforms []types.Platform
		sets      bool
		want
		args
	}{
		{
			name: "deletes",
			args: args{platforms: []types.Platform{stakedPlatform, secondStakedPlatform}},
			sets: false,
			want: want{length: 1, stakedPlatforms: false},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			context, _, keeper := createTestInput(t, true)
			for _, platform := range tt.args.platforms {
				keeper.SetPlatform(context, platform)
			}
			keeper.SetUnstakingPlatform(context, tt.args.platforms[0])
			_ = keeper.getAllUnstakingPlatforms(context)

			keeper.deleteUnstakingPlatform(context, tt.args.platforms[1])

			if got := keeper.getAllUnstakingPlatforms(context); len(got) != tt.want.length {
				t.Errorf("KeeperCoins.BurnStakedTokens()= %v, want %v", len(got), tt.want.length)
			}
		})
	}
}

func TestPlatformUnstaked_DeleteUnstakingPlatforms(t *testing.T) {
	stakedPlatform := getStakedPlatform()
	secondaryStakedPlatform := getStakedPlatform()

	type want struct {
		stakedPlatforms bool
		length          int
	}
	type args struct {
		platforms []types.Platform
	}
	tests := []struct {
		name      string
		platform  types.Platform
		platforms []types.Platform
		want
		args
	}{
		{
			name: "deletes all unstaking platform",
			args: args{platforms: []types.Platform{stakedPlatform, secondaryStakedPlatform}},
			want: want{length: 0, stakedPlatforms: false},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			context, _, keeper := createTestInput(t, true)
			for _, platform := range test.args.platforms {
				keeper.SetPlatform(context, platform)
				keeper.SetUnstakingPlatform(context, platform)
				keeper.deleteUnstakingPlatforms(context, platform.UnstakingCompletionTime)
			}

			platforms := keeper.getAllUnstakingPlatforms(context)

			assert.Equalf(t, test.want.length, len(platforms), "length of the platforms does not match want on %v", test.name)
		})
	}
}

func TestPlatformUnstaked_GetAllMaturePlatforms(t *testing.T) {
	stakingPlatform := getUnstakingPlatform()

	type want struct {
		platforms       []types.Platform
		stakedPlatforms bool
		length          int
	}
	type args struct {
		platforms []types.Platform
	}
	tests := []struct {
		name      string
		platform  types.Platform
		platforms []types.Platform
		want
		args
	}{
		{
			name: "gets all mature platforms",
			args: args{platforms: []types.Platform{stakingPlatform}},
			want: want{platforms: []types.Platform{stakingPlatform}, length: 1, stakedPlatforms: false},
		},
		{
			name: "gets empty slice if no mature platforms",
			args: args{platforms: []types.Platform{}},
			want: want{platforms: []types.Platform{stakingPlatform}, length: 0, stakedPlatforms: false},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			context, _, keeper := createTestInput(t, true)
			for _, platform := range tt.args.platforms {
				keeper.SetPlatform(context, platform)
			}
			if got := keeper.getMaturePlatforms(context); len(got) != tt.want.length {
				t.Errorf("platformUnstaked.unstakeAllMaturePlatforms()= %v, want %v", len(got), tt.want.length)
			}
		})
	}
}

//func TestPlatformUnstaked_UnstakeAllMaturePlatforms(t *testing.T) {
//	stakingPlatform := getUnstakingPlatform()
//
//	type want struct {
//		platforms       []types.Platform
//		stakedPlatforms bool
//		length             int
//	}
//	type args struct {
//		stakedVal         types.Platform
//		platforms      []types.Platform
//		stakedPlatform types.Platform
//	}
//	tests := []struct {
//		name         string
//		platform  types.Platform
//		platforms []types.Platform
//		want
//		args
//	}{
//		{
//			name: "unstake mature platforms",
//			args: args{platforms: []types.Platform{stakingPlatform}},
//			want: want{platforms: []types.Platform{stakingPlatform}, length: 0, stakedPlatforms: false},
//		},
//	}
//
//	for _, tt := range tests {
//		t.Run(tt.name, func(t *testing.T) {
//			context, _, keeper := createTestInput(t, true)
//			for _, platform := range tt.args.platforms {
//				keeper.SetPlatform(context, platform)
//				keeper.SetUnstakingPlatform(context, platform)
//			}
//			keeper.unstakeAllMaturePlatforms(context)
//			if got := keeper.getAllUnstakingPlatforms(context); len(got) != tt.want.length {
//				t.Errorf("platformUnstaked.unstakeAllMaturePlatforms()= %v, want %v", len(got), tt.want.length)
//			}
//		})
//	}
//}

func TestPlatformUnstaked_UnstakingPlatformsIterator(t *testing.T) {
	stakedPlatform := getStakedPlatform()
	unstakedPlatform := getUnstakedPlatform()

	tests := []struct {
		name      string
		platforms []types.Platform
		panics    bool
		amount    sdk.BigInt
	}{
		{
			name:      "recieves a valid iterator",
			platforms: []types.Platform{stakedPlatform, unstakedPlatform},
			panics:    false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			context, _, keeper := createTestInput(t, true)
			for _, platform := range tt.platforms {
				keeper.SetPlatform(context, platform)
				keeper.SetStakedPlatform(context, platform)
			}

			it, _ := keeper.unstakingPlatformsIterator(context, context.BlockHeader().Time)
			if v, ok := it.(sdk.Iterator); !ok {
				t.Errorf("platformUnstaked.UnstakingPlatformsIterator()= %v does not implement sdk.Iterator", v)
			}
		})
	}
}
