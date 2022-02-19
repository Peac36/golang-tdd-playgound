package database

import (
	"fmt"
	"os"
	"site/database/models"

	"gorm.io/driver/mysql"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func NewDatabaseConnection() (*gorm.DB, error) {
	var dbUser string = os.Getenv("DB_USER")
	var dbPassword string = os.Getenv("DB_PASSWORD")
	var dbHost string = os.Getenv("DB_HOST")
	var dbDatabase string = os.Getenv("DB_NAME")
	dsn := fmt.Sprintf("%s:%s@tcp(%s)/%s?charset=utf8mb4&parseTime=True&loc=Local", dbUser, dbPassword, dbHost, dbDatabase)
	return gorm.Open(mysql.Open(dsn), &gorm.Config{})
}

func NewTestDatabaseConnection() (*gorm.DB, error) {
	return gorm.Open(sqlite.Open("file::memory:?cache=shared"), &gorm.Config{})
}

func CloseDatabaseConnection(connection *gorm.DB) {
	dbInstance, err := connection.DB()
	if err != nil {
		fmt.Printf("Error closing the database %v \n", err)
	}
	dbInstance.Close()
	fmt.Println("The database connection has been closed")
}

func RunMigrations(connection *gorm.DB) error {
	connection.AutoMigrate(&models.Event{})
	connection.AutoMigrate(&models.User{})
	connection.AutoMigrate(&models.Media{})
	return nil
}
