package balls

import (
	"context"
	"time"
)

// Brand is a brand that makes bowling equipment.
type Brand string

// All active brands.
const (
	Global      Brand = "900 Global"
	BigBowling  Brand = "BIG Bowling"
	Brunswick   Brand = "Brunswick"
	Columbia300 Brand = "Columbia"
	DV8         Brand = "DV8"
	Ebonite     Brand = "Ebonite"
	Motiv       Brand = "Motiv"
	Radical     Brand = "Radical"
	RotoGrip    Brand = "Roto Grip"
	Storm       Brand = "Storm"
	Swag        Brand = "Swag"
	Track       Brand = "Track Inc."
)

var allBrands = []Brand{
	Global,
	BigBowling,
	Brunswick,
	Columbia300,
	DV8,
	Ebonite,
	Motiv,
	Radical,
	RotoGrip,
	Storm,
	Swag,
	Track,
}

type Ball struct {
	ID           string
	Brand        Brand
	Name         string
	ApprovalDate time.Time
	ImageURL     string
}

type Repository interface {
	Add(ctx context.Context, balls ...Ball) error
	FindAll(ctx context.Context, filter Filter) ([]Ball, error)
}

type Filter struct {
	Brand Brand
}
