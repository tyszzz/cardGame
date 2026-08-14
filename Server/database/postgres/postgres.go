package postgres

import (
	"cardGame/global"
	"cardGame/utils"
	"fmt"
	"time"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

func Init() *gorm.DB {
	var (
		conf = global.GameConf
		dsn  = fmt.Sprintf(
			"host=%s user=%s password=%s dbname=%s port=%s sslmode=disable TimeZone=Asia/Shanghai",
			conf.Postgres.Host,
			conf.Postgres.User,
			conf.Postgres.Password,
			conf.Postgres.DBName,
			conf.Postgres.Port,
		)
	)
	db, _err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if _err != nil {
		panic(fmt.Sprintf("failed to connect database %v", _err))
	}
	sqlDB, err := db.DB()
	if err != nil {
		panic(fmt.Sprintf("failed to get database instance %v", err))
	}

	if err := sqlDB.Ping(); err != nil {
		panic(fmt.Sprintf("failed to ping database %v", err))
	}

	sqlDB.SetMaxOpenConns(50)
	sqlDB.SetMaxIdleConns(10)
	sqlDB.SetConnMaxLifetime(time.Hour)
	sqlDB.SetConnMaxIdleTime(10 * time.Minute)

	utils.Log().Infof("database connected: host=%s port=%s dbname=%s", conf.Postgres.Host, conf.Postgres.Port, conf.Postgres.DBName)
	return db
}
