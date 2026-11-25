package PolygonFromPrimitiveD

import (
	"errors"
	"fmt"
	"math"
	"time"

	"github.com/marine-br/geoafence-service/utils/IsPointInsidePolygon"
	"go.mongodb.org/mongo-driver/bson/primitive"
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

// ErrNoPolygon means the geojson has no usable Polygon or MultiPolygon area.
var ErrNoPolygon = errors.New("geojson has no usable Polygon or MultiPolygon")

// PolygonsFromPrimitiveD reads a geofence geojson stored as a FeatureCollection
// and returns the outer ring of every Polygon and of every polygon inside a
// MultiPolygon, in every feature. Holes (inner rings) are ignored, as before.
// Features with other geometry types are ignored. Rings with fewer than three
// points are skipped, since they contain no point.
//
// It never exits the process: a geojson without "features", or with
// coordinates of an unexpected shape or type, returns an error so the caller
// can skip that fence. A geojson with no usable area returns ErrNoPolygon.
func PolygonsFromPrimitiveD(geojson primitive.D) ([][]IsPointInsidePolygon.Point, error) {
	features, ok := getValueByKey(geojson, "features").(primitive.A)
	if !ok {
		return nil, errors.New("geojson has no features array")
	}

	var polygons [][]IsPointInsidePolygon.Point
	for i, feature := range features {
		featureMap, ok := feature.(primitive.D)
		if !ok {
			continue
		}
		geometry, ok := getValueByKey(featureMap, "geometry").(primitive.D)
		if !ok {
			continue
		}
		geometryType, _ := getValueByKey(geometry, "type").(string)
		if geometryType != "Polygon" && geometryType != "MultiPolygon" {
			continue
		}
		coordinates, ok := getValueByKey(geometry, "coordinates").(primitive.A)
		if !ok {
			return nil, fmt.Errorf("feature %d: %s coordinates are not an array", i, geometryType)
		}

		polygonList := primitive.A{coordinates}
		if geometryType == "MultiPolygon" {
			polygonList = coordinates
		}
		for j, rawPolygon := range polygonList {
			rings, ok := rawPolygon.(primitive.A)
			if !ok {
				return nil, fmt.Errorf("feature %d polygon %d: rings are not an array", i, j)
			}
			if len(rings) == 0 {
				continue
			}
			ring, err := ringToPoints(rings[0])
			if err != nil {
				return nil, fmt.Errorf("feature %d polygon %d: %w", i, j, err)
			}
			if len(ring) < 3 {
				continue
			}
			polygons = append(polygons, ring)
		}
	}

	if len(polygons) == 0 {
		return nil, ErrNoPolygon
	}
	return polygons, nil
}

// ringToPoints converts [[lon, lat], ...] to points with X = latitude and
// Y = longitude, the convention main.go uses for the tracker position.
func ringToPoints(rawRing any) ([]IsPointInsidePolygon.Point, error) {
	ring, ok := rawRing.(primitive.A)
	if !ok {
		return nil, errors.New("ring is not an array")
	}
	points := make([]IsPointInsidePolygon.Point, 0, len(ring))
	for k, rawCoord := range ring {
		coord, ok := rawCoord.(primitive.A)
		if !ok || len(coord) < 2 {
			return nil, fmt.Errorf("position %d is not [longitude, latitude]", k)
		}
		lon, err := toFloat64(coord[0])
		if err != nil {
			return nil, fmt.Errorf("position %d longitude: %w", k, err)
		}
		lat, err := toFloat64(coord[1])
		if err != nil {
			return nil, fmt.Errorf("position %d latitude: %w", k, err)
		}
		points = append(points, IsPointInsidePolygon.Point{X: lat, Y: lon})
	}
	return points, nil
}

func getValueByKey(d primitive.D, key string) interface{} {
	for _, e := range d {
		if e.Key == key {
			return e.Value
		}
	}
	return nil
}

func toFloat64(v interface{}) (float64, error) {
	var f float64
	switch val := v.(type) {
	case float64:
		f = val
	case float32:
		f = float64(val)
	case int32:
		f = float64(val)
	case int64:
		f = float64(val)
	case int:
		f = float64(val)
	default:
		return 0, fmt.Errorf("unsupported type %T for coordinate value", v)
	}
	if math.IsNaN(f) || math.IsInf(f, 0) {
		return 0, fmt.Errorf("coordinate value %v is not finite", f)
	}
	return f, nil
}
