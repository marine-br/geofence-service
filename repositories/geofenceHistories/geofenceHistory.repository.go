package geofenceHistoryRepositories

import "github.com/marine-br/golib-utils/models"

type InsertGeofenceHistoryParams struct {
	TrackerMessage models.TrackerMsgType
	Status         string
	Geofence       models.GeofenceType
}

type GeofenceRepository interface {
	InsertGeofenceHistory(param InsertGeofenceHistoryParams) error
}
