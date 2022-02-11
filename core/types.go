package core

import (
	"encoding/xml"

	"github.com/rs/zerolog"
)

// AllBrands is a list of all active brands
var AllBrands = []string{"900 Global", "BIG Bowling", "Brunswick", "Columbia", "DV8", "Ebonite", "Hammer", "Motiv", "Radical", "Roto Grip", "Storm", "Track Inc."}

// CurrentBrands is a map of active brands
var CurrentBrands = map[string]bool{
	"900 Global":  true,
	"BIG Bowling": true,
	"Brunswick":   true,
	"Columbia":    true,
	"DV8":         true,
	"Ebonite":     true,
	"Hammer":      true,
	"Motiv":       true,
	"Radical":     true,
	"Roto Grip":   true,
	"Storm":       true,
	"Track Inc.":  true,
}

// Config is the configurable values for the service
type Config struct {
	Logger          *zerolog.Logger
	DiscordChannels map[string]DiscordChannel // Channel name is key, Channel ID is value
	Repository      Repository
	Alerter         Alerter
	USBC            USBC
}

// DiscordChannel configures the discord channel that alerts can be sent to
type DiscordChannel struct {
	Name   string
	ID     string
	Brands []string
}

func (d DiscordChannel) containsBrand(brand string) bool {
	for _, b := range d.Brands {
		if b == brand {
			return true
		}
	}

	return false
}

// Ball is the domain object for a ball
type Ball struct {
	XMLName      xml.Name `xml:"Brand" json:"-" firestore:"-"`
	Brand        string   `xml:"name,attr" json:"brand" firestore:"brand"`
	Name         string   `xml:"BallName" json:"name" firestore:"name"`
	DateApproved string   `xml:"DateApproved" json:"date_approved" firestore:"date_approved"`
	ImageURL     string   `xml:"link" json:"image_url" firestore:"image_url"`
}

// BallList is used to unmarshal result from usbc
type BallList struct {
	XMLName xml.Name `xml:"BallList"`
	Balls   []Ball   `xml:"Brand"`
}

// RetrieveBallResult is a common return type for results from USBC and Repository
type RetrieveBallResult struct {
	Balls []Ball
	Err   error
}
