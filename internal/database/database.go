package database

import (
	"fmt"
	"log"
	"os"

	"VelocityDBGo/internal/models"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

var DB *gorm.DB

// ConnectDB initializes the database connection and runs auto-migration
func ConnectDB() {
	dsn := os.Getenv("DATABASE_URL")
	if dsn == "" {
		// Fallback to individual components if DATABASE_URL is not set
		dsn = fmt.Sprintf(
			"host=%s user=%s password=%s dbname=%s port=%s sslmode=disable TimeZone=UTC",
			os.Getenv("DB_HOST"),
			os.Getenv("DB_USER"),
			os.Getenv("DB_PASSWORD"),
			os.Getenv("DB_NAME"),
			os.Getenv("DB_PORT"),
		)
	}

	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		log.Fatalf("Failed to connect to database: %v\n", err)
	}

	log.Println("Successfully connected to the PostgreSQL database.")

	// Auto Migrate the domain schemas
	err = db.AutoMigrate(
		&models.User{},
		&models.Project{},
		&models.Collection{},
		&models.Document{},
	)
	if err != nil {
		log.Fatalf("Failed to migrate database schema: %v\n", err)
	}

	log.Println("Database schemas migrated successfully.")
	DB = db
}
