package database

import (
	"encoding/json"
	"os"
	"path"
	"strings"
	"time"
	"log"

	"github.com/alireza0/s-ui/config"
	"github.com/alireza0/s-ui/database/model"

	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

var db *gorm.DB

func initUser() error {
	var count int64
	db.Model(&model.User{}).Count(&count)

	forceReset := os.Getenv("SUI_FORCE_RESET") == "true"
	tokenStr := os.Getenv("TOKEN")
	var user model.User

	if count == 0 || forceReset {
		username := os.Getenv("SUI_USERNAME")
		password := os.Getenv("SUI_PASSWORD")
		if username == "" { username = "admin" }
		if password == "" { password = "admin" }
		
		user = model.User{Username: username, Password: password}
		if err := db.Create(&user).Error; err != nil {
			return err
		}
		log.Printf("[DB] Created new user: %s (ID: %d)", user.Username, user.Id)
	} else {
		if err := db.First(&user).Error; err == nil {
			log.Printf("[DB] Using existing user: %s (ID: %d)", user.Username, user.Id)
		}
	}

	if tokenStr != "" && user.Id > 0 {
		var tokenCount int64
		db.Model(&model.Tokens{}).Where("token = ?", tokenStr).Count(&tokenCount)

		if tokenCount == 0 {
			expiry := time.Now().Add(24 * time.Hour).Unix()
			newToken := &model.Tokens{
				Token:  tokenStr,
				Desc:   "start_tocken",
				UserId: user.Id,
				Expiry: expiry,
			}

			if err := db.Create(newToken).Error; err != nil {
				log.Printf("[DB] Error creating token: %v", err)
			} else {
				log.Printf("[DB] Token '%s' successfully linked to User ID: %d", tokenStr, user.Id)
			}
		} else {
			log.Printf("[DB] Token '%s' already exists in database", tokenStr)
		}
	}

	return nil
}



func OpenDB(dbPath string) error {
	dir := path.Dir(dbPath)
	err := os.MkdirAll(dir, 01740)
	if err != nil {
		return err
	}

	var gormLogger logger.Interface

	if config.IsDebug() {
		gormLogger = logger.Default
	} else {
		gormLogger = logger.Discard
	}

	c := &gorm.Config{
		Logger: gormLogger,
	}
	sep := "?"
	if strings.Contains(dbPath, "?") {
		sep = "&"
	}
	dsn := dbPath + sep + "_busy_timeout=10000&_journal_mode=WAL"
	db, err = gorm.Open(sqlite.Open(dsn), c)
	if err != nil {
		return err
	}

	sqlDB, err := db.DB()
	if err != nil {
		return err
	}
	sqlDB.SetMaxOpenConns(25)
	sqlDB.SetMaxIdleConns(5)
	sqlDB.SetConnMaxLifetime(time.Hour)

	if config.IsDebug() {
		db = db.Debug()
	}
	return nil
}

func InitDB(dbPath string) error {
	err := OpenDB(dbPath)
	if err != nil {
		return err
	}

	// Default Outbounds
	if !db.Migrator().HasTable(&model.Outbound{}) {
		db.Migrator().CreateTable(&model.Outbound{})
		defaultOutbound := []model.Outbound{
			{Type: "direct", Tag: "direct", Options: json.RawMessage(`{}`)},
		}
		db.Create(&defaultOutbound)
	}

		err = db.AutoMigrate(
		&model.Setting{},
		&model.Tls{},
		&model.Inbound{},
		&model.Outbound{},
		&model.Service{},
		&model.Endpoint{},
		&model.User{},
		&model.Tokens{},
		&model.Stats{},
		&model.Client{},
		&model.Changes{},
		&model.GitSync{},
	)
	if err != nil {
		return err
	}
	err = initUser()
	if err != nil {
		return err
	}

	return nil
}

func GetDB() *gorm.DB {
	return db
}

func IsNotFound(err error) bool {
	return err == gorm.ErrRecordNotFound
}
