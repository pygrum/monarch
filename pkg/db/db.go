package db

import (
	"fmt"
	"github.com/pygrum/monarch/pkg/config"
	"github.com/pygrum/monarch/pkg/log"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

var db *gorm.DB
var l log.Logger

// Initialize database
func Initialize() {
	l, _ = log.NewLogger(log.ConsoleLogger, "")
	conf := config.MonarchConfig{}
	err := config.YamlConfig(config.MonarchConfigFile, &conf)
	if err != nil {
		l.Fatal("failed to retrieve configuration for database: %v. Monarch cannot continue to operate", err)
	}
	// mysql operates on localhost
	dsn := fmt.Sprintf("%s:%s@tcp(%s:3306)/monarch?charset=utf8mb4&parseTime=True&loc=Local",
		conf.MysqlUsername, conf.MysqlPassword, conf.MysqlAddress)

	db, err = gorm.Open(mysql.Open(dsn), &gorm.Config{})
	if err != nil {
		l.Fatal("failed to connect to database: %v. Monarch cannot continue to operate", err)
	}
	if err = db.AutoMigrate(&Agent{}, &Translator{}); err != nil {
		l.Fatal("failed to migrate schema: %v. Monarch cannot continue to operate", err)
	}
}

func Create(v interface{}) error {
	result := db.Create(v)
	return result.Error
}
