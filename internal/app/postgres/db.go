package db

import (
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

// Connect is a function that connects to the database
func Connect() (*gorm.DB, error) {
	dsn := "host=localhost user=postgres password=postgres dbname=spiderweb port=5432 sslmode=disable"
	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		return nil, err
	}
	return db, nil
}

// Close is a function that closes the database connection
func Close(db *gorm.DB) error {
	dbSQL, err := db.DB()
	if err != nil {
		return err
	}
	err = dbSQL.Close()
	if err != nil {
		return err
	}
	return nil
}

