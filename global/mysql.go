package global

import (
	"time"

	"go.elastic.co/apm/module/apmsql/v2"
	_ "go.elastic.co/apm/module/apmsql/v2/mysql"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"gorm.io/plugin/opentelemetry/tracing"
)

var DB *gorm.DB

const (
	poolMaxIdleConns = 8
	poolMaxOpenConns = 32
	poolMaxLifetime  = time.Hour
)

func SetupMySQL() {
	driverName := mysql.DefaultDriverName
	if CFG.TraceBackend == "apm" {
		driverName = apmsql.DriverPrefix + mysql.DefaultDriverName
	}

	var err error
	DB, err = gorm.Open(mysql.New(mysql.Config{DriverName: driverName, DSN: CFG.MysqlDsn}), &gorm.Config{})
	if err != nil {
		panic(err)
	}

	if CFG.TraceBackend == "otlp" {
		if err = DB.Use(tracing.NewPlugin(tracing.WithoutMetrics())); err != nil {
			panic(err)
		}
	}

	sqlDB, err := DB.DB()
	if err != nil {
		panic(err)
	}
	sqlDB.SetMaxIdleConns(poolMaxIdleConns)
	sqlDB.SetMaxOpenConns(poolMaxOpenConns)
	sqlDB.SetConnMaxLifetime(poolMaxLifetime)

	if err = sqlDB.Ping(); err != nil {
		panic(err)
	}
}
