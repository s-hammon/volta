package hl7

import "fmt"

const messageHeader = "MSH"

// a node representing a field
type fieldNode struct {
	bytIdx int
	next   *fieldNode
}

const (
	stateContinue int = iota
	stateSegmentName
	stateSegmentNameEnd
	stateFieldDelimiter
	stateSegmentEnd
	stateFieldValEnd
)

type hl7Scanner struct {
	lastIdx int
	step    func(scan *hl7Scanner, c byte) int

	segDelim, fldDelim, comDelim, repDelim, escDelim, subDelim byte
}

func newScanner2(sd, fd byte) *hl7Scanner {
	return &hl7Scanner{
		lastIdx:  0,
		step:     scanSegmentName,
		segDelim: sd,
		fldDelim: fd,
	}
}

func scanSegmentName(scan *hl7Scanner, c byte) int {
	if validSegNameChar(c) {
		scan.step = scanSegmentName
		return stateSegmentName
	}
	if c == scan.fldDelim {
		scan.step = scanFieldVal
		return stateSegmentNameEnd
	}
	return -1
}

func scanFieldVal(scan *hl7Scanner, c byte) int {
	if c == scan.fldDelim {
		return stateFieldValEnd
	}
	if c == scan.segDelim {
		scan.step = scanSegmentName
		return stateSegmentEnd
	}
	// still in field
	return stateContinue
}

func validSegNameChar(c byte) bool {
	return (c >= 'A' && c <= 'Z') || (c >= '1' && c <= '9')
}

// keys are the 1-based HL7 index of each field delimiter
// values are the field node representing the byte index
// they may point to a succeeding field node
type fieldMap map[int]*fieldNode

func newFieldMap() fieldMap { return fieldMap{} }

// idx1 is the 1-based HL7 index of the field delimiter
// returns true if there is already a node registered for the given idx1
func (m fieldMap) setFieldNodeIdx(idx1 int, n *fieldNode) (exists bool) {
	_, exists = m[idx1]
	if !exists {
		m[idx1] = n
	}
	return exists
}

func (m fieldMap) getFieldNodeIdx(idx1 int) (n *fieldNode, exists bool) {
	val, exists := m[idx1]
	return val, exists
}

type segment struct {
	name   string
	size   int
	fields fieldMap
}

func NewSegment(s string) *segment {
	return &segment{
		name:   s,
		fields: newFieldMap(),
	}
}

// idx1 is the 1-based HL7 index of the field delimiter
// idx0 is the 0-based index of the field delimiter (in the raw byte slice
// AddField will set the scanner's lastIdx field to the current one
func (s *segment) AddField(idx1, idx0 int, scan *hl7Scanner) (exists bool) {
	n := &fieldNode{bytIdx: idx0}
	if prev, ok := s.fields.getFieldNodeIdx(scan.lastIdx); ok {
		prev.next = n
	}
	exists = s.fields.setFieldNodeIdx(idx1, n)
	if !exists {
		scan.lastIdx = idx1
	}
	return exists
}

func (s *segment) GetFieldNode(idx1 int) *fieldNode {
	if n, ok := s.fields.getFieldNodeIdx(idx1); ok {
		return n
	}
	return nil
}

func (s *segment) GetFieldIdx(idx1 int) int {
	idx, ok := s.fields.getFieldNodeIdx(idx1)
	if !ok {
		return -1
	}
	return idx.bytIdx
}

// keys are the 3-letter segment names (MSH, PID, OBX, etc)
// values are the 0-based index (or indices) of the first pipe delimiter (in the raw byte slice)
// in other words, it can be an int or []int
type segmentMap map[string]any

func newSegmentMap() segmentMap { return map[string]any{} }

func (m segmentMap) getSegmentIdx(s string) (val any, exists bool) {
	val, exists = m[s]
	return val, exists
}

func (m segmentMap) addSegment(s string, idx0 int) {
	value, ok := m[s]
	if ok {
		switch v := any(value).(type) {
		case []int:
			m[s] = append(v, idx0)
		case int:
			m[s] = []int{v, idx0}
		default:
			panic("whooooooooooops!")
		}
		return
	} else {
		m[s] = idx0
		return
	}
}

type decoder struct {
	data       []byte
	start, off int

	lastState int
	length    int
	segMap    segmentMap       // this converts the name to the (list of) zero-based idx
	segments  map[int]*segment // key is zero-based idx of segment
	scan      *hl7Scanner
}

func newDecoder() *decoder {
	return &decoder{
		segMap:   newSegmentMap(),
		segments: make(map[int]*segment),
	}
}

func (d *decoder) init(data []byte, segDelim byte) error {
	l := len(data)
	if l < 8 {
		return fmt.Errorf("message is too short (length: %d)\n", len(data))
	}
	d.data = data
	d.off = 0
	d.length = l
	d.scan = &hl7Scanner{
		step:     scanSegmentName,
		segDelim: segDelim,
		fldDelim: data[3],
		comDelim: data[4],
		repDelim: data[5],
		escDelim: data[6],
		subDelim: data[7],
	} // switch to pool

	idx1 := 1
	currentSegIdx := 0
	currentSegName := ""
	var currentSegFields *segment
	for d.off < d.length {
		s, i := d.scan, d.off
		state := s.step(s, d.data[i])
		d.off++
		switch state {
		case stateSegmentName:
			d.scanWhile(stateSegmentName)
			if d.lastState != stateSegmentNameEnd {
				panic("huh!?")
			}
			currentSegIdx = d.start
			currentSegName = string(d.data[currentSegIdx : d.off-1])
			currentSegFields = NewSegment(currentSegName)
			if currentSegName == "MSH" {
				idx1++
			}
			if exists := currentSegFields.AddField(idx1, d.off-1, d.scan); exists {
				panic("this shouldn't happen!")
			}
			idx1++
		case stateSegmentEnd:
			currentSegFields.size = d.off - currentSegIdx - 1
			d.segMap.addSegment(currentSegName, currentSegIdx)
			d.segments[currentSegIdx] = currentSegFields
			idx1 = 1
			d.start = d.off
			currentSegName = ""
			currentSegFields = nil
		case stateFieldValEnd:
			if exists := currentSegFields.AddField(idx1, d.off-1, d.scan); exists {
				panic("this shouldn't happen!")
			}
			idx1++
		case stateContinue:
			continue
		default:
			panic(fmt.Sprintf("unrecognized state! got: '%d'\n", state))
		}
	}
	if exists := currentSegFields.AddField(idx1, d.off, d.scan); exists {
		panic("this shouldn't happen!")
	}
	d.segMap.addSegment(currentSegName, d.start)
	d.segments[currentSegIdx] = currentSegFields
	return nil
}

func (d *decoder) scanWhile(state int) {
	s, data, i := d.scan, d.data, d.off
	for i < len(data) {
		newState := s.step(s, data[i])
		i++
		if newState != state {
			d.off = i
			d.lastState = newState
			return
		}
	}
	d.off = len(data)
}

// n is the "nth" segment repeat
func (d *decoder) getFieldVal(s string, idx1, n int) string {
	if s == messageHeader {
		if idx1 == 1 {
			return string(d.scan.fldDelim)
		}
		if idx1 == 2 {
			return fmt.Sprintf(
				"%c%c%c%c",
				d.scan.comDelim,
				d.scan.repDelim,
				d.scan.escDelim,
				d.scan.subDelim,
			)
		}
	}
	segIdx, exists := d.segMap.getSegmentIdx(s)
	if !exists {
		return ""
	}

	switch segIdx.(type) {
	case int:
		return d.scanField(idx1, segIdx.(int))
	case []int:
		if len(segIdx.([]int)) <= max(0, n) {
			return ""
		}
		indices := segIdx.([]int)
		return d.scanField(idx1, indices[n])
	default:
		panic("unknown index type")
	}
}

func (d *decoder) scanField(idx1, idx0 int) string {
	var start, end int
	segment, ok := d.segments[idx0]
	if ok {
		n, exists := segment.fields.getFieldNodeIdx(idx1)
		if exists {
			start = n.bytIdx + 1
			if n.next != nil {
				end = n.next.bytIdx
			} else {
				end = segment.size
			}
			fmt.Printf("start: %d\tend: %d\n", start, end)
			return string(d.data[start:end])
		}
	}
	return ""
}

func max(m, n int) int {
	if n > m {
		return n
	}
	return m
}
