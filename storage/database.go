package storage

import (
	"jotno-server/models"
	"log"
	"os"

	"github.com/joho/godotenv"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

var DB *gorm.DB

func connection() *gorm.DB {
	err := godotenv.Load()
	if err != nil {
		log.Panic("Error loading env file")
	}
	dsn := os.Getenv("DB_CONNECTION_STRING")
	db, dbError := gorm.Open(postgres.Open(dsn), &gorm.Config{})

	if dbError != nil {
		log.Panic("error connecting to db")
	}
	DB = db
	return db
}

func performMigrations(db *gorm.DB) {
	db.AutoMigrate(
		&models.User{},
		&models.Specialist{},
		&models.Job{},
		&models.Post{},
		&models.Review{},
		&models.JobPost{},
		&models.Comment{},
		&models.Chat{},
		&models.Message{},
	)
}

func InitializeDB() *gorm.DB {
	db := connection()
	performMigrations(db)
	return db
}
