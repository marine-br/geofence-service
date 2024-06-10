package PolygonFromPrimitiveD

import (
	"github.com/marine-br/geoafence-service/utils/IsPointInsidePolygon"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"log"
	"time"
)

type GeofenceType struct {
	ID          primitive.ObjectID `bson:"_id,omitempty"`
	Name        string             `bson:"name"`
	Description string             `bson:"description"`
	Tags        []any              `bson:"tags"`
	Geojson     any                `bson:"geojson"`
	Companies   []string           `bson:"companies"`
	CreatedBy   primitive.ObjectID `bson:"createdBy"`
	UpdatedBy   primitive.ObjectID `bson:"updatedBy"`
	CreatedAt   time.Time          `bson:"createdAt"`
	UpdatedAt   time.Time          `bson:"updatedAt"`
	Deleted     bool               `bson:"deleted"`
	DeletedAt   time.Time          `bson:"deletedAt"`
	DeletedBy   primitive.ObjectID `bson:"deletedBy"`
}

func PolygonFromPrimitiveD(geojson primitive.D) []IsPointInsidePolygon.Point {
	features, ok := getValueByKey(geojson, "features").(primitive.A)
	if !ok {
		log.Fatalf("features not found or incorrect type")
	}

	for i, feature := range features {
		featureMap, ok := feature.(primitive.D)
		if !ok {
			continue
		}

		geometry, ok := getValueByKey(featureMap, "geometry").(primitive.D)
		if !ok {
			continue
		}

		if geometryType, ok := getValueByKey(geometry, "type").(string); ok && geometryType == "Polygon" {
			coordinates, ok := getValueByKey(geometry, "coordinates").(primitive.A)
			if !ok || len(coordinates) == 0 {
				continue
			}

			// Extract the first polygon (assuming it's the correct structure)
			polygonCoordinates := coordinates[0].(primitive.A)

			// Convert to []Point
			var polygon []IsPointInsidePolygon.Point
			for _, coord := range polygonCoordinates {
				point := coord.(primitive.A)
				polygon = append(polygon, IsPointInsidePolygon.Point{X: point[1].(float64), Y: point[0].(float64)})
			}

			// Remove the polygon from the features
			features = append(features[:i], features[i+1:]...)
			setValueByKey(&geojson, "features", features)

			return polygon
		}
	}

	return nil
}

func getValueByKey(d primitive.D, key string) interface{} {
	for _, e := range d {
		if e.Key == key {
			return e.Value
		}
	}
	return nil
}

func setValueByKey(d *primitive.D, key string, value interface{}) {
	for i, e := range *d {
		if e.Key == key {
			(*d)[i].Value = value
			return
		}
	}
	*d = append(*d, primitive.E{Key: key, Value: value})
}
