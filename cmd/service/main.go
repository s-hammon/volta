package main

import (
	"context"
	"net"
	"net/http"
	"os"
	"os/signal"
	"sync"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"github.com/s-hammon/volta/internal/api"
	"github.com/s-hammon/volta/internal/database"
)

var (
	dbURL string
	host  string
	port  string

	db *pgxpool.Pool
	a  *http.ServeMux

	wg     sync.WaitGroup
	ctx    context.Context
	cancel context.CancelFunc
)

func init() {
	zerolog.TimeFieldFormat = zerolog.TimeFormatUnix
	zerolog.DurationFieldInteger = true

	// TODO: put next 3 as CLI args
	dbURL = os.Getenv("DATABASE_URL")
	if dbURL == "" {
		log.Fatal().Msg("DATABASE_URL is required")
	}

	host = os.Getenv("SERVICE_HOST")
	if host == "" {
		host = "localhost"
	}
	port = os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}
	log.Info().Str("host", host).Str("port", port).Msg("service configuration")

	client, err := api.NewHl7Client(ctx)
	if err != nil {
		log.Fatal().Err(err).Msg("could not create HL7 client")
	}

	ctx, cancel = context.WithCancel(context.Background())

	db, err = pgxpool.New(ctx, dbURL)
	if err != nil {
		log.Fatal().Err(err).Str("dbURL", dbURL).Msg("could not connect to DB")
	}

	a = api.New(database.New(db), client)
}

func main() {
	defer cleanup()
	go handleShutdown()

	srv := &http.Server{
		Addr:    net.JoinHostPort(host, port),
		Handler: a,
	}

	wg.Add(1)
	log.Info().Str("host", host).Str("port", port).Msgf("service listening on port %s", port)
	log.Fatal().Err(srv.ListenAndServe()).Msg("server error")
	wg.Wait()
}

func cleanup() {
	log.Info().Msg("shutting down services...")

	if db != nil {
		db.Close()
	}

	cancel()
}

func handleShutdown() {
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt)

	<-sigChan
	log.Info().Msg("received shutdown signal, shutting down...")

	cleanup()
	os.Exit(0)
}
