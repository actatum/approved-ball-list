package models

import (
	"encoding/xml"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestBallList_Filter(t *testing.T) {
	type fields struct {
		XMLName xml.Name
		Balls   []Ball
	}
	tests := []struct {
		name   string
		fields fields
	}{
		// TODO: Add test cases.
		{
			name: "filter balls",
			fields: fields{
				XMLName: xml.Name{},
				Balls: []Ball{
					{
						XMLName:      xml.Name{},
						Brand:        "Storm",
						Name:         "xD",
						DateApproved: "today",
						ImageURL:     "",
					},
					{
						XMLName:      xml.Name{},
						Brand:        "SwagMoney",
						Name:         "lul",
						DateApproved: "yesterday",
						ImageURL:     "",
					},
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			b := &BallList{
				XMLName: tt.fields.XMLName,
				Balls:   tt.fields.Balls,
			}

			b.Filter()

			assert.Equal(t, 1, len(b.Balls))
			assert.Equal(t, Ball{
				XMLName:      xml.Name{},
				Brand:        "Storm",
				Name:         "xD",
				DateApproved: "today",
				ImageURL:     "",
			}, b.Balls[0])
		})
	}
}
