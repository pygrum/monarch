package db

import (
	"errors"
	"fmt"
	"github.com/google/uuid"
	"github.com/pygrum/monarch/pkg/config"
	"github.com/pygrum/monarch/pkg/log"
	"github.com/pygrum/monarch/pkg/teamserver/roles"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

var (
	ErrPlayerNotFound = errors.New("player not found")
	db                *gorm.DB
	l                 log.Logger
)

// Initialize database
func Initialize() string {
	l, _ = log.NewLogger(log.ConsoleLogger, "")
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
	if err = db.AutoMigrate(&Builder{}, &Agent{}, &Player{}, &Profile{}, &ProfileRecord{}); err != nil {
		l.Fatal("failed to migrate schema: %v. Monarch cannot continue to operate", err)
	}
	uid := uuid.New().String()
	consoleUser := &Player{}
	if db.Where("username = ?", "console").First(consoleUser); len(consoleUser.UUID) == 0 {
		consoleUser = &Player{
			UUID:     uid,
			Username: "console",
			Role:     roles.RoleAdmin,
		}
		if result := db.Create(consoleUser); result.Error != nil {
			l.Fatal("could not create default 'console' user: %v", result.Error)
		}
		config.ClientConfig.UUID = uid
	} else {
		config.ClientConfig.UUID = consoleUser.UUID
		return consoleUser.UUID
	}
	return uid
}

func Create(v interface{}) error {
	result := db.Create(v)
	return result.Error
}

func Find(v interface{}) error {
	result := db.Find(v)
	return result.Error
}

// FindConditional retrieves rows based on one specific condition
func FindConditional(query, target, v interface{}) error {
	result := db.Where(query, target).Find(v)
	return result.Error
}

// FindOneConditional works like FindConditional but returns the first instance
func FindOneConditional(query, target, v interface{}) error {
	result := db.Where(query, target).First(v)
	return result.Error
}

func Delete(v interface{}) error {
	result := db.Delete(v)
	return result.Error
}

func DeleteOne(v interface{}) error {
	result := db.Delete(v)
	return result.Error
}

func Where(query interface{}, target ...interface{}) *gorm.DB {
	return db.Where(query, target...)
}

func GetIDByUsername(user string) (string, error) {
	p := &Player{}
	if err := FindOneConditional("username = ?", user, &p); err != nil {
		return "", err
	}
	if len(p.UUID) == 0 {
		return p.UUID, ErrPlayerNotFound
	}
	return p.UUID, nil
}
