package database

import (
	"os"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

var dbInstance *gorm.DB

func GetClientOrPanic() *gorm.DB {
	if dbInstance != nil {
		return dbInstance
	}

	db, err := gorm.Open(postgres.Open(os.Getenv("DATABASE_DSN")), &gorm.Config{})

	if err != nil {
		panic("failed to connect database")
	}

	dbInstance = db

	err = dbInstance.AutoMigrate(&TweetModel{}, &OpenAccountModel{})

	if err != nil {
		panic("failed to migrate database")
	}

	return dbInstance
}
