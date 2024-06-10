package geofencesRepositories

import (
	"fmt"
	dotenv "github.com/joho/godotenv"
	"github.com/marine-br/geoafence-service/setups"
	"github.com/marine-br/golib-utils/models"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"reflect"
	"testing"
)

func TestMongoGeofenceRepository_GetGeofences(t *testing.T) {
	dotenv.Load("../../.env")
	mongoClient := setups.SetupMongo()
	repo := NewMongoGeofenceRepository(mongoClient)
	id, _ := primitive.ObjectIDFromHex("5e829fa80c502411fbb06c31")

	type fields struct {
		GeofenceCollection *mongo.Collection
	}
	type args struct {
		param GetGeofenceParams
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    []models.GeofenceType
		wantErr bool
	}{
		{
			name:   "get geofences",
			fields: fields{},
			args: args{
				param: GetGeofenceParams{
					CompanyId: id,
				},
			},
			want: []models.GeofenceType{},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, _ := repo.GetGeofences(tt.args.param)
			for _, i := range got {
				fmt.Println(reflect.TypeOf(i.Geojson))
			}
			//if (err != nil) != tt.wantErr {
			//	t.Errorf("GetGeofences() error = %v, wantErr %v", err, tt.wantErr)
			//	return
			//}
			//if !reflect.DeepEqual(got, tt.want) {
			//	t.Errorf("GetGeofences() got = %v, want %v", got, tt.want)
			//}
		})
	}
}
