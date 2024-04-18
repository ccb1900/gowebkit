package db

import (
	"sync"
	"time"

	"github.com/ccb1900/gocommon/config"
	"github.com/ccb1900/gocommon/logger"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

type DbClientConfig struct {
	Dsn         string `json:"dsn"`
	Type        string `json:"type"`
	MaxIdle     int    `json:"max_idle"`
	MaxOpen     int    `json:"max_open"`
	MaxLifetime int    `json:"max_lifetime"`
}

type DbConfig struct {
	Clients map[string]DbClientConfig `json:"clients"`
	Default string                    `json:"default"`
}

var once sync.Once

var clientMap map[string]*gorm.DB

func Default() *gorm.DB {
	return clientMap[config.Default().GetString("db.default")]
}

func Init() {
	once.Do(func() {
		var dbConfigDic DbConfig
		c := config.Default()

		if err := c.UnmarshalKey("db", &dbConfigDic); err != nil {
			logger.Default().Error("get db config ", "err", err)
		} else {
			var err error
			var sqlDB *gorm.DB
			clientMap = make(map[string]*gorm.DB)
			for name, client := range dbConfigDic.Clients {
				switch client.Type {
				case "mysql":
					sqlDB, err = gorm.Open(mysql.Open(client.Dsn), &gorm.Config{})
				case "postgres":
				case "sqlite":
				case "sqlserver":
				default:
					logger.Default().Error("db type not support", "name", name, "type", client.Type)
					return
				}
				if err != nil {
					logger.Default().Error("init db err", "err", err)
				} else {
					db, err := sqlDB.DB()
					if err != nil {
						logger.Default().Error("init get db err", "err", err)
					} else {
						if client.MaxIdle == 0 {
							client.MaxIdle = 10
						}

						if client.MaxLifetime == 0 {
							client.MaxLifetime = 1
						}

						if client.MaxOpen == 0 {
							client.MaxOpen = 100
						}
						db.SetMaxIdleConns(client.MaxIdle)
						db.SetMaxOpenConns(client.MaxOpen)
						db.SetConnMaxLifetime(time.Duration(client.MaxLifetime) * time.Hour)
						if config.Default().GetBool("debug") {
							sqlDB = sqlDB.Debug()
						}
						clientMap[name] = sqlDB
					}
				}
			}
		}
	})
}
