package hl7

const messageHeader = "MSH"

// a node representing a field
type fieldNode struct {
	idx  int
	next *fieldNode
}

const (
	stateBegin scanState = iota
	stateContinue
	stateSegmentName
	stateSegmentNameEnd
	stateFieldDelimiter
	stateSegmentEnd
	stateFieldValEnd
	stateDone

	stateErr
)

type scanState int
type scanner struct {
	lastIdx            int
	step               scanStep
	segDelim, fldDelim byte
	comDelim, repDelim byte
	escDelim, subDelim byte
}

type scanStep func(scan *scanner, c byte) scanState

func NewScanner(segmentDelim, fieldDelim byte) *scanner {
	return &scanner{
		step:     scanSegmentName,
		segDelim: segmentDelim,
		fldDelim: fieldDelim,
	}
}

func scanSegmentName(scan *scanner, c byte) scanState {
	switch {
	case isValidSegNameChar(c):
		return stateSegmentName
	case c == scan.fldDelim:
		scan.step = scanFieldVal
		return stateSegmentNameEnd
	default:
		return stateErr
	}
}

func scanFieldVal(scan *scanner, c byte) scanState {
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

func isValidSegNameChar(c byte) bool {
	return (c >= 'A' && c <= 'Z') || (c >= '1' && c <= '9')
}

// keys are the 1-based HL7 index of each field delimiter
// values are the field node representing the byte index
// they may point to a succeeding field node
type fieldMap map[int]*fieldNode

func newFieldMap() fieldMap { return make(fieldMap) }

// idx is the 1-based HL7 index of the field delimiter
// returns true if there is already a node registered for the given idx1
func (m fieldMap) setFieldNode(idx int, n *fieldNode) (exists bool) {
	_, exists = m[idx]
	if !exists {
		m[idx] = n
	}
	return exists
}

// idx is the 1-based HL7 index of the field delimiter
// if idx doesn't exist, exists returns false
func (m fieldMap) getFieldNode(idx int) (n *fieldNode, exists bool) {
	val, exists := m[idx]
	return val, exists
}

type segment struct {
	name   string
	endIdx int
	fields fieldMap
}

func NewSegment(s string) *segment {
	return &segment{name: s, fields: newFieldMap()}
}

// hl7Idx is the 1-based HL7 index of the field delimiter
// bytIdx is the 0-based index of the field delimiter (in the raw byte slice
// AddField will set the scanner's lastIdx field to the current one
func (s *segment) AddField(hl7Idx, bytIdx int, scan *scanner) (exists bool) {
	n := &fieldNode{idx: bytIdx}
	if prev, ok := s.fields.getFieldNode(scan.lastIdx); ok {
		prev.next = n
	}
	exists = s.fields.setFieldNode(hl7Idx, n)
	if !exists {
		scan.lastIdx = hl7Idx
	}
	return exists
}

func (s *segment) GetFieldIdx(hl7Idx int) int {
	idx, ok := s.fields.getFieldNode(hl7Idx)
	if !ok {
		return -1
	}
	return idx.idx
}

// keys are the 3-letter segment names (MSH, PID, OBX, etc)
// values are the segment structure of each occurring segment
// for the given key
type segmentMap map[string][]int

func newSegmentMap() segmentMap { return make(segmentMap) }

func (m segmentMap) getSegmentIndices(s string) (indices []int, exists bool) {
	indices, exists = m[s]
	return indices, exists
}

func (m segmentMap) addSegment(s string, idx int) {
	m[s] = append(m[s], idx)
}
