package PolygonFromPrimitiveD

import (
	"errors"
	"math"
	"testing"

	"github.com/marine-br/geoafence-service/utils/IsPointInsidePolygon"
	"github.com/marine-br/golib-utils/models"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

// square returns the rings of a closed square polygon, [lon, lat] order.
func square(minLon, minLat, maxLon, maxLat float64) primitive.A {
	return primitive.A{primitive.A{
		primitive.A{minLon, minLat}, primitive.A{maxLon, minLat}, primitive.A{maxLon, maxLat},
		primitive.A{minLon, maxLat}, primitive.A{minLon, minLat},
	}}
}

func feature(geometryType string, coordinates any) primitive.D {
	return primitive.D{
		{Key: "type", Value: "Feature"},
		{Key: "properties", Value: primitive.D{}},
		{Key: "geometry", Value: primitive.D{{Key: "type", Value: geometryType}, {Key: "coordinates", Value: coordinates}}},
	}
}

func collection(features ...any) primitive.D {
	return primitive.D{{Key: "type", Value: "FeatureCollection"}, {Key: "features", Value: primitive.A(features)}}
}

// point builds the tracker position with the service convention X=lat, Y=lon.
func point(lon, lat float64) IsPointInsidePolygon.Point {
	return IsPointInsidePolygon.Point{X: lat, Y: lon}
}

func contains(t *testing.T, geojson primitive.D, p IsPointInsidePolygon.Point) bool {
	t.Helper()
	polygons, err := PolygonsFromPrimitiveD(geojson)
	if err != nil {
		t.Fatalf("PolygonsFromPrimitiveD() error = %v", err)
	}
	return IsPointInsidePolygon.IsPointInAnyPolygon(p, polygons)
}

func TestMalformedGeojsonReturnsErrorInsteadOfExiting(t *testing.T) {
	tests := map[string]primitive.D{
		"bare Feature":                feature("Polygon", square(-48, -27, -47, -26)),
		"bare Polygon":                {{Key: "type", Value: "Polygon"}, {Key: "coordinates", Value: square(-48, -27, -47, -26)}},
		"features is not array":       {{Key: "type", Value: "FeatureCollection"}, {Key: "features", Value: "x"}},
		"empty features":              collection(),
		"empty coordinates":           collection(feature("Polygon", primitive.A{})),
		"empty ring":                  collection(feature("Polygon", primitive.A{primitive.A{}})),
		"coordinates not array":       collection(feature("Polygon", "x")),
		"ring not array":              collection(feature("Polygon", primitive.A{"x"})),
		"string coordinate":           collection(feature("Polygon", primitive.A{primitive.A{primitive.A{"-48", -27.0}, primitive.A{-47.0, -27.0}, primitive.A{-47.0, -26.0}, primitive.A{"-48", -27.0}}})),
		"null coordinate":             collection(feature("Polygon", primitive.A{primitive.A{primitive.A{nil, -27.0}, primitive.A{-47.0, -27.0}, primitive.A{-47.0, -26.0}, primitive.A{nil, -27.0}}})),
		"position with one value":     collection(feature("Polygon", primitive.A{primitive.A{primitive.A{-48.0}, primitive.A{-47.0, -27.0}, primitive.A{-47.0, -26.0}, primitive.A{-48.0}}})),
		"NaN coordinate":              collection(feature("Polygon", primitive.A{primitive.A{primitive.A{math.NaN(), -27.0}, primitive.A{-47.0, -27.0}, primitive.A{-47.0, -26.0}, primitive.A{math.NaN(), -27.0}}})),
		"only a Point feature":        collection(feature("Point", primitive.A{-48.0, -27.0})),
		"MultiPolygon item not array": collection(feature("MultiPolygon", primitive.A{"x"})),
	}
	for name, geojson := range tests {
		t.Run(name, func(t *testing.T) {
			polygons, err := PolygonsFromPrimitiveD(geojson)
			if err == nil {
				t.Fatalf("want error, got %d polygons", len(polygons))
			}
		})
	}
}

func TestNoAreaIsErrNoPolygon(t *testing.T) {
	_, err := PolygonsFromPrimitiveD(collection(feature("LineString", primitive.A{primitive.A{-48.0, -27.0}, primitive.A{-47.0, -26.0}})))
	if !errors.Is(err, ErrNoPolygon) {
		t.Fatalf("err = %v, want ErrNoPolygon", err)
	}
}

func TestMultiPolygonContainsPointInAnyPolygon(t *testing.T) {
	geojson := collection(feature("MultiPolygon", primitive.A{square(-48, -27, -47, -26), square(-40, -20, -39, -19)}))

	if !contains(t, geojson, point(-47.5, -26.5)) {
		t.Error("point inside the first polygon counted as outside")
	}
	if !contains(t, geojson, point(-39.5, -19.5)) {
		t.Error("point inside the second polygon counted as outside")
	}
	if contains(t, geojson, point(-44, -23)) {
		t.Error("point between the polygons counted as inside")
	}
}

func TestFeatureCollectionUsesEveryPolygonFeature(t *testing.T) {
	geojson := collection(
		feature("Point", primitive.A{0.0, 0.0}),
		feature("Polygon", square(-48, -27, -47, -26)),
		feature("Polygon", square(-40, -20, -39, -19)),
	)
	if !contains(t, geojson, point(-39.5, -19.5)) {
		t.Error("point inside the second Polygon feature counted as outside")
	}
	if !contains(t, geojson, point(-47.5, -26.5)) {
		t.Error("point inside the first Polygon feature counted as outside")
	}
}

func TestHolesAreIgnored(t *testing.T) {
	outer := square(-48, -27, -47, -26)[0]
	hole := square(-47.6, -26.6, -47.4, -26.4)[0]
	geojson := collection(feature("Polygon", primitive.A{outer, hole}))
	if !contains(t, geojson, point(-47.5, -26.5)) {
		t.Error("point inside a hole must still count as inside (outer ring only)")
	}
}

func TestDegenerateRingIsSkippedNotFatal(t *testing.T) {
	line := primitive.A{primitive.A{primitive.A{-48.0, -27.0}, primitive.A{-47.0, -26.0}}}
	geojson := collection(feature("Polygon", line), feature("Polygon", square(-40, -20, -39, -19)))
	polygons, err := PolygonsFromPrimitiveD(geojson)
	if err != nil || len(polygons) != 1 {
		t.Fatalf("got %d polygons, err %v; want only the valid square", len(polygons), err)
	}
}

// A ring with three or more points that does not come back to its first
// position is malformed (GeoJSON requires the ring to be closed), so the fence
// is rejected instead of being evaluated with an implicit closing edge.
func TestUnclosedRingIsRejected(t *testing.T) {
	unclosedSquare := primitive.A{
		primitive.A{-48.0, -27.0}, primitive.A{-47.0, -27.0}, primitive.A{-47.0, -26.0}, primitive.A{-48.0, -26.0},
	}
	unclosedTriangle := primitive.A{
		primitive.A{-48.0, -27.0}, primitive.A{-47.0, -27.0}, primitive.A{-47.0, -26.0},
	}
	wrongClosingPoint := primitive.A{
		primitive.A{-48.0, -27.0}, primitive.A{-47.0, -27.0}, primitive.A{-47.0, -26.0}, primitive.A{-48.0, -26.5},
	}
	tests := map[string]primitive.D{
		"Polygon missing the closing point":             collection(feature("Polygon", primitive.A{unclosedSquare})),
		"MultiPolygon member missing the closing point": collection(feature("MultiPolygon", primitive.A{primitive.A{unclosedSquare}})),
		"triangle missing the closing point":            collection(feature("Polygon", primitive.A{unclosedTriangle})),
		"closing point with a different value":          collection(feature("Polygon", primitive.A{wrongClosingPoint})),
	}
	for name, geojson := range tests {
		t.Run(name, func(t *testing.T) {
			polygons, err := PolygonsFromPrimitiveD(geojson)
			if err == nil {
				t.Fatalf("want an error for an unclosed ring, got %d polygons", len(polygons))
			}
			if polygons != nil {
				t.Fatalf("want nil polygons with the error, got %v", polygons)
			}
		})
	}

	t.Run("closed ring stays valid", func(t *testing.T) {
		polygons, err := PolygonsFromPrimitiveD(collection(feature("Polygon", square(-48, -27, -47, -26))))
		if err != nil || len(polygons) != 1 {
			t.Fatalf("closed ring: %d polygons, err %v", len(polygons), err)
		}
	})
}

// The fence comes from Mongo decoded into models.GeofenceType.Geojson (any):
// documents arrive as primitive.D, arrays as primitive.A and integer
// coordinates as int32. This runs the real decoding path.
func TestDecodedFromBSON(t *testing.T) {
	raw, err := bson.Marshal(bson.M{
		"name": "dummy-fence",
		"geojson": bson.M{
			"type": "FeatureCollection",
			"features": bson.A{bson.M{
				"type": "Feature",
				"geometry": bson.M{
					"type":        "MultiPolygon",
					"coordinates": bson.A{bson.A{bson.A{bson.A{-48, -27}, bson.A{-47, -27}, bson.A{-47, -26}, bson.A{-48, -26}, bson.A{-48, -27}}}},
				},
			}},
		},
	})
	if err != nil {
		t.Fatal(err)
	}
	var fence models.GeofenceType
	if err := bson.Unmarshal(raw, &fence); err != nil {
		t.Fatal(err)
	}
	geojson, ok := fence.Geojson.(primitive.D)
	if !ok {
		t.Fatalf("geojson decoded as %T", fence.Geojson)
	}
	if !contains(t, geojson, point(-47.5, -26.5)) {
		t.Error("point inside the decoded MultiPolygon counted as outside")
	}
}
