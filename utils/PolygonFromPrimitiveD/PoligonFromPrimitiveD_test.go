package PolygonFromPrimitiveD

import (
	"github.com/marine-br/geoafence-service/utils/IsPointInsidePolygon"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"reflect"
	"testing"
)

func TestRemovePolygon(t *testing.T) {
	geojson1 := primitive.D{primitive.E{Key: "type", Value: "FeatureCollection"}, primitive.E{Key: "features", Value: primitive.A{primitive.D{primitive.E{Key: "id", Value: "a094593b18cb46db09d1890ae7ccd47a"}, primitive.E{Key: "type", Value: "Feature"}, primitive.E{Key: "properties", Value: primitive.D{}}, primitive.E{Key: "geometry", Value: primitive.D{primitive.E{Key: "coordinates", Value: primitive.A{primitive.A{primitive.A{-48.60565411678576, -27.600050127488487}, primitive.A{-48.605719721111626, -27.597440868325883}, primitive.A{-48.60728492673141, -27.596360196755654}, primitive.A{-48.60793904044644, -27.597087580984365}, primitive.A{-48.60862281774649, -27.5979069757488}, primitive.A{-48.60705325170399, -27.598984360719403}, primitive.A{-48.60665880670879, -27.599249422897003}, primitive.A{-48.606489223861864, -27.60005403460859}, primitive.A{-48.60606511528023, -27.600049176486444}, primitive.A{-48.60565411678576, -27.600050127488487}}}}, primitive.E{Key: "type", Value: "Polygon"}}}}}}}
	want1 := []IsPointInsidePolygon.Point{IsPointInsidePolygon.Point{Y: -48.60565411678576, X: -27.600050127488487}, IsPointInsidePolygon.Point{Y: -48.605719721111626, X: -27.597440868325883}, IsPointInsidePolygon.Point{Y: -48.60728492673141, X: -27.596360196755654}, IsPointInsidePolygon.Point{Y: -48.60793904044644, X: -27.597087580984365}, IsPointInsidePolygon.Point{Y: -48.60862281774649, X: -27.5979069757488}, IsPointInsidePolygon.Point{Y: -48.60705325170399, X: -27.598984360719403}, IsPointInsidePolygon.Point{Y: -48.60665880670879, X: -27.599249422897003}, IsPointInsidePolygon.Point{Y: -48.606489223861864, X: -27.60005403460859}, IsPointInsidePolygon.Point{Y: -48.60606511528023, X: -27.600049176486444}, IsPointInsidePolygon.Point{Y: -48.60565411678576, X: -27.600050127488487}}
	type args struct {
		geojson primitive.D
	}
	tests := []struct {
		name string
		args args
		want []IsPointInsidePolygon.Point
	}{
		{
			name: "test 1",
			args: args{
				geojson: geojson1,
			},
			want: want1,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := PolygonFromPrimitiveD(tt.args.geojson)

			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("PolygonFromPrimitiveD() got1 = %#v, want %v", got, tt.want)
			}
		})
	}
}
