package hl7

import (
	"bytes"
	"errors"
	"fmt"
	"strconv"
)

var (
	fieldDelim, componentDelim, repeatDelim, escapeDelim, subComponentDelim byte
)

const (
	HeaderSegment = "MSH"
	CR            = '\r'
)

type MsgWriter struct {
	buf *bytes.Buffer
}

func NewMsgWriter() *MsgWriter {
	buf := pool.Get().(*bytes.Buffer)
	buf.Reset()
	return &MsgWriter{buf: buf}
}

func (w *MsgWriter) Release() {
	w.buf.Reset()
	pool.Put(w.buf)
}

type Message []byte

func NewMessage(msg []byte) (Message, error) {
	if len(msg) < 8 {
		return nil, fmt.Errorf("message too short: %d bytes", len(msg))
	}
	extractDelimiters(msg[3:8])

	segments := bytes.Split(bytes.TrimSpace(msg), []byte{CR})
	if len(segments) < 2 {
		return nil, fmt.Errorf("couldn't split segments, unrecognized line ending")
	}
	repeatSegments := getRepeatedSegments(segments)

	w := NewMsgWriter()
	defer w.Release()

	processedCounts := make(map[string]*bytes.Buffer)
	w.buf.WriteByte('{')

	firstSegment := true
	for i, seg := range segments {
		fields := bytes.Split(seg, []byte{fieldDelim})
		if len(fields) < 2 {
			return nil, errors.New("segment must have at least 2 fields")
		}

		segName := string(fields[0])
		if i == 0 {
			if segName != HeaderSegment {
				return nil, fmt.Errorf("first segment must be %s, got %s", HeaderSegment, segName)
			}
			if err := w.handleMSH(seg); err != nil {
				return nil, err
			}
			firstSegment = false
			continue
		}
		if len(segName) != 3 {
			return nil, fmt.Errorf("invalid segment name: %s", segName)
		}

		if _, ok := repeatSegments[segName]; ok {
			rBuf := processedCounts[segName]
			if rBuf == nil {
				rBuf = pool.Get().(*bytes.Buffer)
				rBuf.Reset()
				rBuf.WriteString(`"` + segName + `":[`)
			} else {
				rBuf.WriteString(`,`)
			}
			if err := segmentToJSON(rBuf, segName, fields[1:]); err != nil {
				pool.Put(rBuf)
				return nil, err
			}
			repeatSegments[segName]--
			if repeatSegments[segName] == 0 {
				rBuf.WriteByte(']')
				if !firstSegment {
					w.buf.WriteByte(',')
				}
				w.buf.Write(rBuf.Bytes())
				pool.Put(rBuf)
				delete(repeatSegments, segName)
				firstSegment = false
			} else {
				processedCounts[segName] = rBuf
			}
			continue
		}
		if !firstSegment {
			w.buf.WriteByte(',')
		}
		firstSegment = false

		w.buf.WriteString(`"` + segName + `":`)
		if err := segmentToJSON(w.buf, segName, fields[1:]); err != nil {
			return nil, err
		}
	}

	w.buf.WriteByte('}')
	return w.buf.Bytes(), nil
}

func segmentToJSON(buf *bytes.Buffer, name string, fields [][]byte) error {
	buf.WriteByte('{')

	isFirst := true
	for i, f := range fields {
		if len(f) == 0 {
			continue
		}

		if !isFirst {
			buf.WriteByte(',')
		}
		isFirst = false

		fName := formatKey(name, i+1)
		buf.WriteString(`"` + fName + `":`)

		if bytes.IndexByte(f, repeatDelim) != -1 {
			parseRepeatsToJSON(buf, fName, f)
		} else if bytes.IndexByte(f, componentDelim) != -1 {
			parseComponentsToJSON(buf, fName, f)
		} else {
			buf.Write(strconv.AppendQuote(nil, replaceEscapes(f)))
		}
	}

	buf.WriteByte('}')
	return nil
}

func (w *MsgWriter) handleMSH(segment []byte) error {
	w.buf.WriteString(`"` + HeaderSegment + `":{`)

	fields := bytes.Split(segment, []byte{fieldDelim})
	if len(fields) < 2 {
		return errors.New("invalid MSH segment")
	}

	w.buf.WriteString(`"` + HeaderSegment + `.1":` + strconv.Quote(string(fieldDelim)) + `,`)
	w.buf.WriteString(`"` + HeaderSegment + `.2":` + strconv.Quote(string(fields[1])))

	for i := 2; i < len(fields); i++ {
		if len(fields[i]) == 0 {
			continue
		}
		w.buf.WriteByte(',')
		fName := formatKey(HeaderSegment, i+1)
		w.buf.WriteString(`"` + fName + `":`)
		if bytes.Contains(fields[i], []byte{componentDelim}) {
			parseComponentsToJSON(w.buf, fName, fields[i])
			continue
		} else {
			w.buf.WriteString(strconv.Quote(replaceEscapes(fields[i])))
		}

	}

	w.buf.WriteByte('}')
	return nil
}

func parseComponentsToJSON(buf *bytes.Buffer, name string, field []byte) {
	components := bytes.Split(field, []byte{componentDelim})
	buf.WriteByte('{')

	isFirst := true
	for i, c := range components {
		if len(c) == 0 {
			continue
		}
		if isFirst {
			isFirst = false
		} else {
			buf.WriteByte(',')
		}
		cName := formatKey(name, i+1)
		buf.WriteString(`"` + cName + `":`)

		if bytes.IndexByte(c, subComponentDelim) != -1 {
			parseSubComponentsToJSON(buf, cName, c)
		} else {
			buf.Write(strconv.AppendQuote(nil, replaceEscapes(c)))
		}
	}

	buf.WriteByte('}')
}

func parseSubComponentsToJSON(buf *bytes.Buffer, name string, component []byte) {
	subComponents := bytes.Split(component, []byte{subComponentDelim})
	buf.WriteByte('{')

	isFirst := true
	for i, s := range subComponents {
		if len(s) == 0 {
			continue
		}
		if isFirst {
			isFirst = false
		} else {
			buf.WriteByte(',')
		}
		sName := formatKey(name, i+1)
		buf.WriteString(fmt.Sprintf("\"%s\":\"%s\"", sName, replaceEscapes(s)))
	}

	buf.WriteByte('}')
}

func parseRepeatsToJSON(buf *bytes.Buffer, name string, field []byte) {
	repeats := bytes.Split(field, []byte{repeatDelim})
	buf.WriteByte('[')

	for i, r := range repeats {
		if i > 0 {
			buf.WriteByte(',')
		}
		if bytes.Contains(r, []byte{componentDelim}) {
			parseComponentsToJSON(buf, name, r)
		} else {
			buf.WriteString(fmt.Sprintf("\"%s\"", replaceEscapes(r)))
		}
	}

	buf.WriteByte(']')
}

func extractDelimiters(d []byte) {
	fieldDelim = d[0]
	componentDelim = d[1]
	repeatDelim = d[2]
	escapeDelim = d[3]
	subComponentDelim = d[4]
}

func formatKey(name string, i int) string {
	return name + "." + strconv.Itoa(i)
}

func replaceEscapes(s []byte) string {
	if !bytes.Contains(s, []byte{'\\'}) {
		return string(s)
	}

	result := make([]byte, 0, len(s))

	i := 0
	for i < len(s) {
		if s[i] == '\\' && i+2 < len(s) {
			switch {
			case bytes.HasPrefix(s[i:], []byte{escapeDelim, 'F', escapeDelim}):
				result = append(result, '|')
				i += 3
			case bytes.HasPrefix(s[i:], []byte{escapeDelim, 'R', escapeDelim}):
				result = append(result, '~')
				i += 3
			case bytes.HasPrefix(s[i:], []byte{escapeDelim, 'S', escapeDelim}):
				result = append(result, '^')
				i += 3
			case bytes.HasPrefix(s[i:], []byte{escapeDelim, 'T', escapeDelim}):
				result = append(result, '&')
				i += 3
			case bytes.HasPrefix(s[i:], []byte{escapeDelim, 'E', escapeDelim}):
				result = append(result, '\\')
				i += 3
			case bytes.HasPrefix(s[i:], []byte{escapeDelim, '\x0D', escapeDelim}):
				result = append(result, '\r')
				i += 5
			case bytes.HasPrefix(s[i:], []byte{escapeDelim, '\x0A', escapeDelim}):
				result = append(result, '\n')
				i += 5
			default:
				result = append(result, s[i])
				i++
			}
		} else {
			result = append(result, s[i])
			i++
		}
	}
	return string(result)
}

func getRepeatedSegments(segments [][]byte) map[string]int {
	segCounts := make(map[string]int)

	for _, seg := range segments {
		if len(seg) < 3 {
			continue
		}

		name := string(seg[:3])
		if _, ok := segCounts[name]; !ok {
			segCounts[name] = 0
		}
		segCounts[name]++
	}

	for k, v := range segCounts {
		if v < 2 {
			delete(segCounts, k)
		}
	}

	return segCounts
}
