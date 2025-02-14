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

type Message map[string]interface{}

func NewMessage(msg []byte) (Message, error) {
	if err := extractDelimiters(msg[3:8]); err != nil {
		return nil, err
	}

	segments := bytes.Split(bytes.TrimSpace(msg), []byte{CR})
	if len(segments) < 2 {
		return nil, fmt.Errorf("couldn't split segments, unrecognized line ending")
	}

	message := make(map[string]interface{}, len(segments))
	var repeatSegments []map[string]interface{}

	for _, seg := range segments {
		segFields := bytes.Split(seg, []byte{delims.field})
		if len(segFields) < 2 {
			return nil, errors.New("segment must have at least 2 fields")
		}

		segName := string(segFields[0])
		if segName == HeaderSegment {
			parsed, err := handleMSH(seg)
			if err != nil {
				return nil, err
			}

			message[HeaderSegment] = parsed
			continue
		}

		fields := segFields[1:]

		parsed, err := parseSegment(segName, fields)
		if err != nil {
			return nil, err
		}
		if segName == "OBX" {
			repeatSegments = append(repeatSegments, parsed)
			continue
		}
		message[segName] = parsed
	}

	if len(repeatSegments) > 0 {
		message["OBX"] = repeatSegments
	}

	return message, nil
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

func (m Message) Type() string {
	msh, ok := m["MSH"].(map[string]interface{})
	if !ok {
		return ""
	}
	typeField, ok := msh["MSH.9"].(map[string]interface{})
	if !ok {
		return ""
	}
	msgType, ok := typeField["MSH.9.1"].(string)
	if !ok {
		return ""
	}

	return msgType
}

func parseSegment(name string, fields [][]byte) (map[string]interface{}, error) {
	parsed := make(map[string]interface{}, len(fields))

	for i, field := range fields {
		if len(field) == 0 {
			continue
		}

		fName := fmt.Sprintf("%s.%d", name, i+1)
		if bytes.Contains(field, []byte{delims.repeat}) {
			// handle repeats
			repeats, err := parseRepeats(fName, bytes.Split(field, []byte{delims.repeat}))
			if err != nil {
				return nil, err
			}
			parsed[fName] = repeats
			continue
		}
		// if components := bytes.Split(field, []byte{delims.component}); len(components) > 1 {
		if bytes.Contains(field, []byte{delims.component}) {
			parsed[fName] = parseComponents(fName, bytes.Split(field, []byte{delims.component}))
			continue
		}

		parsed[fName] = replaceEscapes(field)
	}

	return parsed, nil
}

func handleMSH(segment []byte) (map[string]interface{}, error) {
	if len(segment) < 8 {
		return nil, errors.New("invalid MSH segment")
	}

	msh := make(map[string]interface{})

	fields := bytes.Split(segment, []byte{delims.field})
	if len(fields) < 2 {
		return nil, errors.New("invalid MSH segment")
	}
	msh[fmt.Sprintf("%s.1", HeaderSegment)] = string(delims.field)
	msh[fmt.Sprintf("%s.2", HeaderSegment)] = string(fields[1])

	for i := 2; i < len(fields); i++ {
		if len(fields[i]) == 0 {
			continue
		}

		fName := fmt.Sprintf("%s.%d", HeaderSegment, i+1)
		// if components := bytes.Split(field, []byte{delims.component}); len(components) > 1 {
		if bytes.Contains(fields[i], []byte{delims.component}) {
			msh[fName] = parseComponents(fName, bytes.Split(fields[i], []byte{delims.component}))
			continue
		}

		msh[fName] = replaceEscapes(fields[i])
	}

	return msh, nil
}

func parseFields(name string, data [][]byte) map[string]interface{} {
	fields := make(map[string]interface{}, len(data))
	for i, d := range data {
		if len(d) == 0 {
			continue
		}

		fName := fmt.Sprintf("%s.%d", name, i+1)
		if components := bytes.Split(d, []byte{delims.component}); len(components) > 1 {
			fields[fName] = parseComponents(fName, components)
			continue
		}

		fields[fName] = replaceEscapes(d)
	}

	return fields
}

func parseComponents(name string, data [][]byte) map[string]interface{} {
	components := make(map[string]interface{}, len(data))
	for i, d := range data {
		if len(d) == 0 {
			continue
		}

		pName := fmt.Sprintf("%s.%d", name, i+1)
		if subcomponents := bytes.Split(d, []byte{delims.subcomponent}); len(subcomponents) > 1 {
			// handle subcomponents
		}
		components[pName] = replaceEscapes(d)
	}

	return components
}

func parseRepeats(name string, data [][]byte) ([]map[string]interface{}, error) {
	repeats := make([]map[string]interface{}, 0, len(data))

	for _, d := range data {
		if len(d) == 0 {
			continue
		}

		rMap := make(map[string]interface{})

		if bytes.Contains(d, []byte{delims.component}) {
			rMap = parseComponents(name, bytes.Split(d, []byte{delims.component}))
		} else {
			rMap[name] = replaceEscapes(d)
		}

		repeats = append(repeats, rMap)
	}

	return repeats, nil
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
