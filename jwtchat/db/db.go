package db

import (
	"libruary/logs"
	"log"

	"github.com/sirupsen/logrus"

	"github.com/jinzhu/gorm"
	_ "github.com/jinzhu/gorm/dialects/mysql"
)

var (
	dbr *gorm.DB
)

// Open function recieve an expr as argument
// and making connection to our database
func Open(expr string, logger *logrus.Logger) {
	db, err := gorm.Open("mysql", expr)
	if err != nil {
		log.Panic("Couldn't opendatabase", err.Error())
	}
	db.LogMode(true)
	db.SetLogger(&logs.GormLogger{
		Name:   "db gorm logger",
		Logger: logger,
	})
	dbr = db
}

func GetDB() *gorm.DB {
	return dbr
}
