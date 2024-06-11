package main

import (
	"encoding/json"
	"errors"
	"fmt"
	geofenceHistoryRepositories "github.com/marine-br/geoafence-service/repositories/geofenceHistories"
	"github.com/marine-br/geoafence-service/repositories/geofencesRepositories"
	"github.com/marine-br/geoafence-service/setups"
	"github.com/marine-br/geoafence-service/utils/IsPointInsidePolygon"
	"github.com/marine-br/geoafence-service/utils/PolygonFromPrimitiveD"
	"github.com/marine-br/golib-logger/logger"
	"github.com/marine-br/golib-utils/models"
	"github.com/marine-br/golib-utils/utils"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"net/http"
	"os"
)

func main() {
	setups.SetupEnv()
	rabbitmq := setups.SetupRabbitmq()
	mongo := setups.SetupMongo()
	geofenceRepository := geofencesRepositories.NewMongoGeofenceRepository(mongo)
	geofenceHistoryRepository := geofenceHistoryRepositories.NewMongoGeofenceHistoryRepository(mongo)

	go func() {
		//http.Handle("/metrics", promhttp.Handler())
		http.HandleFunc(
			"/health", utils.HealthCheckHandler(
				utils.HealthCheckArgs{
					MongoClient: &mongo,
				},
			),
		)

		logger.Log("server listening on port ", os.Getenv("HTTP_SERVER_PORT"))
		if err := http.ListenAndServe(os.Getenv("HTTP_SERVER_PORT"), nil); !errors.Is(err, http.ErrServerClosed) {
			logger.LogError(err)
		}
	}()

	forever := make(chan bool)
	go func() {
		for {
			message := <-rabbitmq

			var trackerMessage models.TrackerMsgType
			err := json.Unmarshal(message.Body, &trackerMessage)
			if err != nil {
				logger.Error(string(message.Body))
				logger.Error("cant unmarshal tracker message, ack sent", err)
				err := message.Ack(true)
				if err != nil {
					logger.Error("cant ack message", err)
					return
				}

				continue
			}

			geoFences, err := geofenceRepository.GetGeofences(geofencesRepositories.GetGeofenceParams{CompanyId: trackerMessage.COMPANY})
			if err != nil {
				logger.Error("cant get geofences from company", err)
				message.Ack(true)
				continue
			}

			logger.Log(fmt.Sprintf("found %d geofences", len(geoFences)))

			// para cada geofence, valida se o vehicle está dentro da geofence

			var inCounter int
			for _, geoFence := range geoFences {
				geojson, ok := geoFence.Geojson.(primitive.D)
				if !ok {
					logger.Error("geojson is not a primitive.D")
					continue
				}

				geofencePolygon := PolygonFromPrimitiveD.PolygonFromPrimitiveD(geojson)
				point := IsPointInsidePolygon.Point{X: trackerMessage.LATITUDE, Y: trackerMessage.LONGITUDE}

				status := IsPointInsidePolygon.IsPointInPolygon(point, geofencePolygon)

				stringStatus := "OUT"
				if status {
					stringStatus = "IN"
					inCounter += 1
				}

				err = geofenceHistoryRepository.InsertGeofenceHistory(geofenceHistoryRepositories.InsertGeofenceHistoryParams{
					TrackerMessage: trackerMessage,
					Geofence:       geoFence,
					Status:         stringStatus,
				})
				if err != nil {
					logger.Error("cant insert in the database", err)
					continue
				}
			}
			logger.Log("in counter", inCounter)
			logger.Log("out counter", len(geoFences)-inCounter)

			message.Ack(true)
		}
	}()
	<-forever
}
