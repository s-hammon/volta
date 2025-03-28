package testdb

import (
	"database/sql"
	"fmt"
	"strconv"

	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/ory/dockertest/v3"
	"github.com/ory/dockertest/v3/docker"
)

const (
	POSTGRES_IMAGE    = "postgres"
	POSTGRES_VERSION  = "16-alpine"
	POSTGRES_DB       = "testdb"
	POSTGRES_USER     = "postgres"
	POSTGRES_PASSWORD = "postgres"
)

type options struct {
	bindPort int
	debug    bool
}

type OptionsFunc func(o *options)

func WithBindPort(n int) OptionsFunc {
	return func(o *options) { o.bindPort = n }
}

func WithDebug(b bool) OptionsFunc {
	return func(o *options) { o.debug = b }
}

func NewPostgres(options ...OptionsFunc) (db *sql.DB, cleanup func(), err error) {
	return newPostgres(options...)
}

func newPostgres(opts ...OptionsFunc) (*sql.DB, func(), error) {
	option := &options{}
	for _, f := range opts {
		f(option)
	}

	pool, err := dockertest.NewPool("")
	if err != nil {
		return nil, nil, fmt.Errorf("failed to connect to docker: %v", err)
	}
	options := &dockertest.RunOptions{
		Repository: POSTGRES_IMAGE,
		Tag:        POSTGRES_VERSION,
		Env: []string{
			"POSTGRES_USER=" + POSTGRES_USER,
			"POSTGRES_PASSWORD=" + POSTGRES_PASSWORD,
			"POSTGRES_DB=" + POSTGRES_DB,
			"listen_addresses = '*'",
		},
		Labels:       map[string]string{"volta_test": "1"},
		PortBindings: make(map[docker.Port][]docker.PortBinding),
	}
	if option.bindPort > 0 {
		options.PortBindings[docker.Port("5432/tcp")] = []docker.PortBinding{{HostPort: strconv.Itoa(option.bindPort)}}
	}
	container, err := pool.RunWithOptions(
		options,
		func(config *docker.HostConfig) {
			config.AutoRemove = true
			config.RestartPolicy = docker.RestartPolicy{Name: "no"}
		},
	)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to create container: %v", err)
	}
	cleanup := func() {
		if option.debug {
			return
		}
		if err := pool.Purge(container); err != nil {
			fmt.Printf("failed to purge container: %v", err)
		}
	}
	psqlInfo := fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=disable",
		"localhost",
		container.GetPort("5432/tcp"),
		POSTGRES_USER,
		POSTGRES_PASSWORD,
		POSTGRES_DB,
	)

	var db *sql.DB
	if err := pool.Retry(
		func() error {
			var err error
			db, err = sql.Open("pgx", psqlInfo)
			if err != nil {
				return err
			}
			return db.Ping()
		},
	); err != nil {
		return nil, cleanup, fmt.Errorf("could not connect to docker: %v", err)
	}
	return db, cleanup, nil
}
