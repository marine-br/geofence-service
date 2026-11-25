package PolygonFromPrimitiveD

import (
	"os"
	"os/exec"
	"testing"

	"github.com/marine-br/geoafence-service/utils/IsPointInsidePolygon"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

// A fence stored as a bare Feature (accepted by the API) has no "features".
func TestBareFeatureDoesNotKillTheProcess(t *testing.T) {
	if os.Getenv("REPRO_CHILD") == "1" {
		PolygonFromPrimitiveD(primitive.D{
			{Key: "type", Value: "Feature"},
			{Key: "geometry", Value: primitive.D{{Key: "type", Value: "Polygon"}, {Key: "coordinates", Value: primitive.A{}}}},
		})
		return
	}
	cmd := exec.Command(os.Args[0], "-test.run=TestBareFeatureDoesNotKillTheProcess")
	cmd.Env = append(os.Environ(), "REPRO_CHILD=1")
	if out, err := cmd.CombinedOutput(); err != nil {
		t.Fatalf("process died: %v\n%s", err, out)
	}
}

func TestMultiPolygonContainsPoint(t *testing.T) {
	square := primitive.A{primitive.A{
		primitive.A{-48.0, -27.0}, primitive.A{-47.0, -27.0}, primitive.A{-47.0, -26.0}, primitive.A{-48.0, -26.0}, primitive.A{-48.0, -27.0},
	}}
	geojson := primitive.D{
		{Key: "type", Value: "FeatureCollection"},
		{Key: "features", Value: primitive.A{primitive.D{
			{Key: "type", Value: "Feature"},
			{Key: "geometry", Value: primitive.D{{Key: "type", Value: "MultiPolygon"}, {Key: "coordinates", Value: primitive.A{square}}}},
		}}},
	}
	polygon := PolygonFromPrimitiveD(geojson)
	if !IsPointInsidePolygon.IsPointInPolygon(IsPointInsidePolygon.Point{X: -26.5, Y: -47.5}, polygon) {
		t.Fatal("point inside the MultiPolygon counted as outside")
	}
}
