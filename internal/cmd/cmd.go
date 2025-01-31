package cmd

import (
	"context"
	"io"
	"net"
	"net/http"
	"os/exec"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"github.com/s-hammon/volta/internal/api"
	"github.com/s-hammon/volta/internal/database"
	"github.com/spf13/cobra"
)

var (
	dbURL string
	host  string
	port  string

	db *pgxpool.Pool

	ctx    context.Context
	cancel context.CancelFunc
)

func init() {
	ctx, cancel = context.WithCancel(context.Background())

	serveCmd.PersistentFlags().StringVarP(&host, "host", "H", "localhost", "host for service (default: localhost)")
	serveCmd.PersistentFlags().StringVarP(&port, "port", "p", "8080", "port to listen on (default: 8080)")
	serveCmd.PersistentFlags().StringVarP(&dbURL, "db-url", "d", "", "database URL (required)")
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
			return err
		}

		port, err := cmd.Flags().GetString("port")
		if err != nil {
			return err
		}

		dbURL, err := cmd.Flags().GetString("db-url")
		if err != nil {
			return err
		}
		if dbURL == "" {
			return cmd.Usage()
		}

		log.Info().Str("host", host).Str("port", port).Msg("service configuration")

		client, err := api.NewHl7Client(cmd.Context())
		if err != nil {
			return err
		}

		db, err = pgxpool.New(ctx, dbURL)
		if err != nil {
			return err
		}

		srv := &http.Server{
			Addr:    net.JoinHostPort(host, port),
			Handler: api.New(database.New(db), client),
		}

		return srv.ListenAndServe()
	},
}

func initLogger(cmd *cobra.Command, args []string) {
	zerolog.TimeFieldFormat = zerolog.TimeFormatUnix
	zerolog.DurationFieldInteger = true

	log.Logger = zerolog.New(cmd.OutOrStdout())
}

func cleanup() {
	log.Info().Msg("shutting down services...")

	if db != nil {
		db.Close()
	}

	cancel()
}
