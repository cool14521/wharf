package models

import (
	"os"

	log "github.com/Sirupsen/logrus"
	"github.com/jinzhu/gorm"

	"github.com/containerops/configure"
)

var (
	DB *gorm.DB
)

// init()
func init() {

}

// OpenDatabase is
func OpenDatabase() {
	var err error
	if DB, err = gorm.Open(configure.GetString("database.driver"), configure.GetString("database.uri")); err != nil {
		log.Fatal("Initlization database connection error.")
		os.Exit(1)
	} else {
		DB.DB()
		DB.DB().Ping()
		DB.DB().SetMaxIdleConns(10)
		DB.DB().SetMaxOpenConns(100)
		DB.SingularTable(true)
	}
}

// Migrate is
func Migrate() {
	OpenDatabase()

	log.Info("Auto Migrate Wharf Database Structs Done.")
}
