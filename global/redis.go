package global

import (
	"github.com/redis/go-redis/extra/redisotel/v9"
	"github.com/redis/go-redis/v9"
	"github.com/whoisnian/tracing-benchmark/pkg/apmredis"
	"github.com/whoisnian/tracing-benchmark/pkg/zipkinredis"
)

var RDB *redis.Client

func SetupRedis() {
	opts, err := redis.ParseURL(CFG.RedisUri)
	if err != nil {
		panic(err)
	}

	RDB = redis.NewClient(opts)
	switch CFG.TraceBackend {
	case "otlp":
		err = redisotel.InstrumentTracing(RDB)
	case "apm":
		RDB.AddHook(apmredis.NewHook())
	case "zipkin":
		RDB.AddHook(zipkinredis.NewHook())
	}
	if err != nil {
		panic(err)
	}
}
