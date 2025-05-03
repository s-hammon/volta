package hl7

import "fmt"

const messageHeader = "MSH"

// a node representing a field
type fieldNode struct {
	bytIdx int
	next   *fieldNode
}

const (
	stateContinue scanState = iota
	stateSegmentName
	stateSegmentNameEnd
	stateFieldDelimiter
	stateSegmentEnd
	stateFieldValEnd

	stateErr
)

type scanState int
type scanStep func(scan *hl7Scanner, c byte) scanState

type hl7Scanner struct {
	lastIdx            int
	step               scanStep
	segDelim, fldDelim byte
	comDelim, repDelim byte
	escDelim, subDelim byte
}

func newScanner2(sd, fd byte) *hl7Scanner {
	return &hl7Scanner{
		lastIdx:  0,
		step:     scanSegmentName,
		segDelim: sd,
		fldDelim: fd,
	}
}

func scanSegmentName(scan *hl7Scanner, c byte) scanState {
	switch {
	case validSegNameChar(c):
		return stateSegmentName
	case c == scan.fldDelim:
		scan.step = scanFieldVal
		return stateSegmentNameEnd
	default:
		return stateErr
	}
}

func scanFieldVal(scan *hl7Scanner, c byte) scanState {
	switch c {
	case scan.fldDelim:
		return stateFieldValEnd
	case scan.segDelim:
		scan.step = scanSegmentName
		return stateSegmentEnd
	// TODO: add case for true utf-8 so we can default panic ;)
	default:
		return stateContinue
	}
}

func validSegNameChar(c byte) bool {
	return (c >= 'A' && c <= 'Z') || (c >= '1' && c <= '9')
}

// keys are the 1-based HL7 index of each field delimiter
// values are the field node representing the byte index
// they may point to a succeeding field node
type fieldMap map[int]*fieldNode

func newFieldMap() fieldMap { return fieldMap{} }

// idx is the 1-based HL7 index of the field delimiter
// returns true if there is already a node registered for the given idx1
func (m fieldMap) setFieldNodeIdx(idx int, n *fieldNode) (exists bool) {
	_, exists = m[idx]
	if !exists {
		m[idx] = n
	}
	return exists
}

// idx is the 1-based HL7 index of the field delimiter
// if idx doesn't exist, exists returns false
func (m fieldMap) getFieldNodeIdx(idx int) (n *fieldNode, exists bool) {
	val, exists := m[idx]
	return val, exists
}

type segment struct {
	name   string
	size   int
	fields fieldMap
}

func NewSegment(s string) *segment {
	return &segment{name: s, fields: newFieldMap()}
}

// hl7Idx is the 1-based HL7 index of the field delimiter
// bytIdx is the 0-based index of the field delimiter (in the raw byte slice
// AddField will set the scanner's lastIdx field to the current one
func (s *segment) AddField(hl7Idx, bytIdx int, scan *hl7Scanner) (exists bool) {
	n := &fieldNode{bytIdx: bytIdx}
	if prev, ok := s.fields.getFieldNodeIdx(scan.lastIdx); ok {
		prev.next = n
	}
	exists = s.fields.setFieldNodeIdx(hl7Idx, n)
	if !exists {
		scan.lastIdx = hl7Idx
	}
	return exists
}

func (s *segment) GetFieldNode(hl7Idx int) *fieldNode {
	if n, ok := s.fields.getFieldNodeIdx(hl7Idx); ok {
		return n
	}
	return nil
}

func (s *segment) GetFieldIdx(hl7Idx int) int {
	idx, ok := s.fields.getFieldNodeIdx(hl7Idx)
	if !ok {
		return -1
	}
	return idx.bytIdx
}

// keys are the 3-letter segment names (MSH, PID, OBX, etc)
// values are the segment structure of each occurring segment
// for the given key
type segmentMap map[string][]int

func newSegmentMap() segmentMap { return make(segmentMap) }

func (m segmentMap) getSegmentIdx(s string) (val []int, exists bool) {
	val, exists = m[s]
	return val, exists
}

func (m segmentMap) addSegment(s string, idx int) {
	m[s] = append(m[s], idx)
}

type decoder struct {
	data       []byte
	start, off int
	lastState  scanState
	length     int

	scan     *hl7Scanner
	segMap   segmentMap       // this converts the name to the (list of) zero-based idx
	segments map[int]*segment // key is zero-based idx of segment
}

func newDecoder() *decoder {
	return &decoder{
		segMap:   newSegmentMap(),
		segments: make(map[int]*segment),
	}
}

func (d *decoder) init(data []byte, segDelim byte) error {
	if len(data) < 8 {
		return fmt.Errorf("message is too short (length: %d)\n", len(data))
	}
	d.data = data
	l := len(data)
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

	var (
		idx1             = 1
		currentSegIdx    int
		currentSegName   string
		currentSegFields *segment
	)

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
			if currentSegName == messageHeader {
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

func (d *decoder) scanWhile(state scanState) {
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
		switch idx1 {
		case 1:
			return string(d.scan.fldDelim)
		case 2:
			return fmt.Sprintf("%c%c%c%c", d.scan.comDelim, d.scan.repDelim, d.scan.escDelim, d.scan.subDelim)
		}
	}
	indices, found := d.segMap.getSegmentIdx(s)
	if !found || n >= len(indices) {
		return ""
	}
	return d.scanField(idx1, indices[n])
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
