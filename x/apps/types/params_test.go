package types

import (
	"fmt"
	"reflect"
	"testing"
	"time"
)

func TestDefaultParams(t *testing.T) {
	tests := []struct {
		name string
		want Params
	}{
		{"Default Test",
			Params{
				UnstakingTime:       DefaultUnstakingTime,
				MaxApplications:     DefaultMaxApplications,
				MinAppStake:         DefaultMinStake,
				BaseRelaysPerVIPR:   DefaultBaseRelaysPerVIP,
				StabilityModulation: DefaultStabilityModulation,
				ParticipationRate:   DefaultParticipationRateOn,
				MaxChains:           DefaultMaxChains,
			},
		}}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := DefaultParams(); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("DefaultParams() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestParams_Equal(t *testing.T) {
	type fields struct {
		UnstakingTime              time.Duration `json:"unstaking_time" yaml:"unstaking_time"`       // duration of unstaking
		MaxApplications            int64         `json:"max_applications" yaml:"max_applications"`   // maximum number of applications
		MinAppStake                int64         `json:"minimum_app_stake" yaml:"minimum_app_stake"` // minimum amount needed to stake
		BaslineThroughputStakeRate int64
		StabilityModulation        int64
		ParticipationRate          bool
	}
	type args struct {
		p2 Params
	}
	tests := []struct {
		name   string
		fields fields
		args   args
		want   bool
	}{
		{"Default Test Equal", fields{
			UnstakingTime:              0,
			MaxApplications:            0,
			MinAppStake:                0,
			BaslineThroughputStakeRate: 0,
			StabilityModulation:        0,
			ParticipationRate:          false,
		}, args{Params{
			UnstakingTime:       0,
			MaxApplications:     0,
			MinAppStake:         0,
			BaseRelaysPerVIPR:   0,
			StabilityModulation: 0,
			ParticipationRate:   false,
		}}, true},
		{"Default Test False", fields{
			UnstakingTime:              0,
			MaxApplications:            0,
			MinAppStake:                0,
			BaslineThroughputStakeRate: 0,
			StabilityModulation:        0,
			ParticipationRate:          false,
		}, args{Params{
			UnstakingTime:       0,
			MaxApplications:     1,
			MinAppStake:         0,
			BaseRelaysPerVIPR:   0,
			StabilityModulation: 0,
			ParticipationRate:   false,
		}}, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := Params{
				UnstakingTime:       tt.fields.UnstakingTime,
				MaxApplications:     tt.fields.MaxApplications,
				MinAppStake:         tt.fields.MinAppStake,
				BaseRelaysPerVIPR:   tt.fields.BaslineThroughputStakeRate,
				StabilityModulation: tt.fields.StabilityModulation,
				ParticipationRate:   tt.fields.ParticipationRate,
			}
			if got := p.Equal(tt.args.p2); got != tt.want {
				t.Errorf("Equal() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestParams_Validate(t *testing.T) {
	type fields struct {
		UnstakingTime               time.Duration `json:"unstaking_time" yaml:"unstaking_time"`       // duration of unstaking
		MaxApplications             int64         `json:"max_applications" yaml:"max_applications"`   // maximum number of applications
		MinAppStake                 int64         `json:"minimum_app_stake" yaml:"minimum_app_stake"` // minimum amount needed to stake
		BaselineThrouhgputStakeRate int64         `json:"baseline_throughput_stake_rate" yaml:"baseline_throughput_stake_rate"`
		StabilityModulation         int64         `json:"staking_adjustment" yaml:"staking_adjustment"`
		ParticipationRate           bool          `json:"participation_rate_on" yaml:"participation_rate_on"`
	}
	tests := []struct {
		name    string
		fields  fields
		wantErr bool
	}{
		{"Default Validation Test / Wrong All Parameters", fields{
			UnstakingTime:               0,
			MaxApplications:             0,
			MinAppStake:                 0,
			BaselineThrouhgputStakeRate: 1,
			StabilityModulation:         0,
			ParticipationRate:           false,
		}, true},
		{"Default Validation Test / Wrong Appstake", fields{
			UnstakingTime:               0,
			MaxApplications:             2,
			MinAppStake:                 0,
			BaselineThrouhgputStakeRate: 0,
		}, true},
		{"Default Validation Test / Wrong BaselinethroughputStakeRate", fields{
			UnstakingTime:               10000,
			MaxApplications:             2,
			MinAppStake:                 1000000,
			BaselineThrouhgputStakeRate: -1,
		}, true},
		{"Default Validation Test / Valid", fields{
			UnstakingTime:               10000,
			MaxApplications:             2,
			MinAppStake:                 1000000,
			BaselineThrouhgputStakeRate: 90,
			StabilityModulation:         100,
			ParticipationRate:           false,
		}, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := Params{
				UnstakingTime:       tt.fields.UnstakingTime,
				MaxApplications:     tt.fields.MaxApplications,
				MinAppStake:         tt.fields.MinAppStake,
				BaseRelaysPerVIPR:   tt.fields.BaselineThrouhgputStakeRate,
				StabilityModulation: tt.fields.StabilityModulation,
				ParticipationRate:   tt.fields.ParticipationRate,
			}
			if err := p.Validate(); (err != nil) != tt.wantErr {
				t.Errorf("Validate() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestParams_String(t *testing.T) {
	tests := []struct {
		name string
		want string
	}{
		{
			"Default Test",
			fmt.Sprintf(`Params:
  Unstaking Time:              %s
  Max Applications:            %d
  Minimum Stake:     	       %d
  BaseRelaysPerVIPR            %d
  Stability Adjustment         %d
  Participation Rate On        %v
  Maximum Chains               %d,`,
				DefaultUnstakingTime,
				DefaultMaxApplications,
				DefaultMinStake,
				DefaultBaseRelaysPerVIP,
				DefaultStabilityModulation,
				DefaultParticipationRateOn,
				DefaultMaxChains),
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {

		})
	}
}
