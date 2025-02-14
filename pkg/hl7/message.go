package hl7

import (
	"bytes"
	"errors"
	"fmt"
	"strconv"
)

var delims struct {
	field, repeat, component, subcomponent, escape byte
}

var repeatableSegments = map[string]struct{}{
	"OBX": {}, "NTE": {}, "AL1": {}, "DG1": {},
}

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
	pool.Put(w.buf)
}

type Message []byte

func NewMessage(msg []byte) (Message, error) {
	if err := extractDelimiters(msg[3:8]); err != nil {
		return nil, err
	}

	segments := bytes.Split(bytes.TrimSpace(msg), []byte{CR})
	if len(segments) < 2 {
		return nil, fmt.Errorf("couldn't split segments, unrecognized line ending")
	}

	repeatSegments := getRepeatedSegments(segments)
	w := NewMsgWriter()
	defer w.Release()

	w.buf.WriteByte('{')
	processedCounts := make(map[string]int)

	for i := 0; i < len(segments); i++ {
		seg := segments[i]
		fields := bytes.Split(seg, []byte{delims.field})
		if len(fields) < 2 {
			return nil, errors.New("segment must have at least 2 fields")
		}

		segName := string(fields[0])
		if count, ok := repeatSegments[segName]; ok {
			if processedCounts[segName] == 0 {
				if i > 0 {
					w.buf.WriteByte(',')
				}
				w.buf.WriteString(`"` + segName + `":[`)
			} else {
				w.buf.WriteByte(',')
			}
			if err := w.segmentToJSON(segName, fields[1:]); err != nil {
				return nil, err
			}
			processedCounts[segName]++
			if processedCounts[segName] == count {
				w.buf.WriteByte(']')
			}
			continue
		}

		if i > 0 {
			w.buf.WriteByte(',')
		}
		w.buf.WriteString(`"` + segName + `":`)

		if segName == HeaderSegment {
			if err := w.handleMSH(seg); err != nil {
				return nil, err
			}
		} else {
			if err := w.segmentToJSON(segName, fields[1:]); err != nil {
				return nil, err
			}
		}
	}

	w.buf.WriteByte('}')
	return w.buf.Bytes(), nil
}

func (w *MsgWriter) segmentToJSON(name string, fields [][]byte) error {
	w.buf.WriteByte('{')

	isFirst := true
	for i, f := range fields {
		if len(f) == 0 {
			continue
		}

		if !isFirst {
			w.buf.WriteByte(',')
		}
		isFirst = false

		fName := formatKey(name, i+1)
		w.buf.WriteString(`"` + fName + `":`)

		if bytes.IndexByte(f, delims.repeat) != -1 {
			w.parseRepeatsToJSON(fName, f)
		} else if bytes.IndexByte(f, delims.component) != -1 {
			w.parseComponentsToJSON(fName, f)
		} else {
			w.buf.Write(strconv.AppendQuote(nil, replaceEscapes(f)))
		}
	}

	w.buf.WriteByte('}')
	return nil
}

func (w *MsgWriter) handleMSH(segment []byte) error {
	if len(segment) < 8 {
		return errors.New("invalid MSH segment")
	}

	w.buf.WriteByte('{')

	fields := bytes.Split(segment, []byte{delims.field})
	if len(fields) < 2 {
		return errors.New("invalid MSH segment")
	}

	w.buf.WriteString(`"` + HeaderSegment + `.1":` + strconv.Quote(string(delims.field)) + `,`)
	w.buf.WriteString(`"` + HeaderSegment + `.2":` + strconv.Quote(string(fields[1])))

	for i := 2; i < len(fields); i++ {
		if len(fields[i]) == 0 {
			continue
		}
		if i > 2 {
			w.buf.WriteByte(',')
		}
		fName := formatKey(HeaderSegment, i+1)
		w.buf.WriteString(fmt.Sprintf("\"%s\":", fName))
		if bytes.Contains(fields[i], []byte{delims.component}) {
			w.parseComponentsToJSON(fName, fields[i])
			continue
		} else {
			w.buf.WriteString(fmt.Sprintf("\"%s\"", replaceEscapes(fields[i])))
		}

	}

	w.buf.WriteByte('}')
	return nil
}

func (w *MsgWriter) parseComponentsToJSON(name string, field []byte) {
	components := bytes.Split(field, []byte{delims.component})
	w.buf.WriteByte('{')

	for i, c := range components {
		if len(c) == 0 {
			continue
		}
		if i > 0 {
			w.buf.WriteByte(',')
		}
		cName := formatKey(name, i+1)
		w.buf.WriteString(fmt.Sprintf("\"%s\":\"%s\"", cName, replaceEscapes(c)))
	}

	w.buf.WriteByte('}')
}

func (w *MsgWriter) parseRepeatsToJSON(name string, field []byte) {
	repeats := bytes.Split(field, []byte{delims.repeat})
	w.buf.WriteByte('[')

	for i, r := range repeats {
		if i > 0 {
			w.buf.WriteByte(',')
		}
		if bytes.Contains(r, []byte{delims.component}) {
			w.parseComponentsToJSON(name, r)
		} else {
			w.buf.WriteString(fmt.Sprintf("\"%s\"", replaceEscapes(r)))
		}
	}

	w.buf.WriteByte(']')
}

func extractDelimiters(d []byte) error {
	if len(d) < 5 {
		return errors.New("could not extract delimiters")
	}

	delims.field = d[0]
	delims.component = d[1]
	delims.repeat = d[2]
	delims.escape = d[3]
	delims.subcomponent = d[4]

	return nil
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
			case bytes.HasPrefix(s[i:], []byte("\\F\\")):
				result = append(result, '|')
				i += 3
			case bytes.HasPrefix(s[i:], []byte("\\R\\")):
				result = append(result, '~')
				i += 3
			case bytes.HasPrefix(s[i:], []byte("\\S\\")):
				result = append(result, '^')
				i += 3
			case bytes.HasPrefix(s[i:], []byte("\\T\\")):
				result = append(result, '&')
				i += 3
			case bytes.HasPrefix(s[i:], []byte("\\E\\")):
				result = append(result, '\\')
				i += 3
			case bytes.HasPrefix(s[i:], []byte("\\X0D\\")):
				result = append(result, '\r')
				i += 5
			case bytes.HasPrefix(s[i:], []byte("\\X0A\\")):
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
		if _, ok := repeatableSegments[name]; ok {
			segCounts[name]++
		}
	}

	for k, v := range segCounts {
		if v < 2 {
			delete(segCounts, k)
		}
	}

	return segCounts
}
