package geofenceHistoryRepositories

import (
	"github.com/marine-br/golib-utils/models"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type InsertGeofenceHistoryParams struct {
	TrackerMessage models.TrackerMsgType
	Status         string
	Geofence       models.GeofenceType
}

type FindFirstAfterGeofenceHistoryParams struct {
	TrackerMessage models.TrackerMsgType
	Geofence       models.GeofenceType
}

type DeleteGeofenceHistoryParams struct {
	GeofenceHistoryId primitive.ObjectID
}

type FindLastGeofenceHistoryParams struct {
	TrackerMessage models.TrackerMsgType
	Geofence       models.GeofenceType
}

type GeofenceHistoryRepository interface {
	InsertGeofenceHistory(param InsertGeofenceHistoryParams) (GeofenceHistoryModel, error)
	FindFirstAfterGeofenceHistory(param FindFirstAfterGeofenceHistoryParams) (*GeofenceHistoryModel, error)
	FindLastGeofenceHistory(param FindLastGeofenceHistoryParams) (GeofenceHistoryModel, error)
	DeleteGeofenceHistory(param DeleteGeofenceHistoryParams) error
}
