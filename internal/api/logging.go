package api

import (
	"net/http"

	"github.com/rs/zerolog/log"
)

type logMsg struct {
	NotifSize string  `json:"notif_size"`
	Hl7Size   string  `json:"hl7_size"`
	Result    string  `json:"result"`
	Elapsed   float64 `json:"elapsed"`
}

type logWriter struct {
	http.ResponseWriter
	StatusCode int
	Message    []byte
}

func (l *logWriter) WriteHeader(code int) {
	l.StatusCode = code
	l.ResponseWriter.WriteHeader(code)
}

func (l *logWriter) Write(b []byte) (int, error) {
	l.Message = b
	return l.ResponseWriter.Write(b)
}

func middlwareLogging(next http.Handler) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		lw := &logWriter{w, http.StatusOK, nil}
		next.ServeHTTP(lw, r)

		msg := string(lw.Message)
		if lw.StatusCode > 499 {
			log.Error().
				Str("method", r.Method).
				Str("path", r.URL.Path).
				Int("status", lw.StatusCode).
				Msg(msg)
		} else {
			log.Info().
				Str("method", r.Method).
				Str("path", r.URL.Path).
				Int("status", lw.StatusCode).
				Msg(msg)
		}
	}
}
