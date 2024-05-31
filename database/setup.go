package database

import (
	"AttestationVerifier/models"
	"log"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

var DB *gorm.DB

func Connect() {
	dsn := "host=localhost user=postgres password=123 dbname=postgres port=5432 sslmode=disable"
	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		log.Fatal("Failed to connect to database:", err)
	}

	DB = db
}

func Migrate() {
	if err := DB.AutoMigrate(&models.AttestationReport{}); err != nil {
		log.Fatal("Failed to migrate database:", err)
	}
}
