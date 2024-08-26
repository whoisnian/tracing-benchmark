package global

import (
	"os"
	"strconv"
)

var CFG Config

type Config struct {
	Version bool // show version and quit

	ListenAddr string // http server listen addr
	MysqlDsn   string // mysql dsn from https://github.com/go-sql-driver/mysql/blob/master/README.md#dsn-data-source-name
	RedisUri   string // redis uri from https://github.com/redis/redis-specifications/blob/master/uri/redis.txt
}

func SetupConfig() {
	CFG.Version = boolFromEnv("CFG_VERSION", false)

	CFG.ListenAddr = stringFromEnv("CFG_LISTEN_ADDR", "127.0.0.1:8080")
	CFG.MysqlDsn = stringFromEnv("CFG_MYSQL_DSN", "root:password@tcp(127.0.0.1:3306)/dbname?charset=utf8mb4&parseTime=True&loc=UTC")
	CFG.RedisUri = stringFromEnv("CFG_REDIS_URI", "redis://default:password@127.0.0.1:6379/0")
}

func boolFromEnv(envKey string, defVal bool) bool {
	if str, ok := os.LookupEnv(envKey); !ok {
		return defVal
	} else if val, err := strconv.ParseBool(str); err == nil {
		return val
	} else {
		panic(err)
	}
}

func stringFromEnv(envKey string, defVal string) string {
	if val, ok := os.LookupEnv(envKey); !ok {
		return defVal
	} else {
		return val
	}
}
