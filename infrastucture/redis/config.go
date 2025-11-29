package redis

import (
	"os"
	"time"
)

type RedisConfig struct {
	Addr        string
	User        string
	Password    string
	DB          int
	MaxRetries  int
	DialTimeout time.Duration
	Timeout     time.Duration
}

func LoadRedisConfig() RedisConfig {
	host := getEnv("REDIS_HOST", "localhost")
	addr := host + ":" + getEnv("REDIS_PORT", "6379")
	return RedisConfig{
		Addr:        addr,
		User:        getEnv("REDIS_USER", ""),
		Password:    getEnv("REDIS_PASSWORD", ""),
		DB:          0,
		MaxRetries:  3,
		DialTimeout: 5 * time.Second,
		Timeout:     3 * time.Second,
	}
}

func getEnv(key, def string) string {
	if val, ok := os.LookupEnv(key); ok {
		return val
	}
	return def
}
