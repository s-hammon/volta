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
		log.Error().Err(err).Msg("couldn't marshal log message")
		http.Error(w, `{"error":"internal server error"}`, http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)

	if _, err := w.Write(dat); err != nil {
		log.Error().Err(err).Msg("couldn't write JSON response")
	}

	log.Info().
		Int("status", code).
		Str("result", l.Result).
		Float64("elapsed", l.Elapsed).
		Msg("response sent")
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

		defer func() {
			if rec := recover(); rec != nil {
				log.Error().
					Str("method", r.Method).
					Str("path", r.URL.Path).
					Interface("recover", rec).
					Msg("panic recovered")
				http.Error(lw, "internal server error", http.StatusInternalServerError)
			}

			if lw.StatusCode > 499 {
				log.Error().
					Str("method", r.Method).
					Str("path", r.URL.Path).
					Int("status", lw.StatusCode).
					RawJSON("response", lw.Message).
					Msg("error response")
			} else {
				log.Info().
					Str("method", r.Method).
					Str("path", r.URL.Path).
					Int("status", lw.StatusCode).
					RawJSON("response", lw.Message).
					Msg("successful response")
			}
		}()

		next.ServeHTTP(lw, r)
	}
}
