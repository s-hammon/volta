package hl7

import (
	"bytes"
	"fmt"
	"os"

	json "github.com/json-iterator/go"
)

const (
	fieldDelimIdx = iota
	repeatDelimIdx
	componentDelimIdx
	subcomponentDelimIdx
	escapeDelimIdx
)

type Message map[string]interface{}

func NewMessage(msg []byte, segDelim byte) (Message, error) {
	segments := bytes.Split(msg, []byte{segDelim})

	segBytes := segments[:0]
	for _, seg := range segments {
		if len(bytes.TrimSpace(seg)) > 0 {
			segBytes = append(segBytes, seg)
		}
	}
	if len(segBytes) < 2 {
		return nil, fmt.Errorf("couldn't split segments, unrecognized line ending")
	}

	msh := segBytes[0]
	if len(msh) < 8 {
		return nil, fmt.Errorf("invalid MSH segment")
	}
	delimiters := extractDelimiters(msh)

	message := getMsgMap()
	var repeatSegments []map[string]interface{}

	for _, seg := range segBytes {
		segFields := bytes.Split(seg, delimiters[0])
		if len(segFields) < 2 {
			putMsgMap(message)
			return nil, fmt.Errorf("segment must have at least 2 fields: %s", string(seg))
		}

		segName := string(segFields[0])
		fields := segFields[1:]
		if segName == "MSH" {
			fields = append([][]byte{delimiters[0]}, fields...)
		}

		parsed, err := parseSegment(segName, fields, delimiters)
		if err != nil {
			putMsgMap(message)
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

func parseSegment(name string, fields [][]byte, delimiters map[int][]byte) (map[string]interface{}, error) {
	parsed := getMsgMap()

	for i, field := range fields {
		if len(field) == 0 {
			continue
		}

		fName := fmt.Sprintf("%s.%d", name, i+1)
		if name == "MSH" && i == 1 {
			parsed[fName] = string(field)
			continue
		}

		if bytes.Contains(field, delimiters[repeatDelimIdx]) {
			repeats, err := parseRepeats(fName, field, delimiters)
			if err != nil {
				return nil, err
			}
			parsed[fName] = repeats
			continue
		}

		if err := parseObj(fName, field, componentDelimIdx, parsed, delimiters); err != nil {
			return nil, err
		}
	}

	return parsed, nil
}

func parseRepeats(name string, field []byte, delimiters map[int][]byte) ([]map[string]interface{}, error) {
	repeats := []map[string]interface{}{}
	parts := bytes.Split(field, delimiters[repeatDelimIdx])
	for _, part := range parts {
		rMap := getMsgMap()
		if err := parseObj(name, part, componentDelimIdx, rMap, delimiters); err != nil {
			return nil, err
		}

		if parsed, ok := rMap[name].(map[string]interface{}); ok {
			repeats = append(repeats, parsed)
		}
	}

	return repeats, nil
}

func parseObj(name string, obj []byte, delimIdx int, parentMap map[string]interface{}, delimiters map[int][]byte) error {
	if delimIdx > len(delimiters) {
		return fmt.Errorf("delimiter index out of bounds")
	}

	subObjs := bytes.Split(obj, delimiters[delimIdx])
	// base case
	if len(subObjs) < 2 {
		parentMap[name] = replaceEscapes(obj)
		return nil
	}

	parsed := make(map[string]interface{}, len(subObjs))
	for i, subObj := range subObjs {
		if len(subObj) == 0 {
			continue
		}

		subObjName := fmt.Sprintf("%s.%d", name, i+1)
		if err := parseObj(subObjName, subObj, delimIdx+1, parsed, delimiters); err != nil {
			return err
		}
	}

	parentMap[name] = parsed
	return nil
}

func extractDelimiters(msh []byte) map[int][]byte {
	return map[int][]byte{
		fieldDelimIdx:        {msh[3]}, // field
		repeatDelimIdx:       {msh[5]}, // repeat
		componentDelimIdx:    {msh[4]}, // component
		subcomponentDelimIdx: {msh[7]}, // subcomponent
		escapeDelimIdx:       {msh[6]}, // escape
	}
}

func replaceEscapes(s []byte) string {
	s = bytes.ReplaceAll(s, []byte("\\F\\"), []byte("|"))
	s = bytes.ReplaceAll(s, []byte("\\S\\"), []byte("^"))
	s = bytes.ReplaceAll(s, []byte("\\R\\"), []byte("~"))
	s = bytes.ReplaceAll(s, []byte("\\T\\"), []byte("&"))
	s = bytes.ReplaceAll(s, []byte("\\E\\"), []byte("\\"))
	s = bytes.ReplaceAll(s, []byte("\\X0D\\"), []byte("\r"))
	s = bytes.ReplaceAll(s, []byte("\\X0A\\"), []byte("\n"))

	return string(s)
}
