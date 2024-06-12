package setups

import (
	golib_redis "github.com/marine-br/golib-redis"
	"os"
)

func SetupRedis() *golib_redis.RedisClient {
	client := &golib_redis.RedisClient{}
	if err := client.Connect(); err != nil {
		os.Exit(1)
	}

	return client
}
