package hl7

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestHeaderParsing(t *testing.T) {
	data := "MSH|^~\\&|"
	s := newScanner('\r')
	s.step = stateInit

	var lastResult int
	for i := range len(data) {
		lastResult = s.step(s, data[i])
		require.NotEqual(t,
			scanError,
			lastResult,
			"unexpected error at index %d (%c): got \"%v\"\n",
			i+1, data[i], s.err,
		)
	}

	assert.Equal(t, byte('|'), s.fldDelim)
	assert.Equal(t, byte('^'), s.comDelim)
	assert.Equal(t, byte('~'), s.repDelim)
	assert.Equal(t, byte('\\'), s.escDelim)
	assert.Equal(t, byte('&'), s.subDelim)
	assert.Equal(t, scanEndHeader, lastResult)
}

func TestInvalidFirstChar(t *testing.T) {
	s := newScanner('\r')
	s.step = stateInit

	result := s.step(s, 'X')
	require.Equal(t, scanError, result)
	assert.NotNil(t, s.err)
	assert.Contains(t, s.err.Error(), "expected first character to be 'M'")
}

func TestEscapeSequence(t *testing.T) {
	s := newScanner('\r')
	s.fldDelim = '|'
	s.comDelim = '^'
	s.repDelim = '~'
	s.escDelim = '\\'
	s.subDelim = '&'

	s.step = stateBeginEscape

	result := s.step(s, 'F')
	require.Equal(t, scanContinue, result)

	result = s.step(s, '\\')
	assert.Equal(t, scanEndEscape, result)
}

func TestInvalidEscape(t *testing.T) {
	s := newScanner('\r')
	s.escDelim = '\\'

	s.step = stateBeginEscape

	result := s.step(s, '!')
	require.Equal(t, scanError, result)
	assert.NotNil(t, s.err)
	assert.Contains(t, s.err.Error(), "invalid escape literal")
}
