package hl7

import (
	"strconv"
	"sync"
)

// enums representing the current state
const (
	scanContinue = iota

	scanBeginHeader
	scanEndHeader

	scanBeginEscape
	scanEndEscape
	scanEndField
	scanEndComponent
	scanEndSubComponent
	scanEndRepeat
	scanEndSegment
	scanBeginLiteral
	scanBeginSegmentName
	scanEndSegmentName

	scanEnd
	scanError
)

// SyntaxError describes the HL7 syntax error
type SyntaxError struct {
	msg    string
	Offset int64
}

func (e *SyntaxError) Error() string { return e.msg }

// a HL7 scanning state machine
// takes inspiration from the `encoding/json` library
// fortunately, HL7 is easier to parse in some ways than JSON
// unfortunately, it is harder in others
type scanner struct {
	step  func(*scanner, byte) int
	err   error
	bytes int64

	charDict                                                   utfChars
	segDelim, fldDelim, comDelim, repDelim, escDelim, subDelim byte
}

var scannerPool = sync.Pool{
	New: func() any {
		return &scanner{}
	},
}

func newScanner(segDelim byte) *scanner {
	scan := scannerPool.Get().(*scanner)
	scan.charDict = newCharDict()
	scan.err = nil
	if segDelim == 0 {
		segDelim = '\r'
	}
	scan.segDelim = segDelim
	return scan
}

func freeScanner(scan *scanner) {
	scannerPool.Put(scan)
}

func stateInit(s *scanner, c byte) int {
	if c != 'M' {
		return s.error(c, "expected first character to be 'M'")
	}
	s.step = stateFirstHeaderSegNameChar
	return scanBeginHeader
}

func stateFirstHeaderSegNameChar(s *scanner, c byte) int {
	if c != 'S' {
		return s.error(c, "expected second character to be 'S'")
	}
	s.step = stateSecondHeaderSegNameChar
	return scanBeginHeader
}

func stateSecondHeaderSegNameChar(s *scanner, c byte) int {
	if c != 'H' {
		return s.error(c, "expected second character to be 'H'")
	}
	s.step = stateThirdHeaderSegNameChar
	return scanBeginHeader
}

func stateThirdHeaderSegNameChar(s *scanner, c byte) int {
	if c == s.segDelim {
		return s.error(c, "delimiter already in use")
	}
	if _, ok := s.charDict[c]; !ok {
		return s.error(c, "invalid delimiter character")
	}
	s.fldDelim = c
	delete(s.charDict, c)
	s.step = stateFieldDelim
	return scanBeginHeader
}

func stateFieldDelim(s *scanner, c byte) int {
	switch c {
	case s.segDelim, s.fldDelim:
		return s.error(c, "delimiter already in use")
	}
	if _, ok := s.charDict[c]; !ok {
		return s.error(c, "invalid delimiter character")
	}
	s.comDelim = c
	delete(s.charDict, c)
	s.step = stateComDelim
	return scanBeginHeader
}

func stateComDelim(s *scanner, c byte) int {
	switch c {
	case s.segDelim, s.fldDelim, s.comDelim:
		return s.error(c, "delimiter already in use")
	}
	if _, ok := s.charDict[c]; !ok {
		return s.error(c, "invalid delimiter character")
	}
	s.repDelim = c
	delete(s.charDict, c)
	s.step = stateRepDelim
	return scanBeginHeader
}

func stateRepDelim(s *scanner, c byte) int {
	switch c {
	case s.segDelim, s.fldDelim, s.comDelim, s.repDelim:
		return s.error(c, "delimiter already in use")
	}
	if _, ok := s.charDict[c]; !ok {
		return s.error(c, "invalid delimiter character")
	}
	s.escDelim = c
	delete(s.charDict, c)
	s.step = stateEscDelim
	return scanBeginHeader
}

func stateEscDelim(s *scanner, c byte) int {
	switch c {
	case s.segDelim, s.fldDelim, s.comDelim, s.repDelim, s.escDelim:
		return s.error(c, "delimiter already in use")
	}
	if _, ok := s.charDict[c]; !ok {
		return s.error(c, "invalid delimiter character")
	}
	s.subDelim = c
	delete(s.charDict, c)
	s.step = stateSubDelim
	return scanBeginHeader
}

func stateSubDelim(s *scanner, c byte) int {
	if c != s.fldDelim {
		return s.error(c, "expected field delimiter after encoding characters (MSH.2)")
	}
	s.step = stateEndLiteral
	return scanEndHeader
}

func stateEndLiteral(s *scanner, c byte) int {
	switch c {
	case s.segDelim:
		s.step = stateEndSegment
		return scanEndSegment
	case s.fldDelim:
		s.step = stateEndLiteral
		return scanEndField
	case s.comDelim:
		s.step = stateEndLiteral
		return scanEndComponent
	case s.repDelim:
		s.step = stateEndLiteral
		return scanEndRepeat
	case s.subDelim:
		s.step = stateEndLiteral
		return scanEndSubComponent
	case s.escDelim:
		s.step = stateBeginEscape
		return scanBeginEscape
		// add other delims
	}
	if _, ok := s.charDict[c]; ok {
		s.step = stateInLiteral
		return scanBeginLiteral
	}
	return s.error(c, "invalid character in message")
}

func stateBeginEscape(s *scanner, c byte) int {
	// for now, just see if A-Z
	if c >= 'A' && c <= 'Z' {
		s.step = stateInEscapeChar
		return scanContinue
	}
	return s.error(c, "invalid escape literal")
}

func stateInEscapeChar(s *scanner, c byte) int {
	if c == s.escDelim {
		s.step = stateEndEscape
		return scanEndEscape
	}
	return s.error(c, "unclosed escape delimiter")
}

func stateEndEscape(s *scanner, c byte) int {
	switch c {
	case s.segDelim:
		s.step = stateEndSegment
		return scanEndSegment
	case s.fldDelim:
		s.step = stateEndLiteral
		return scanEndField
	case s.comDelim:
		s.step = stateEndLiteral
		return scanEndComponent
	case s.repDelim:
		s.step = stateEndLiteral
		return scanEndRepeat
	case s.subDelim:
		s.step = stateEndLiteral
		return scanEndSubComponent
	case s.escDelim:
		s.step = stateBeginEscape
		return scanBeginEscape
	}
	if _, ok := s.charDict[c]; ok {
		s.step = stateInLiteral
		return scanBeginLiteral
	}
	return s.error(c, "invalid character")
}

func stateEndSegment(s *scanner, c byte) int {
	if isValidSegmentNameChar(c) {
		s.step = stateFirstSegNameChar
		return scanBeginSegmentName
	}
	return s.error(c, "expected segment name (A-Z and/or 1-9)")
}

func stateFirstSegNameChar(s *scanner, c byte) int {
	if isValidSegmentNameChar(c) {
		s.step = stateSecondSegNameChar
		return scanContinue
	}
	return s.error(c, "expected segment name (A-Z and/or 1-9)")
}

func stateSecondSegNameChar(s *scanner, c byte) int {
	if isValidSegmentNameChar(c) {
		s.step = stateThirdSegNameChar
		return scanContinue
	}
	return s.error(c, "expected segment name (A-Z and/or 1-9)")
}

func stateThirdSegNameChar(s *scanner, c byte) int {
	if c == s.fldDelim {
		s.step = stateEndLiteral
		return scanEndSegmentName
	}
	return s.error(c, "expected field delimiter")
}

func stateInLiteral(s *scanner, c byte) int {
	switch c {
	case s.segDelim:
		s.step = stateEndSegment
		return scanEndSegment
	case s.fldDelim:
		s.step = stateEndLiteral
		return scanEndField
	case s.comDelim:
		s.step = stateEndLiteral
		return scanEndComponent
	case s.repDelim:
		s.step = stateEndLiteral
		return scanEndRepeat
	case s.subDelim:
		s.step = stateEndLiteral
		return scanEndSubComponent
	case s.escDelim:
		s.step = stateBeginEscape
		return scanBeginEscape
	}
	if _, ok := s.charDict[c]; ok {
		s.step = stateInLiteral
		return scanContinue
	}
	return s.error(c, "invalid character in message")
}

func stateError(s *scanner, c byte) int {
	return scanError
}

func (s *scanner) reset() {
	s.step = stateEndLiteral
	s.err = nil
}

func (s *scanner) error(c byte, context string) int {
	s.step = stateError
	s.err = &SyntaxError{"invalid character " + quoteChar(c) + " " + context, s.bytes}
	return scanError
}

// characters in segment names are usually
// A-Z or 1-9
func isValidSegmentNameChar(c byte) bool {
	return (c >= 'A' && c <= 'Z') || (c >= '1' && c <= '9')
}

func quoteChar(c byte) string {
	if c == '\'' {
		return `'\''`
	}
	if c == '"' {
		return `'"'`
	}
	s := strconv.Quote(string(c))
	return "'" + s[1:len(s)-1] + "'"
}
