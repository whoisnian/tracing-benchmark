package global

import (
	"time"

	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

var DB *gorm.DB

const (
	poolMaxIdleConns = 8
	poolMaxOpenConns = 32
	poolMaxLifetime  = time.Hour
)

func SetupMySQL() {
	var err error
	DB, err = gorm.Open(mysql.New(mysql.Config{DSN: CFG.MysqlDsn}), &gorm.Config{})
	if err != nil {
		panic(err)
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
