package database

import (
	"os"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

func GetClientOrPanic() *gorm.DB {
	db, err := gorm.Open(postgres.Open(os.Getenv("DATABASE_DSN")), &gorm.Config{})

	if err != nil {
		panic("failed to connect database")
	}

	err = db.AutoMigrate(&TweetModel{}, &OpenAccountModel{})

	if err != nil {
		panic("failed to migrate database")
	}

	return db
}
