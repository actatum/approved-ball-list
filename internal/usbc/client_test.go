package usbc

import (
	"reflect"
	"testing"
)

func Test_filterBrands(t *testing.T) {
	type args struct {
		balls []ball
	}
	tests := []struct {
		name string
		args args
		want []ball
	}{
		{
			name: "filter non active brands",
			args: args{
				balls: []ball{
					{
						Brand:        "Lane #1",
						Name:         "Hex",
						DateApproved: "12-20",
						ImageURL:     "xD",
					},
					{
						Brand:        "Track Inc.",
						Name:         "Kinetic",
						DateApproved: "20-12",
						ImageURL:     "def",
					},
				},
			},
			want: []ball{
				{
					Brand:        "Track Inc.",
					Name:         "Kinetic",
					DateApproved: "20-12",
					ImageURL:     "def",
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := filterBrands(tt.args.balls); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("filterBrands() = %v, want %v", got, tt.want)
			}
		})
	}
}
