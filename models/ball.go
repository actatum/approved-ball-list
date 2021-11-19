package models

import "encoding/xml"

var currentBrands = map[string]bool{
	"900 Global":  true,
	"BIG Bowling": true,
	"Brunswick":   true,
	"Columbia":    true,
	"Ebonite":     true,
	"Hammer":      true,
	"Motiv":       true,
	"Radical":     true,
	"Roto Grip":   true,
	"Storm":       true,
}

type BallList struct {
	XMLName xml.Name `xml:"BallList"`
	Balls   []Ball   `xml:"Brand"`
}

type Ball struct {
	XMLName      xml.Name `xml:"Brand" json:"-" firestore:"-"`
	Brand        string   `xml:"name,attr" json:"brand" firestore:"brand"`
	Name         string   `xml:"BallName" json:"name" firestore:"name"`
	DateApproved string   `xml:"DateApproved" json:"date_approved" firestore:"date_approved"`
	ImageURL     string   `xml:"link" json:"image_url" firestore:"image_url"`
}

func (b *BallList) Filter() {
	n := 0
	for _, ball := range b.Balls {
		if _, ok := currentBrands[ball.Brand]; ok {
			b.Balls[n] = ball
			n++
		}
	}

	b.Balls = b.Balls[:n]
}
