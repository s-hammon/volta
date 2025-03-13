package api

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/rs/zerolog/log"
)

type logMsg struct {
	NotifSize string  `json:"notif_size,omitempty"`
	Hl7Size   string  `json:"hl7_size,omitempty"`
	Result    string  `json:"result,omitempty"`
	Elapsed   float64 `json:"elapsed,omitempty"`

	start time.Time `json:"-"`
}

func NewLogMsg() *logMsg {
	return &logMsg{
		start: time.Now(),
	}
}

func (l *logMsg) RespondJSON(w http.ResponseWriter, code int, result string) {
	l.Result = result
	if l.start.IsZero() {
		l.start = time.Now()
	}
	l.Elapsed = float64(time.Since(l.start).Milliseconds()) / 1000

	dat, err := json.Marshal(l)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		if _, err = w.Write([]byte(`{"error": "error marshalling log message"}`)); err != nil {
			log.Error().Err(err).Msg("error writing response")
		}
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	if _, err := w.Write(dat); err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
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

func middlewareLogging(next http.Handler) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		lw := &logWriter{w, http.StatusOK, nil}
		next.ServeHTTP(lw, r)

		if lw.StatusCode > 499 {
			log.Error().
				Str("method", r.Method).
				Str("path", r.URL.Path).
				Int("status", lw.StatusCode).
				Interface("response", lw.Message)
		} else {
			log.Info().
				Str("method", r.Method).
				Str("path", r.URL.Path).
				Int("status", lw.StatusCode).
				Interface("response", lw.Message)
		}
	}
}
