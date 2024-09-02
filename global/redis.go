package global

import (
	"github.com/redis/go-redis/extra/redisotel/v9"
	"github.com/redis/go-redis/v9"
)

var RDB *redis.Client

func SetupRedis() {
	opts, err := redis.ParseURL(CFG.RedisUri)
	if err != nil {
		panic(err)
	}

	RDB = redis.NewClient(opts)
	if err := redisotel.InstrumentTracing(RDB); err != nil {
		panic(err)
	}
}
