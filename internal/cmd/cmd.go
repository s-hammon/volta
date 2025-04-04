package cmd

import (
	"context"
	"fmt"
	"io"
	"net"
	"net/http"
	"os"
	"os/exec"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"github.com/s-hammon/volta/internal/api"
	"github.com/spf13/cobra"
)

// TODO: include YAML config file to parse & pass to service.
var (
	dbURL          string
	host           string
	port           string
	debugMode      bool
	sqlProxyDriver string

	db *pgxpool.Pool

	ctx    context.Context
	cancel context.CancelFunc
)

func init() {
	ctx, cancel = context.WithCancel(context.Background())
	// TODO: init config

	serveCmd.PersistentFlags().StringVarP(&host, "host", "H", "localhost", "host for service (default: localhost)")
	serveCmd.PersistentFlags().StringVarP(&port, "port", "p", "8080", "port to listen on (default: 8080)")
	serveCmd.PersistentFlags().StringVarP(&dbURL, "db-url", "d", "", "database URL (required unless using debug mode)")
	serveCmd.PersistentFlags().StringVar(&sqlProxyDriver, "sql-proxy-driver", "", "if using Cloud SQL proxy, specify the driver (e.g., 'postgres'). DB_USER, DB_PASSWORD, and DB_NAME must be set. Will overwrite db-url if specified.")
	serveCmd.PersistentFlags().BoolVarP(&debugMode, "debug", "D", false, "enable debug mode; results are just logged to stdout, not written to the database (cannot use with -d)")
}

func Execute(args []string, stdin io.Reader, stdout, stderr io.Writer) int {
	rootCmd := &cobra.Command{
		Use:              "volta",
		SilenceUsage:     true,
		PersistentPreRun: initLogger,
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

		if sqlProxyDriver != "" {
			switch sqlProxyDriver {
			case "postgres":
				proxyConfig, err := getPostgresConfig()
				if err != nil {
					log.Info().Err(err).Msg("failed to load DB secrets from environment")
					return err
				}
				dbURL = proxyConfig.String()
			default:
				return fmt.Errorf("unsupported SQL proxy driver: %s", sqlProxyDriver)
			}
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

		var db api.DB
		if !debugMode {
			pool, err := pgxpool.New(ctx, dbURL)
			if err != nil {
				log.Info().Err(err).Msg("failed to connect to database")
				return err
			}
			log.Info().Msg("connected to database")
			db = api.NewDB(pool)
		} else {
			log.Info().Msg("debug mode enabled; printing messages to stdout")
		}

		srv := &http.Server{
			Addr:              net.JoinHostPort(host, port),
			Handler:           api.New(db, client, debugMode),
			ReadHeaderTimeout: 3 * time.Second,
		}

		log.Info().Msg(fmt.Sprintf("starting server on %s", srv.Addr))
		return srv.ListenAndServe()
	},
}

func initLogger(cmd *cobra.Command, args []string) {
	zerolog.TimeFieldFormat = zerolog.TimeFormatUnix
	zerolog.DurationFieldInteger = true

	log.Logger = zerolog.New(os.Stdout).With().Timestamp().Logger()
}

func cleanup() {
	log.Info().Msg("shutting down services...")

	if db != nil {
		db.Close()
	}

	cancel()
}
