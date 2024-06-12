package eventValidators

import (
	"fmt"
	geofenceHistoryRepositories "github.com/marine-br/geoafence-service/repositories/geofenceHistories"
	"github.com/marine-br/geoafence-service/repositories/redisRepositories"
	"github.com/marine-br/golib-logger/logger"
	"github.com/marine-br/golib-utils/models"
)

const (
	StatusIn           = "IN"
	StatusOut          = "OUT"
	StatusBeforeCache  = "BEFORE_CACHE"
	StatusDuplicated   = "DUP"
	StatusSameAsBefore = "SAME_AS_BEFORE"
)

func ValidateFromCache(cache redisRepositories.CacheRepository, trackerMessage models.TrackerMsgType, geofence models.GeofenceType, status string) (string, bool) {
	LastGeofenceHistoryFromCache, _ := cache.GetLastGeofence(redisRepositories.GetLastGeofenceParams{
		VehicleID:  trackerMessage.VEHICLE,
		GeofenceId: geofence.ID,
	})

	if LastGeofenceHistoryFromCache == nil {
		return StatusBeforeCache, false
	}

	if LastGeofenceHistoryFromCache.Date.Equal(trackerMessage.GPS_TIME) {
		return LastGeofenceHistoryFromCache.Type, true
	}

	if LastGeofenceHistoryFromCache.Date.Before(trackerMessage.GPS_TIME) {
		return StatusBeforeCache, true
	}

	if LastGeofenceHistoryFromCache.Type == status {
		return StatusSameAsBefore, true
	}

	return status, true
}

func ValidateFromDatabase(
	geofenceRepo geofenceHistoryRepositories.GeofenceHistoryRepository,
	trackerMessage models.TrackerMsgType,
	geofence models.GeofenceType,
	status string,

	redisRepo redisRepositories.RedisCacheRepository,
	hasLastFromCache bool,
) string {

	LastGeofenceHistoryFromDb, _ := geofenceRepo.FindLastGeofenceHistory(geofenceHistoryRepositories.FindLastGeofenceHistoryParams{
		TrackerMessage: trackerMessage,
		Geofence:       geofence,
	})

	// sets the message on redis if there is none, preventing searchs on the database
	if (LastGeofenceHistoryFromDb != geofenceHistoryRepositories.GeofenceHistoryModel{}) && !hasLastFromCache {
		logger.LogWithLabel(trackerMessage.DID, "setting cache for redis")
		redisRepo.SetLastGeofence(redisRepositories.SetLastGeofenceParams{
			GeofenceId: LastGeofenceHistoryFromDb.Geofence,
			VehicleID:  LastGeofenceHistoryFromDb.Vehicle,
			Value:      LastGeofenceHistoryFromDb,
		})
	}

	// event already created
	if LastGeofenceHistoryFromDb.Date.Equal(trackerMessage.GPS_TIME) {
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

	if FirstAfterGeofenceHistoryFromDb.Date.Equal(trackerMessage.GPS_TIME) {
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
