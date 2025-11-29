package postgres

import (
	"fmt"
	"os"
)

type PostgresConfig struct {
	Host     string
	Port     int
	User     string
	Password string
	DBName   string
}

func LoadPostgresConfig() PostgresConfig {
	port := 5432

	return PostgresConfig{
		Host:     getEnv("POSTGRES_HOST", "localhost"),
		Port:     port,
		User:     getEnv("POSTGRES_USER", "marketflow"),
		Password: getEnv("POSTGRES_PASSWORD", "marketflow"),
		DBName:   getEnv("POSTGRES_DB", "marketflow"),
	}
}

func (c PostgresConfig) ConnString() string {
	return fmt.Sprintf("postgres://%s:%s@%s:%d/%s",
		c.User, c.Password, c.Host, c.Port, c.DBName)
}

func getEnv(key, def string) string {
	if val, ok := os.LookupEnv(key); ok {
		return val
	}
	return def
}
