package config

import (
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"log"
)

var DB *gorm.DB

func InitDB() {
	var err error
	DB, err = gorm.Open(sqlite.Open("nas_app.db"), &gorm.Config{})
	if err != nil {
		log.Fatal("Failed to connect to database:", err)
	}

	// Auto Migrate the schema
	err = DB.AutoMigrate(
	//&auth.User{},
	//&auth.Album{},
	//&media.MediaItem{},
	)
	if err != nil {
		log.Fatal("Failed to migrate database:", err)
	}
}
