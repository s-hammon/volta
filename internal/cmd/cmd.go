package cmd

import (
	"context"
	"io"
	"net"
	"net/http"
	"os"
	"os/exec"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/rs/zerolog/log"
	"github.com/s-hammon/volta/internal/api"
	"github.com/s-hammon/volta/internal/entity"
	"github.com/spf13/cobra"

	"github.com/s-hammon/p"
)

// TODO: include YAML config file to parse & pass to service.
var (
	dbURL     string
	host      string
	port      string
	debugMode bool

	db *pgxpool.Pool

	ctx    context.Context
	cancel context.CancelFunc
)

func init() {
	ctx, cancel = context.WithCancel(context.Background())
	// TODO: init config

	serveCmd.PersistentFlags().StringVarP(&host, "host", "H", "localhost", "host for service (default: localhost)")
	serveCmd.PersistentFlags().StringVarP(&port, "port", "p", "8080", "port to listen on (default: 8080)")
	serveCmd.PersistentFlags().StringVarP(&dbURL, "db-url", "d", "", "database URL (required unless DATABASE_URL env var is set or using debug mode)")
	serveCmd.PersistentFlags().BoolVarP(&debugMode, "debug", "D", false, "enable debug mode; results are just logged to stdout, not written to the database (cannot use with -d)")
}

func Execute(args []string, stdin io.Reader, stdout, stderr io.Writer) int {
	rootCmd := &cobra.Command{
		Use:          "volta",
		SilenceUsage: true,
	}
	rootCmd.AddCommand(serveCmd)

	rootCmd.SetArgs(args)
	rootCmd.SetIn(stdin)
	rootCmd.SetOut(stdout)
	rootCmd.SetErr(stderr)

	if err := rootCmd.ExecuteContext(ctx); err != nil {
		defer cleanup()
		if exitError, ok := err.(*exec.ExitError); ok {
			return exitError.ExitCode()
		} else {
			return 1
		}
	}

	return 0
}

var serveCmd = &cobra.Command{
	Use:   "serve",
	Short: "Start the Volta service",
	RunE: func(cmd *cobra.Command, args []string) error {
		host, err := cmd.Flags().GetString("host")
		if err != nil {
			log.Info().Err(err).Msg("failed to get host flag")
			return err
		}

		if dbURL == "" {
			dbURL = os.Getenv("DATABASE_URL")
		}
		if (dbURL == "" && !debugMode) || (dbURL != "" && debugMode) {
			return cmd.Usage()
		}

		log.Info().Str("host", host).Str("port", port).Msg("service configuration")

		client, err := api.NewHl7Client(cmd.Context())
		if err != nil {
			log.Info().Err(err).Msg("failed to create HL7 client")
			return err
		}

		if !debugMode {
			db, err = pgxpool.New(ctx, dbURL)
			if err != nil {
				log.Info().Err(err).Msg("failed to connect to database")
				return err
			}
			if err := db.Ping(ctx); err != nil {
				log.Info().Err(err).Msg("couldn't reach database")
				return err
			}
			log.Info().Msg("connected to database")
		} else {
			log.Info().Msg("debug mode enabled; printing messages to stdout")
		}

		store := entity.NewRepo(db)
		srv := &http.Server{
			Addr:              net.JoinHostPort(host, port),
			Handler:           api.New(store, client, debugMode),
			ReadHeaderTimeout: 3 * time.Second,
		}

		log.Info().Msg(p.Format("starting server on %s", srv.Addr))
		return srv.ListenAndServe()
	},
}

func cleanup() {
	log.Info().Msg("shutting down services...")

	if db != nil {
		db.Close()
	}

	cancel()
}
