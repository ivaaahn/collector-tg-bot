package infra

import (
	"database/sql"
	"fmt"
)

type DatabaseConfig struct {
	Host     string
	Port     string
	Database string
	User     string
	Password string
}

func NewDb(config *DatabaseConfig) (*sql.DB, error) {
	connString := fmt.Sprintf("postgres://%s:%s@%s:%s/%s?sslmode=disable", config.User, config.Password,
		config.Host, config.Port, config.Database)
	return sql.Open("postgres", connString)
}
