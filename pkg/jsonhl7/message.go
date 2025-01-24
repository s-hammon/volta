package jsonhl7

import (
	"bytes"
	"fmt"
	"strings"
)

/*
TODO: Turn all of this into a map[string]interface{} representation of the HL7 message,
rather than a struct. We never use the struct.

Instead of Type, we can directly reference Message["MSH.9"][0].Value (with casting, of course).
*/
type Message struct {
	Type     string    `json:"type"`
	Segments []Segment `json:"segments"`
}

func NewMessage(msg []byte) (Message, error) {
	segBytes := bytes.Split(msg, []byte("\r"))
	// get rid of any blank segments
	for i, seg := range segBytes {
		if len(seg) == 0 {
			segBytes = append(segBytes[:i], segBytes[i+1:]...)
		}
	}

	if len(segBytes) == 1 {
		return Message{}, fmt.Errorf("couldn't split segments, unrecognized line ending")
	}

	msh := segBytes[0]
	if len(msh) < 8 {
		return Message{}, fmt.Errorf("invalid MSH segment")
	}
	delimField := msh[3:8]

	delims, err := getDelimiters(delimField)
	if err != nil {
		return Message{}, err
	}

	segments := []Segment{}
	msgType := ""
	for _, seg := range segBytes {
		s, err := NewSegment(seg, delims)
		if err != nil {
			return Message{}, err
		}

		segments = append(segments, s)
		if s.Name == "MSH" {
			typeField, err := s.getField("MSH.9")
			if err != nil {
				return Message{}, err
			}
			msgType = typeField.Value.([]Hl7Obj)[0].Value.(string)
		}
	}

	return Message{Type: msgType, Segments: segments}, nil
}

// wll handle cases where we have repeated segments--AL1, NK1, OBX, etc.
// in those cases, the value of the segment will be a slice of maps
func (m *Message) Map() map[string]interface{} {
	maps := map[string]interface{}{}
	repeats := map[string][]map[string]interface{}{}

	for _, s := range m.Segments {
		if _, exists := maps[s.Name]; exists {
			if _, ok := repeats[s.Name]; ok {
				repeats[s.Name] = append(repeats[s.Name], s.Map())
			} else {
				// it is what it is
				repeats[s.Name] = []map[string]interface{}{
					maps[s.Name].(map[string]interface{}),
				}
			}
		} else {
			maps[s.Name] = s.Map()
		}
	}

	for name, segments := range repeats {
		maps[name] = segments
	}

	return maps
}

type Segment struct {
	Name   string   `json:"name"`
	Fields []Hl7Obj `json:"fields"`
}

func NewSegment(seg []byte, delims Delimiters) (Segment, error) {
	split := bytes.Split(seg, []byte{delims[0]})
	if len(split) < 2 {
		return Segment{}, fmt.Errorf("segment must have at least 2 fields: %s", string(seg))
	}

	segName := string(split[0])
	if segName == "MSH" {
		split = append([][]byte{split[0], {delims[0]}}, split[1:]...)
	}

	fields := split[1:]
	hl7Ojbs := parseObj(segName, 1, fields, delims)
	return Segment{Name: segName, Fields: hl7Ojbs}, nil
}

// Segment.Map returns a map[string]interface{} representation of the segment's fields
// since Hl7Obj.Value could be another Hl7Obj, this function should be recursive
// the base case is when Hl7Obj.Value is a string
func (s Segment) Map() map[string]interface{} {
	m := make(map[string]interface{})
	for _, f := range s.Fields {
		m = f.Map(m)
	}

	return m
}

func (s *Segment) getField(name string) (Hl7Obj, error) {
	for _, f := range s.Fields {
		if f.Name == name {
			return f, nil
		}
	}

	return Hl7Obj{}, fmt.Errorf("field %s not found", name)
}

type Hl7Obj struct {
	Name  string      `json:"name"`
	Value interface{} `json:"value"`
}

func (h *Hl7Obj) Map(m map[string]interface{}) map[string]interface{} {
	if t, ok := h.Value.(string); ok {
		m[h.Name] = t
		return m
	}

	subMap := make(map[string]interface{})
	t, _ := h.Value.([]Hl7Obj)
	for _, o := range t {
		subMap = o.Map(subMap)
	}
	m[h.Name] = subMap

	return m
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
