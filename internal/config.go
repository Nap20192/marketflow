package internal

import (
	"os"

	"marketflow/infrastucture/postgres"
	"marketflow/infrastucture/redis"
)

type config struct {
	postgres  postgres.PostgresConfig
	redis     redis.RedisConfig
	exchange1 string
	exchange2 string
	exchange3 string
}

func LoadConfig() (*config, error) {
	postgresConfig := postgres.LoadPostgresConfig()
	redisConfig := redis.LoadRedisConfig()
	exchanger1Host := getEnv("EXCHANGE1_HOST", "localhost")
	exchanger2Host := getEnv("EXCHANGE2_HOST", "localhost")
	exchanger3Host := getEnv("EXCHANGE3_HOST", "localhost")
	return &config{
		postgres:  postgresConfig,
		redis:     redisConfig,
		exchange1: exchanger1Host,
		exchange2: exchanger2Host,
		exchange3: exchanger3Host,
	}, nil
}

func getEnv(key, def string) string {
	if val, ok := os.LookupEnv(key); ok {
		return val
	}
	return def
}
