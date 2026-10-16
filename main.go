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
	"github.com/marine-br/geoafence-service/utils/eventValidators"
	"github.com/marine-br/golib-logger/logger"
	"github.com/marine-br/golib-utils/models"
	"github.com/marine-br/golib-utils/utils"
	"github.com/marine-br/golib-utils/utils/metrics"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"net/http"
	"os"
	"time"
)

var (
	errInvalidMessage = errors.New("invalid tracker message")
	errGeofenceLookup = errors.New("geofence lookup failed")
	serviceMetrics    = metrics.New("geofence-service")
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
		mux := newHTTPMux(utils.HealthCheckHandler(
			utils.HealthCheckArgs{
				MongoClient: &mongo,
			},
		))

		logger.Log("server listening on port ", os.Getenv("HTTP_SERVER_PORT"))
		if err := http.ListenAndServe(os.Getenv("HTTP_SERVER_PORT"), mux); !errors.Is(err, http.ErrServerClosed) {
			logger.LogError(err)
		}
	}()

	queue := os.Getenv("RABBITMQ_CON_QUEUE")
	forever := make(chan bool)
	go func() {
		for {
			message, ok := <-rabbitmq
			if !ok {
				logger.Error("rabbitmq delivery channel closed, exiting for restart")
				os.Exit(1)
			}
			start := time.Now()

			var trackerMessage models.TrackerMsgType
			err := json.Unmarshal(message.Body, &trackerMessage)
			if err != nil {
				logger.Error(string(message.Body))
				logger.Error("cant unmarshal tracker message, ack sent", err)
				serviceMetrics.ObserveMessage(queue, message.Redelivered, start, errInvalidMessage)
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
				serviceMetrics.ObserveMessage(queue, message.Redelivered, start, errGeofenceLookup)
				message.Ack(true)
				continue
			}

			logger.LogWithLabel(trackerMessage.DID, fmt.Sprintf("found %d geofences", len(geoFences)))

			// para cada geofence, valida se o vehicle está dentro da geofence

			var inCounter int
			var outCounter int
			for _, geoFence := range geoFences {
				geojson, ok := geoFence.Geojson.(primitive.D)
				if !ok {
					logger.Error("geojson is not a primitive.D")
					continue
				}
				if trackerMessage.LATITUDE == 0 && trackerMessage.LONGITUDE == 0 {
					logger.Debug("Invalid tracker message: Latitude and Longitude are 0")
					continue
				}

				geofencePolygon := PolygonFromPrimitiveD.PolygonFromPrimitiveD(geojson)
				point := IsPointInsidePolygon.Point{X: trackerMessage.LATITUDE, Y: trackerMessage.LONGITUDE}

				status := IsPointInsidePolygon.IsPointInPolygon(point, geofencePolygon)
				// set the default value for the status
				statusString := eventValidators.StatusOut
				if status {
					statusString = eventValidators.StatusIn
				}

				// change the status based on cache
				fromRedisStatus, hasCache := eventValidators.ValidateFromCache(redisRepo, trackerMessage, geoFence, statusString)
				if fromRedisStatus == eventValidators.StatusBeforeCache {
					// change the status based on the database
					statusString = eventValidators.ValidateFromDatabase(
						&geofenceHistoryRepository,
						trackerMessage,
						geoFence,
						statusString,

						*redisRepo,
						hasCache,
					)
				} else {
					statusString = fromRedisStatus
				}

				if statusString == eventValidators.StatusSameAsBefore {
					continue
				}

				if statusString == eventValidators.StatusDuplicated {
					logger.Warning("message duplicated from the databse")
					continue
				}

				// create a new history for the message
				geofenceHistory, err := geofenceHistoryRepository.InsertGeofenceHistory(geofenceHistoryRepositories.InsertGeofenceHistoryParams{
					TrackerMessage: trackerMessage,
					Geofence:       geoFence,
					Status:         statusString,
				})
				if err != nil {
					logger.Error("cant insert in the database", err)
					continue
				}

				redisRepo.SetLastGeofence(redisRepositories.SetLastGeofenceParams{
					GeofenceId: geoFence.ID,
					VehicleID:  trackerMessage.VEHICLE,
					Value:      geofenceHistory,
				})

				if statusString == eventValidators.StatusIn {
					inCounter += 1
				} else {
					outCounter += 1
				}
			}

			if inCounter > 0 {
				logger.LogWithLabel(trackerMessage.DID, "in counter", inCounter)
			}

			if outCounter > 0 {
				logger.LogWithLabel(trackerMessage.DID, "out counter", outCounter)
			}

			serviceMetrics.ObserveMessage(queue, message.Redelivered, start, nil)
			message.Ack(true)
		}
	}()
	<-forever
}

// newHTTPMux serves the health check and /metrics on HTTP_SERVER_PORT, the
// port of the scrape annotation, without using http.DefaultServeMux.
func newHTTPMux(health http.HandlerFunc) *http.ServeMux {
	mux := http.NewServeMux()
	mux.HandleFunc("/health", health)
	mux.Handle("/metrics", serviceMetrics.Handler())
	return mux
}
