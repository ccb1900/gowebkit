package redisx

import (
	"log"
	"strings"
	"sync"

	"github.com/ccb1900/gocommon/config"
	redis "github.com/redis/go-redis/v9"
)

type RedisClientConfig struct {
	Host     string `json:"host"`
	Db       int    `json:"db"`
	MinIdle  int    `json:"min_idle"`
	Password string `json:"password"`
	PoolSize int    `json:"pool_size"`
}

type RedisConfig struct {
	Clients map[string]RedisClientConfig `json:"clients"`
	Default string                       `json:"default"`
}

var (
	once      sync.Once
	clientMap map[string]redis.UniversalClient
)

func Default() redis.UniversalClient {
	return clientMap[config.Default().GetString("redis.default")]
}

func Init() {
	once.Do(func() {
		config := config.Default()
		var redisConfigDic RedisConfig
		if err := config.UnmarshalKey("redis", &redisConfigDic); err != nil {
			log.Println("load redis config", "err", err)
		} else {
			for key, v := range redisConfigDic.Clients {
				addrs := strings.Split(v.Host, ",")
				if len(addrs) == 0 {
					log.Println("load redis config", "key", key)
				}
				if v.MinIdle == 0 {
					v.MinIdle = 10
				}
				if v.PoolSize == 0 {
					v.PoolSize = 10
				}
				clientMap[key] = redis.NewUniversalClient(&redis.UniversalOptions{
					Addrs:        addrs,
					DB:           v.Db,
					Password:     v.Password,
					MinIdleConns: v.MinIdle,
					PoolSize:     v.PoolSize,
				})
			}
		}
	})
}
