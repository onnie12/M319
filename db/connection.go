package db

import (
	"fmt"
	"os"

	"github.com/joho/godotenv"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

// Connect loads .env when present and opens a Postgres connection using
// DB_HOST, DB_PORT, DB_USER, DB_PASSWORD, DB_NAME, and DB_SSLMODE.
func Connect() (*gorm.DB, error) {
	_ = godotenv.Load()

	host := os.Getenv("DB_HOST")
	port := os.Getenv("DB_PORT")
	user := os.Getenv("DB_USER")
	password := os.Getenv("DB_PASSWORD")
	dbname := os.Getenv("DB_NAME")
	sslmode := os.Getenv("DB_SSLMODE")

	dsn := fmt.Sprintf(
		"host=%s port=%s user=%s password=%s dbname=%s sslmode=%s",
		host, port, user, password, dbname, sslmode,
	)

	return gorm.Open(postgres.Open(dsn), &gorm.Config{})
}

// Migrate creates or updates the equipment, skill, and hero tables.
func Migrate(gdb *gorm.DB) error {
	return gdb.AutoMigrate(&Equipment{}, &Skill{}, &Hero{})
}
