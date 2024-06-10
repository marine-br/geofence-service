package setups

import (
	"github.com/marine-br/golib-logger/logger"
	db "github.com/marine-br/golib-mongo"
)

func SetupMongo() (mongoClient db.MongoClient) {
	err := mongoClient.Connect()
	logger.PanicError(err)

	return mongoClient
}
