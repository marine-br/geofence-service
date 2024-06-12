package geofenceHistoryRepositories

import (
	"context"
	"errors"
	"fmt"
	db "github.com/marine-br/golib-mongo"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
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

func (m *MongoGeofenceHistoryRepository) FindFirstAfterGeofenceHistory(param FindFirstAfterGeofenceHistoryParams) (*GeofenceHistoryModel, error) {
	var geofenceHistory GeofenceHistoryModel

	filter := bson.M{
		"tracker":  param.TrackerMessage.TRACKER,
		"geofence": param.Geofence.ID,
		"date": bson.M{
			"$gte": param.TrackerMessage.GPS_TIME,
		},
	}
	opts := options.FindOne().SetSort(bson.D{{"date", 1}})

	err := m.GeofenceHistoryCollection.FindOne(context.Background(), filter, opts).Decode(&geofenceHistory)
	if err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			return nil, nil // No documents found, return empty model
		}
		return nil, fmt.Errorf("failed to find last geofence history: %v", err)
	}

	return &geofenceHistory, nil
}

func (m *MongoGeofenceHistoryRepository) FindLastGeofenceHistory(param FindLastGeofenceHistoryParams) (GeofenceHistoryModel, error) {
	var geofenceHistory GeofenceHistoryModel

	filter := bson.M{
		"tracker":  param.TrackerMessage.TRACKER,
		"geofence": param.Geofence.ID,
		"date": bson.M{
			"$lt": param.TrackerMessage.GPS_TIME,
		},
	}
	opts := options.FindOne().SetSort(bson.D{{"date", -1}})

	err := m.GeofenceHistoryCollection.FindOne(context.Background(), filter, opts).Decode(&geofenceHistory)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return GeofenceHistoryModel{}, nil // No documents found, return empty model
		}
		return GeofenceHistoryModel{}, fmt.Errorf("failed to find last geofence history: %v", err)
	}

	return geofenceHistory, nil
}

func (m *MongoGeofenceHistoryRepository) DeleteGeofenceHistory(param DeleteGeofenceHistoryParams) error {
	filter := bson.M{"_id": param.GeofenceHistoryId}

	_, err := m.GeofenceHistoryCollection.DeleteOne(context.Background(), filter)
	if err != nil {
		return fmt.Errorf("failed to delete geofence history: %v", err)
	}

	return nil
}
