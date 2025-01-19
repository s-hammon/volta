package jsonhl7

import (
	"bytes"
	"fmt"
	"strings"
)

type Message struct {
	Segments []Segment `json:"segments"`
}

func NewMessage(msg []byte) (Message, error) {
	segBytes := bytes.Split(msg, []byte("\r"))
	if len(segBytes) == 1 {
		segBytes = bytes.Split(msg, []byte("\n"))
		if len(segBytes) == 1 {
			return Message{}, fmt.Errorf("couldn't split segments, unrecognized line ending")
		}
	}

	msh := segBytes[0]
	if len(msh) < 8 {
		return Message{}, fmt.Errorf("invalid MSH segment")
	}
	delimField := segBytes[0][3:8]

	delims, err := getDelimiters(delimField)
	if err != nil {
		return Message{}, err
	}

	segments := []Segment{}
	for _, seg := range segBytes {
		s, err := NewSegment(seg, delims)
		if err != nil {
			return Message{}, err
		}

		segments = append(segments, s)
	}

	return Message{Segments: segments}, nil
}

type Segment struct {
	Name   string   `json:"name"`
	Fields []Hl7Obj `json:"fields"`
}

func NewSegment(seg []byte, delims Delimiters) (Segment, error) {
	split := bytes.Split(seg, []byte{delims[0]})
	if len(split) < 2 {
		return Segment{}, fmt.Errorf("segment must have at least 2 fields")
	}

	segName := string(split[0])
	if segName == "MSH" {
		split = append([][]byte{split[0], {delims[0]}}, split[1:]...)
	}

	fields := split[1:]
	hl7Ojbs := parseObj(segName, 1, fields, delims)
	return Segment{Name: segName, Fields: hl7Ojbs}, nil
}

type Hl7Obj struct {
	Name  string      `json:"name"`
	Value interface{} `json:"value"`
}

func parseObj(name string, delimIdx int, objs [][]byte, delimiters Delimiters) []Hl7Obj {
	values := []Hl7Obj{}
	for i, obj := range objs {
		if len(obj) == 0 {
			continue
		}

		if name == "MSH" && i == 1 {
			values = append(values, Hl7Obj{
				Name:  newObjName(name, delimIdx, i),
				Value: string(obj),
			})
			continue
		}

		subObjs := bytes.Split(obj, []byte{delimiters[delimIdx]})
		if len(subObjs) > 1 {
			subObj := Hl7Obj{
				Name:  newObjName(name, delimIdx, i),
				Value: parseObj(newObjName(name, delimIdx, i), delimIdx+1, subObjs, delimiters),
			}
			values = append(values, subObj)
			continue
		}

		if delimIdx == 1 {
			subObjs := bytes.Split(obj, []byte{delimiters[delimIdx+1]})
			if len(subObjs) > 1 {
				subObj := Hl7Obj{
					Name:  newObjName(name, delimIdx, i),
					Value: parseObj(newObjName(name, delimIdx, i), delimIdx+2, subObjs, delimiters),
				}
				values = append(values, subObj)
				continue
			}
		}

		values = append(values, Hl7Obj{
			Name:  newObjName(name, delimIdx, i),
			Value: replaceEscapes(string(obj)),
		})
	}

	return values
}

func newObjName(name string, idx, i int) string {
	if idx-1 == 1 {
		return fmt.Sprintf("%s(%d)", name, i+1)
	}

	return fmt.Sprintf("%s.%d", name, i+1)
}

func replaceEscapes(s string) string {
	replacer := strings.NewReplacer(
		"\\F\\", "|",
		"\\S\\", "^",
		"\\R\\", "~",
		"\\T\\", "&",
		"\\E\\", "\\",
		"\\X0D\\", "\r",
		"\\X0A\\", "\n",
	)

	return replacer.Replace(s)
}
