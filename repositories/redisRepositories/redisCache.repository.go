package redisRepositories

import (
	"encoding/json"
	"fmt"
	geofenceHistoryRepositories "github.com/marine-br/geoafence-service/repositories/geofenceHistories"
	"github.com/marine-br/golib-logger/logger"
	golib_redis "github.com/marine-br/golib-redis"
	"time"
)

type RedisCacheRepository struct {
	db *golib_redis.RedisClient
}

func NewRedisCacheRepository(redis *golib_redis.RedisClient) *RedisCacheRepository {
	return &RedisCacheRepository{
		db: redis,
	}
}

func (r *RedisCacheRepository) GetLastGeofence(param GetLastGeofenceParams) (*geofenceHistoryRepositories.GeofenceHistoryModel, error) {
	key := fmt.Sprintf("%s:%s", param.VehicleID.Hex(), param.GeofenceId.Hex())

	lastMessage, err := r.db.Get(key)
	if err != nil {
		logger.Error(fmt.Sprintf("cant get redis value from key [%s] %s", key, err.Error()))
		return nil, err
	}

	var geofenceHistory geofenceHistoryRepositories.GeofenceHistoryModel
	err = json.Unmarshal([]byte(lastMessage), &geofenceHistory)
	if err != nil {
		logger.Error(fmt.Sprintf("cant unmarshal redis value from key [%s] %s", key, err.Error()))
		return nil, err
	}

	return &geofenceHistory, nil
}

func (r *RedisCacheRepository) SetLastGeofence(param SetLastGeofenceParams) bool {
	key := fmt.Sprintf("%s:%s", param.VehicleID.Hex(), param.GeofenceId.Hex())

	err := r.db.Set(key, param.Value, 5*24*time.Hour)
	if err != nil {
		logger.Error(fmt.Sprintf("cant set redis key [%s] %s", key, err.Error()))
		return false
	}

	return true
}
