package api

import (
	"time"

	"github.com/rs/zerolog/log"
)

type logMsg struct {
	notifSize string
	hl7Size   string
	result    string
	elapsed   time.Duration
}

func (l *logMsg) Log(result string) {
	l.result = result
	log.Info().
		Str("notifSize", l.notifSize).
		Str("hl7Size", l.hl7Size).
		Str("result", l.result).
		Dur("elapsed", l.elapsed).
		Msg("message processed")
}

func (l *logMsg) Error(err error, sendingFac, ControlID string) {
	l.result = err.Error()
	log.Error().
		Err(err).
		Str("sendingFacility", sendingFac).
		Str("controlID", ControlID).
		Msg("could not process message")
}
