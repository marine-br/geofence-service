package geofencesRepositories

import (
	"github.com/marine-br/golib-utils/models"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type GetGeofenceParams struct {
	CompanyId primitive.ObjectID `bson:"_id,omitempty"`
}

type GeofenceRepository interface {
	GetGeofences(param GetGeofenceParams) ([]models.GeofenceType, error)
}
