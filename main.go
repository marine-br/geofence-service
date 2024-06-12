package main

import (
	"encoding/json"
	"errors"
	"fmt"
	geofenceHistoryRepositories "github.com/marine-br/geoafence-service/repositories/geofenceHistories"
	"github.com/marine-br/geoafence-service/repositories/geofencesRepositories"
	"github.com/marine-br/geoafence-service/repositories/redisRepositories"
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

const (
	StatusIn           = "IN"
	StatusOut          = "OUT"
	StatusBeforeCache  = "BEFORE_CACHE"
	StatusDuplicated   = "DUP"
	StatusSameAsBefore = "SAME_AS_BEFORE"
)

func main() {
	setups.SetupEnv()
	rabbitmq := setups.SetupRabbitmq()
	mongo := setups.SetupMongo()
	redis := setups.SetupRedis()

	redisRepo := redisRepositories.NewRedisCacheRepository(redis)
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

				// set the default value for the status
				statusString := StatusOut
				if status {
					statusString = StatusIn
				}

				// change the status based on cache
				statusString = ValidateFromCache(redisRepo, trackerMessage, statusString)
				if statusString == StatusBeforeCache {
					// change the status based on the database
					statusString = ValidateFromDatabase(
						&geofenceHistoryRepository,
						trackerMessage,
						geoFence,
						statusString,
					)
				}
				if statusString == StatusSameAsBefore {
					logger.Warning("was not save on the database")
					continue
				}

				if statusString == StatusDuplicated {
					logger.Warning("message duplicated from the databse")
					continue
				}

				// create a new history for the message
				err = geofenceHistoryRepository.InsertGeofenceHistory(geofenceHistoryRepositories.InsertGeofenceHistoryParams{
					TrackerMessage: trackerMessage,
					Geofence:       geoFence,
					Status:         statusString,
				})
				if err != nil {
					logger.Error("cant insert in the database", err)
					continue
				}

				if statusString == StatusIn {
					inCounter += 1
				}
			}
			logger.Log("in counter", inCounter)
			logger.Log("out counter", len(geoFences)-inCounter)

			message.Ack(true)
		}
	}()
	<-forever
}

func ValidateFromCache(cache redisRepositories.CacheRepository, trackerMessage models.TrackerMsgType, status string) string {
	LastGeofenceHistoryFromCache, _ := cache.GetLastGeofence(redisRepositories.GetLastGeofenceParams{})

	if LastGeofenceHistoryFromCache == nil {
		return StatusBeforeCache
	}

	if LastGeofenceHistoryFromCache.CreatedAt.Equal(trackerMessage.GPS_TIME) {
		return LastGeofenceHistoryFromCache.Type
	}

	if LastGeofenceHistoryFromCache.CreatedAt.After(trackerMessage.GPS_TIME) {
		return StatusBeforeCache
	}

	if LastGeofenceHistoryFromCache.Type == StatusOut {
		status = StatusIn
	}

	return status
}

func ValidateFromDatabase(
	geofenceRepo geofenceHistoryRepositories.GeofenceHistoryRepository,
	trackerMessage models.TrackerMsgType,
	geofence models.GeofenceType,
	status string) string {

	LastGeofenceHistoryFromDb, _ := geofenceRepo.FindLastGeofenceHistory(geofenceHistoryRepositories.FindLastGeofenceHistoryParams{
		TrackerMessage: trackerMessage,
		Geofence:       geofence,
	})

	// event already created
	if LastGeofenceHistoryFromDb.CreatedAt.Equal(trackerMessage.GPS_TIME) {
		return StatusDuplicated
	}

	// if it was in the same status,
	if LastGeofenceHistoryFromDb.Type == status {
		return StatusSameAsBefore
	}

	FirstAfterGeofenceHistoryFromDb, _ := geofenceRepo.FindFirstAfterGeofenceHistory(geofenceHistoryRepositories.FindFirstAfterGeofenceHistoryParams{
		TrackerMessage: trackerMessage,
		Geofence:       geofence,
	})

	if FirstAfterGeofenceHistoryFromDb == nil {
		return status
	}

	if FirstAfterGeofenceHistoryFromDb.CreatedAt.Equal(trackerMessage.GPS_TIME) {
		return StatusDuplicated
	}

	if FirstAfterGeofenceHistoryFromDb.Type == status {
		err := geofenceRepo.DeleteGeofenceHistory(
			geofenceHistoryRepositories.DeleteGeofenceHistoryParams{
				GeofenceHistoryId: FirstAfterGeofenceHistoryFromDb.ID,
			},
		)
		if err != nil {
			logger.Error(fmt.Sprintf("cant delete geofence from db [%s] %s", FirstAfterGeofenceHistoryFromDb.ID.Hex(), err))
		}

		return status
	}

	return status
}
