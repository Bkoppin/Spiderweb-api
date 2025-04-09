package postgres

import (
	"api/internal/app/models"
	"os"

	"github.com/joho/godotenv"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

/*
Connect initializes a new PostgreSQL database connection using environment variables.
It loads the database connection details from a .env file and returns a gorm.DB instance or an error if the connection fails.
The .env file should contain the following variable:
  - POSTGRES_URI: The URI of the PostgreSQL database.
*/
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

/*
Close closes the PostgreSQL database connection.
It retrieves the underlying SQL database connection from the gorm.DB instance and closes it.
It returns an error if the closing operation fails.
*/
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
