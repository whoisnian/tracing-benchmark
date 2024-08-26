package global

import (
	"github.com/redis/go-redis/v9"
)

var RDB *redis.Client

func SetupRedis() {
	opts, err := redis.ParseURL(CFG.RedisUri)
	if err != nil {
		panic(err)
	}
	RDB = redis.NewClient(opts)
}
