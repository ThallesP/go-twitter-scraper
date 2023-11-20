package database

import (
	"os"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

var dbInstance *gorm.DB

func GetClientOrPanic() *gorm.DB {
	if dbInstance != nil {
		return dbInstance
	}

	db, err := gorm.Open(postgres.Open(os.Getenv("DATABASE_DSN")), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Silent),
	})

	if err != nil {
		panic("failed to connect database") // Consider not using panic anywhere in the code
	}

	dbInstance = db

	err = dbInstance.AutoMigrate(&TweetModel{}, &OpenAccountModel{})

	if err != nil {
		panic("failed to migrate database")
	}

	return dbInstance
}
