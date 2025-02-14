package hl7

import (
	"bytes"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	json "github.com/json-iterator/go"
)

var delims = struct {
	field        byte
	repeat       byte
	component    byte
	subcomponent byte
	escape       byte
}{}

const (
	fieldDelimIdx = iota
	repeatDelimIdx
	componentDelimIdx
	subcomponentDelimIdx
	escapeDelimIdx
)

const (
	HeaderSegment = "MSH"
	CR            = '\r'
)

type Message []byte

func NewMessage(msg []byte) (Message, error) {
	if err := extractDelimiters(msg[3:8]); err != nil {
		return nil, err
	}

	segments := bytes.Split(bytes.TrimSpace(msg), []byte{CR})
	if len(segments) < 2 {
		return nil, fmt.Errorf("couldn't split segments, unrecognized line ending")
	}

	var buf bytes.Buffer
	buf.WriteByte('{')

	for i, seg := range segments {
		if i > 0 {
			buf.WriteByte(',')
		}
		segFields := bytes.Split(seg, []byte{delims.field})
		if len(segFields) < 2 {
			return nil, errors.New("segment must have at least 2 fields")
		}

		segName := string(segFields[0])
		buf.WriteString(fmt.Sprintf("\"%s\":", segName))
		if segName == HeaderSegment {
			segmentJSON, err := handleMSH(seg)
			if err != nil {
				return nil, err
			}
			buf.Write(segmentJSON)
			continue
		}

		segmentJSON, err := segmentToJSON(segName, segFields[1:])
		if err != nil {
			return nil, err
		}
		buf.Write(segmentJSON)
	}

	buf.WriteByte('}')

	return buf.Bytes(), nil
}

func FromJSON(filename string) (Message, error) {
	filename = filepath.Clean(filename)
	if !strings.HasPrefix(filename, "/") {
		panic(fmt.Errorf("unsafe input"))
	}
	msg, err := os.ReadFile(filename)
	if err != nil {
		return nil, err
	}

	// marshal to Message map
	var m Message
	if err := json.Unmarshal(msg, &m); err != nil {
		return nil, err
	}

	return m, nil
}

// func (m Message) Type() string {
// 	msh, ok := m["MSH"].(map[string]interface{})
// 	if !ok {
// 		return ""
// 	}
// 	typeField, ok := msh["MSH.9"].(map[string]interface{})
// 	if !ok {
// 		return ""
// 	}
// 	msgType, ok := typeField["MSH.9.1"].(string)
// 	if !ok {
// 		return ""
// 	}

// 	return msgType
// }

func segmentToJSON(name string, fields [][]byte) ([]byte, error) {
	var buf bytes.Buffer
	buf.WriteByte('{')

	for i, f := range fields {
		if len(f) == 0 {
			continue
		}

		if i > 0 {
			buf.WriteByte(',')
		}
		fName := fmt.Sprintf("%s.%d", name, i+1)
		buf.WriteString(fmt.Sprintf("\"%s\":", fName))

		if bytes.Contains(f, []byte{delims.repeat}) {
			repeats, err := parseRepeatsToJSON(fName, f)
			if err != nil {
				return nil, err
			}
			buf.Write(repeats)
			continue
		}
		if bytes.Contains(f, []byte{delims.component}) {
			components := parseComponentsToJSON(fName, f)
			buf.Write(components)
			continue
		}

		buf.WriteString(fmt.Sprintf("\"%s\"", replaceEscapes(f)))
	}

	buf.WriteByte('}')
	return buf.Bytes(), nil
}

func handleMSH(segment []byte) ([]byte, error) {
	if len(segment) < 8 {
		return nil, errors.New("invalid MSH segment")
	}

	var buf bytes.Buffer
	buf.WriteByte('{')

	fields := bytes.Split(segment, []byte{delims.field})
	if len(fields) < 2 {
		return nil, errors.New("invalid MSH segment")
	}

	buf.WriteString(fmt.Sprintf("\"%s.1\":\"%s\",", HeaderSegment, string(delims.field)))
	buf.WriteString(fmt.Sprintf("\"%s.2\":\"%s\",", HeaderSegment, string(fields[1])))

	for i := 2; i < len(fields); i++ {
		if len(fields[i]) == 0 {
			continue
		}
		if i > 2 {
			buf.WriteByte(',')
		}
		fName := fmt.Sprintf("%s.%d", HeaderSegment, i+1)
		buf.WriteString(fmt.Sprintf("\"%s\":", fName))
		if bytes.Contains(fields[i], []byte{delims.component}) {
			components := parseComponentsToJSON(fName, fields[i])
			buf.Write(components)
			continue
		} else {
			buf.WriteString(fmt.Sprintf("\"%s\"", replaceEscapes(fields[i])))
		}

	}

	buf.WriteByte('}')
	return buf.Bytes(), nil
}

func parseComponentsToJSON(name string, field []byte) []byte {
	components := bytes.Split(field, []byte{delims.component})
	var buf bytes.Buffer
	buf.WriteByte('{')

	for i, c := range components {
		if len(c) == 0 {
			continue
		}
		if i > 0 {
			buf.WriteByte(',')
		}
		cName := fmt.Sprintf("%s.%d", name, i+1)
		buf.WriteString(fmt.Sprintf("\"%s\":\"%s\"", cName, replaceEscapes(c)))
	}

	buf.WriteByte('}')
	return buf.Bytes()
}

func parseRepeatsToJSON(name string, field []byte) ([]byte, error) {
	repeats := bytes.Split(field, []byte{delims.repeat})
	var buf bytes.Buffer
	buf.WriteByte('[')

	for i, r := range repeats {
		if i > 0 {
			buf.WriteByte(',')
		}
		if bytes.Contains(r, []byte{delims.component}) {
			buf.Write(parseComponentsToJSON(name, r))
		} else {
			buf.WriteString(fmt.Sprintf("\"%s\"", replaceEscapes(r)))
		}
	}

	buf.WriteByte(']')
	return buf.Bytes(), nil
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

func replaceEscapes(s []byte) string {
	if !bytes.Contains(s, []byte{'\\'}) {
		return string(s)
	}

	escapes := map[string]string{
		"\\F\\":   "|",
		"\\R\\":   "~",
		"\\S\\":   "^",
		"\\T\\":   "&",
		"\\E\\":   "\\",
		"\\X0D\\": "\r",
		"\\X0A\\": "\n",
	}

	var buf bytes.Buffer
	i := 0
	for i < len(s) {
		if s[i] == '\\' {
			found := false
			for k, v := range escapes {
				if bytes.HasPrefix(s[i:], []byte(k)) {
					buf.WriteString(v)
					i += len(k)
					found = true
					break
				}
			}
			if !found {
				buf.WriteByte(s[i])
				i++
			}
		} else {
			buf.WriteByte(s[i])
			i++
		}
	}
	return buf.String()
}
