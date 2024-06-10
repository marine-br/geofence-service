package geofenceHistoryRepositories

import (
	"context"
	"fmt"
	db "github.com/marine-br/golib-mongo"
	"go.mongodb.org/mongo-driver/mongo"
	"os"
	"time"
)

type MongoGeofenceHistoryRepository struct {
	GeofenceHistoryCollection *mongo.Collection
}

func NewMongoGeofenceHistoryRepository(client db.MongoClient) MongoGeofenceHistoryRepository {
	geofenceRepo := MongoGeofenceHistoryRepository{
		GeofenceHistoryCollection: client.GetCollection(os.Getenv("MONGO_GEOFENCE_HISTORIES_COLLECTION")),
	}

	return geofenceRepo
}

func (m *MongoGeofenceHistoryRepository) InsertGeofenceHistory(param InsertGeofenceHistoryParams) error {
	geofenceHistory := GeofenceHistoryModel{
		Type:           param.Status,
		TrackerMessage: param.TrackerMessage.ID,
		Tracker:        param.TrackerMessage.TRACKER,
		Vehicle:        param.TrackerMessage.VEHICLE,
		Company:        param.TrackerMessage.COMPANY,
		Driver:         param.TrackerMessage.DRIVER,
		Geofence:       param.Geofence.ID,
		Date:           param.TrackerMessage.GPS_TIME,
		CreatedAt:      time.Now(),
	}

	_, err := m.GeofenceHistoryCollection.InsertOne(context.Background(), geofenceHistory)
	if err != nil {
		return fmt.Errorf("failed to insert geofence history: %v", err)
	}

	return nil
}
