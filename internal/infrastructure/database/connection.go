package database

import (
	"time"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

func Connect(databaseURL string) (*gorm.DB, error) {
	return gorm.Open(postgres.Open(databaseURL), &gorm.Config{})
}

func timeFromUnix(timestamp int64) time.Time {
	return time.Unix(timestamp, 0)
}
