package geofenceHistoryRepositories

import (
	"go.mongodb.org/mongo-driver/bson/primitive"
	"time"
)

type GeofenceHistoryModel struct {
	ID             primitive.ObjectID `bson:"_id,omitempty"`
	Type           string             `bson:"type"`
	TrackerMessage primitive.ObjectID `bson:"trackerMessage"`
	Vehicle        primitive.ObjectID `bson:"vehicle"`
	Tracker        primitive.ObjectID `bson:"tracker"`
	Company        primitive.ObjectID `bson:"company"`
	Geofence       primitive.ObjectID `bson:"geofence"`
	Date           time.Time          `bson:"date"`
	CreatedAt      time.Time          `bson:"createdAt"`
}
