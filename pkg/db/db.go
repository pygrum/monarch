package db

import (
	"errors"
	"fmt"
	"github.com/google/uuid"
	"github.com/pygrum/monarch/pkg/config"
	"github.com/pygrum/monarch/pkg/consts"
	"github.com/pygrum/monarch/pkg/log"
	"github.com/pygrum/monarch/pkg/teamserver/roles"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
	"strconv"
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
	return initServerUser(&conf)
}

// initialize profiles for default handlers - HTTP, HTTPS, TCP
func initDefaultProfiles(c *config.MonarchConfig) {
	httpProfile := &Profile{
		Name:      consts.ProfileHTTP,
		BuilderID: consts.TypeInternalProfile,
		CreatedBy: consts.ServerUser,
	}
	httpR1 := &ProfileRecord{
		Profile: consts.ProfileHTTP,
		Name:    consts.OpLPort,
		Value:   strconv.Itoa(c.HttpPort),
	}
	if tx := db.Create(httpProfile); tx.Error != nil {
		l.Warn(tx.Error.Error())
	}
	if tx := db.Create(httpR1); tx.Error != nil {
		l.Warn(tx.Error.Error())
	}
	httpsProfile := &Profile{
		Name:      consts.ProfileHTTPS,
		BuilderID: consts.TypeInternalProfile,
		CreatedBy: consts.ServerUser,
	}
	httpsR1 := &ProfileRecord{
		Profile: consts.ProfileHTTPS,
		Name:    consts.OpLPort,
		Value:   strconv.Itoa(c.HttpsPort),
	}
	if tx := db.Create(httpsProfile); tx.Error != nil {
		l.Warn(tx.Error.Error())
	}
	if tx := db.Create(httpsR1); tx.Error != nil {
		l.Warn(tx.Error.Error())
	}
	tcpProfile := &Profile{
		Name:      consts.ProfileTCP,
		BuilderID: consts.TypeInternalProfile,
		CreatedBy: consts.ServerUser,
	}
	tcpR1 := &ProfileRecord{
		Profile: consts.ProfileTCP,
		Name:    consts.OpLPort,
		Value:   strconv.Itoa(c.TcpPort),
	}
	if tx := db.Create(tcpProfile); tx.Error != nil {
		l.Warn(tx.Error.Error())
	}
	if tx := db.Create(tcpR1); tx.Error != nil {
		l.Warn(tx.Error.Error())
	}
}

func initServerUser(c *config.MonarchConfig) string {
	uid := uuid.New().String()
	serverPlayer := &Player{}
	if db.Where("username = ?", consts.ServerUser).First(serverPlayer); len(serverPlayer.UUID) == 0 {
		// we can do first-time-run steps here
		serverPlayer = &Player{
			UUID:     uid,
			Username: consts.ServerUser,
			Role:     roles.RoleAdmin,
		}
		if result := db.Create(serverPlayer); result.Error != nil {
			l.Fatal("could not create default server user: %v", result.Error)
		}
		config.ClientConfig.UUID = uid

		// create default profiles (for now it just loads port numbers for each endpoint type)
		initDefaultProfiles(c)
	} else {
		config.ClientConfig.Name = consts.ServerUser
		config.ClientConfig.UUID = serverPlayer.UUID
		return serverPlayer.UUID
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
