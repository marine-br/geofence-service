package geofencesRepositories

import (
	"context"
	db "github.com/marine-br/golib-mongo"
	"github.com/marine-br/golib-utils/models"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"os"
)

type MongoGeofenceRepository struct {
	GeofenceCollection *mongo.Collection
}

func NewMongoGeofenceRepository(client db.MongoClient) MongoGeofenceRepository {
	geofenceRepo := MongoGeofenceRepository{
		GeofenceCollection: client.GetCollection(os.Getenv("MONGO_GEOFENCE_COLLECTION")),
	}

	return geofenceRepo
}

func (m *MongoGeofenceRepository) GetGeofences(param GetGeofenceParams) ([]models.GeofenceType, error) {
	filter := bson.D{
		{"companies",
			bson.D{
				{"$in",
					bson.A{
						param.CompanyId,
					},
				},
			},
		},
		{"deleted", bson.D{{"$ne", true}}},
	}

	// Executa a busca no MongoDB
	cursor, err := m.GeofenceCollection.Find(context.Background(), filter)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(context.Background())

	// Declara uma slice para armazenar os resultados
	var geofences []models.GeofenceType

	// Itera sobre os resultados e os decodifica para a slice de geofences
	for cursor.Next(context.Background()) {
		var geofence models.GeofenceType
		if err := cursor.Decode(&geofence); err != nil {
			return nil, err
		}
		geofences = append(geofences, geofence)
	}

	// Verifica se houve algum erro durante a iteração
	if err := cursor.Err(); err != nil {
		return nil, err
	}

	return geofences, nil
}
