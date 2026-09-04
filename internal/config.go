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
	exchange1Host := getEnv("EXCHANGE1_HOST", "localhost")
	exchange2Host := getEnv("EXCHANGE2_HOST", "localhost")
	exchange3Host := getEnv("EXCHANGE3_HOST", "localhost")
	return &config{
		postgres:  postgresConfig,
		redis:     redisConfig,
		exchange1: exchange1Host,
		exchange2: exchange2Host,
		exchange3: exchange3Host,
	}, nil
}

func getEnv(key, def string) string {
	if val, ok := os.LookupEnv(key); ok {
		return val
	}
	return def
}
