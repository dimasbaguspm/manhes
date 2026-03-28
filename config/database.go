package config

import (
	"fmt"
)

type DatabaseConfig struct {
	Host     string
	Port     int
	User     string
	Password string
	Name     string
}

func loadDatabaseConfig() DatabaseConfig {
	return DatabaseConfig{
		Host:     envStr("DB_HOST", "localhost"),
		Port:     envInt("DB_PORT", 3306),
		User:     envStr("DB_USER", "manhes"),
		Password: envStr("DB_PASS", "manhes"),
		Name:     envStr("DB_NAME", "manhes"),
	}
}

func (c DatabaseConfig) DSN() string {
	return fmt.Sprintf("%s:%s@tcp(%s:%d)/%s?parseTime=true&charset=utf8mb4",
		c.User, c.Password, c.Host, c.Port, c.Name)
}

func (c DatabaseConfig) MaxOpenConns() int {
	return envInt("DB_MAX_OPEN_CONNS", 25)
}

func (c DatabaseConfig) MaxIdleConns() int {
	return envInt("DB_MAX_IDLE_CONNS", 5)
}
