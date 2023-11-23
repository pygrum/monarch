package db

import (
	"fmt"
	"github.com/pygrum/monarch/pkg/config"
	"github.com/pygrum/monarch/pkg/log"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

var db *gorm.DB
var l log.Logger

// Initialize database
func Initialize() {
	l, _ = log.NewLogger(log.ConsoleLogger, "")
	config.Init()
	conf := config.MainConfig
	dsn := fmt.Sprintf("%s:%s@tcp(%s:3306)/monarch?charset=utf8mb4&parseTime=True&loc=Local",
		conf.MysqlUsername, conf.MysqlPassword, conf.MysqlAddress)
	var err error
	db, err = gorm.Open(mysql.Open(dsn), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Silent),
	})

	if err != nil {
		l.Fatal("failed to connect to database: %v. Monarch cannot continue to operate", err)
	}
	if err = db.AutoMigrate(&Builder{}, &Agent{}); err != nil {
		l.Fatal("failed to migrate schema: %v. Monarch cannot continue to operate", err)
	}
}

func Create(v interface{}) error {
	result := db.Create(v)
	return result.Error
}

func Find(v interface{}) error {
	result := db.Find(v)
	return result.Error
}

// FindConditional retrieves rows based on a specific condition
func FindConditional(query, target, v interface{}) error {
	result := db.Where(query, target).Find(v)
	return result.Error
}

// FindOneConditional works like FindConditional but returns the first instance
func FindOneConditional(query, target, v interface{}) error {
	result := db.Where(query, target).First(v)
	return result.Error
}
