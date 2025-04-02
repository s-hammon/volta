package cmd

import (
	"fmt"
	"os"
)

type proxyConfig struct {
	host     string
	port     string
	user     string
	password string
	dbName   string
}

func getPostgresConfig() (proxyConfig, error) {
	// get DB_USER, DB_PASSWORD, and DB_NAME from environment variables
	user := os.Getenv("DB_USER")
	password := os.Getenv("DB_PASSWORD")
	dbName := os.Getenv("DB_NAME")

	if user == "" || password == "" || dbName == "" {
		return proxyConfig{}, fmt.Errorf("missing required environment variables")
	}

	cfg := proxyConfig{
		host:     "127.0.0.1",
		port:     "5432",
		user:     user,
		password: password,
		dbName:   dbName,
	}
	return cfg, nil
}

func (p proxyConfig) String() string {
	return fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s", p.host, p.port, p.user, p.password, p.dbName)
}
