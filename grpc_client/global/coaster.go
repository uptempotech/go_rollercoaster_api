package global

// NilCoaster is the nil value for an account
var NilCoaster Coaster

// Coaster defines what is stored in mongodb.
type Coaster struct {
	Name         string `json:"name"`
	Manufacturer string `json:"manufacturer"`
	CoasterID    string `json:"coasterId"`
	InPark       string `json:"inPark"`
	Height       int32  `json:"height"`
}
