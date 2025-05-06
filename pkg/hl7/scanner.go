package hl7

import (
	"bytes"
	"fmt"
)

type fieldPos struct {
	start, end int
}

func FastScan(data []byte, segDelim, fldDelim byte) ([]*segment, error) {
	var segments []*segment
	var start, i int

	for i < len(data) {
		start = i
		end := bytes.IndexByte(data[i:], segDelim)
		if end == -1 {
			end = len(data)
		} else {
			end += i
		}
		line := data[start:end]
		fields := bytes.Split(line, []byte{fldDelim})
		if len(fields) == 0 || len(fields[0]) != 3 {
			return nil, fmt.Errorf("invalid segment: %s", fields[0])
		}
		seg := &segment{name: string(fields[0]), endIdx: end}
		offset := start + len(fields[0]) + 1
		for _, f := range fields[1:] {
			if offset+len(f) > len(data) {
				break
			}
			seg.fields = append(seg.fields, fieldPos{
				start: offset,
				end:   offset + len(f),
			})
			offset += len(f) + 1
		}
		segments = append(segments, seg)
		i = end + 1
	}
	return segments, nil
}

type segment struct {
	name   string
	fields []fieldPos
	endIdx int
}

func (s *segment) GetField(data []byte, idx int) string {
	if idx < 1 || idx > len(s.fields) {
		return ""
	}
	pos := s.fields[idx-1]
	return string(data[pos.start:pos.end])
}

func GetSegments(segments []*segment, name string) []*segment {
	var result []*segment
	for _, seg := range segments {
		if seg.name == name {
			result = append(result, seg)
		}
	}
	return result
}
