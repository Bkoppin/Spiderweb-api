package postgres

import (
	"api/internal/app/models"
	"os"

	"github.com/joho/godotenv"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

// Connect is a function that connects to the database
func Connect() (*gorm.DB, error) {
	err := godotenv.Load()
	if err != nil {
		return nil, err
	}
	dbName := os.Getenv("POSTGRES_URI")
	db, err := gorm.Open(postgres.Open(dbName), &gorm.Config{})
	if err != nil {
		return nil, err
	}
	db.AutoMigrate(&models.User{})
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
