package global

import "go.mongodb.org/mongo-driver/bson/primitive"

// NilAccount is the nil value for an account
var NilCoaster Coaster

// Coaster defines what is stored in mongodb.
type Coaster struct {
	ID           primitive.ObjectID `bson:"_id"`
	Name         string             `bson:"name"`
	Manufacturer string             `bson:"manufacturer"`
	CoasterID    string             `bson:"coaster_id"`
	InPark       string             `bson:"inPark"`
	Height       int                `bson:"height"`
}