package redisRepositories

import (
	geofenceHistoryRepositories "github.com/marine-br/geoafence-service/repositories/geofenceHistories"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type GetLastGeofenceParams struct {
	VehicleID  primitive.ObjectID `json:"vehicle_id"`
	GeofenceId primitive.ObjectID `json:"geofence_id"`
}

type SetLastGeofenceParams struct {
	VehicleID  primitive.ObjectID `json:"vehicle_id"`
	GeofenceId primitive.ObjectID `json:"geofence_id"`
	Value      geofenceHistoryRepositories.GeofenceHistoryModel
}

type CacheRepository interface {
	GetLastGeofence(param GetLastGeofenceParams) (*geofenceHistoryRepositories.GeofenceHistoryModel, error)
	SetLastGeofence(param SetLastGeofenceParams) bool
}
