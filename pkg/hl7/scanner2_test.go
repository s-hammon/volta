package hl7

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

var orcBytes = []byte("|CM|12345678||CRX^CHEST XRAY")
var validOBX = []byte("OBX|1|FT|CXR^Chest X-ray||diagnostic\rOBX|2|FT|CXR^Chest X-ray||more diagnostic")

func TestScanner(t *testing.T) {
	scan := newScanner2('\r', '|')
	// mock going through an MSH header
	state := scan.step(scan, 'M')
	require.Equal(t, stateSegmentName, state)
	scan.step(scan, 'S')
	scan.step(scan, 'H')
	state = scan.step(scan, '|')
	require.Equal(t, stateSegmentNameEnd, state)
	state = scan.step(scan, '^')
	require.Equal(t, stateContinue, state)
	scan.step(scan, '~')
	scan.step(scan, '\\')
	scan.step(scan, '&')
	state = scan.step(scan, '|')
	require.Equal(t, stateFieldValEnd, state)
	state = scan.step(scan, '\r')
	require.Equal(t, stateSegmentEnd, state)

	state = scan.step(scan, '\r')
	require.Equal(t, -1, state)
}

func TestFieldNode(t *testing.T) {
	// suppose that this is the index of the first field delimiter
	// it will always be the 0th position if the first 3 bytes (segment name) are removed from the segment byte slice
	// ORC-1 (2 characters)
	orc1 := &fieldNode{bytIdx: 0}
	// ORC-2 (8 characters)
	orc2 := &fieldNode{bytIdx: 3}
	// ORC-3 (0 characters)
	orc3 := &fieldNode{bytIdx: 12}
	orc4 := &fieldNode{bytIdx: 13}

	orcMap := fieldMap{}
	_ = orcMap.setFieldNodeIdx(1, orc1)
	val, exists := orcMap.getFieldNodeIdx(1)
	require.True(t, exists)
	require.NotNil(t, val)
	require.Equal(t, 0, val.bytIdx)

	_ = orcMap.setFieldNodeIdx(2, orc2)
	val, exists = orcMap.getFieldNodeIdx(2)
	require.True(t, exists)
	require.Equal(t, orc2, val)

	_ = orcMap.setFieldNodeIdx(3, orc3)
	val, exists = orcMap.getFieldNodeIdx(3)
	require.True(t, exists)
	require.Equal(t, orc3, val)

	_ = orcMap.setFieldNodeIdx(4, orc4)
	val, exists = orcMap.getFieldNodeIdx(4)
	require.True(t, exists)
	require.Equal(t, orc4, val)

	n, _ := orcMap.getFieldNodeIdx(1)
	n.next = orc2
	require.NotNil(t, n.next)
	assert.Equal(t, "CM", string(orcBytes[n.bytIdx+1:n.next.bytIdx]))
}

func TestSegment(t *testing.T) {
	orc := NewSegment("ORC")
	scan := &hl7Scanner{}

	exists := orc.AddField(1, 0, scan)
	require.True(t, !exists)
	require.NotNil(t, orc.fields[1])
	require.Nil(t, orc.fields[1].next)

	exists = orc.AddField(2, 3, scan)
	require.True(t, !exists)
	require.NotNil(t, orc.fields[2])
	require.Nil(t, orc.fields[2].next)
	require.NotNil(t, orc.fields[1].next)

	assert.Equal(t, 0, orc.GetFieldIdx(1))
	n := orc.GetFieldNode(1)
	assert.Equal(t, 3, n.next.bytIdx)
	assert.Equal(t, "CM", string(orcBytes[n.bytIdx+1:n.next.bytIdx]))

	assert.Nil(t, orc.GetFieldNode(420))
	assert.Equal(t, -1, orc.GetFieldIdx(69))
}

func TestDecoder(t *testing.T) {
	dec := newDecoder()
	err := dec.init(validMSH, '\r')
	require.NoError(t, err)
	require.True(t, dec.scan.fldDelim == '|')
	require.True(t, dec.scan.comDelim == '^')
	require.True(t, dec.scan.repDelim == '~')
	require.True(t, dec.scan.escDelim == '\\')
	require.True(t, dec.scan.subDelim == '&')

	require.Equal(t, 1, len(dec.segments))
	segIdc, exists := dec.segMap.getSegmentIdx("MSH")
	require.True(t, exists)
	require.NotNil(t, segIdc)
	require.IsType(t, 1, segIdc)

	segment, ok := dec.segments[segIdc.(int)]
	require.True(t, ok)
	require.NotNil(t, segment)

	assert.Equal(t, 12, len(segment.fields))

	assert.Equal(t, "|", dec.getFieldVal("MSH", 1, 0))
	assert.Equal(t, "^~\\&", dec.getFieldVal("MSH", 2, 0))
	assert.Equal(t, "ORU^R01", dec.getFieldVal("MSH", 9, 0))
	assert.Equal(t, "", dec.getFieldVal("PID", 2, 0))
}

func TestDecoderRepeatSegments(t *testing.T) {
	dec := newDecoder()
	err := dec.init(validOBX, '\r')
	require.NoError(t, err)

	require.Equal(t, 2, len(dec.segments))
	indices, exists := dec.segMap.getSegmentIdx("OBX")
	require.True(t, exists)
	require.NotNil(t, indices)
	require.IsType(t, []int{}, indices)
	require.Equal(t, 2, len(indices.([]int)))

	assert.Equal(t, "1", dec.getFieldVal("OBX", 1, 0))
	assert.Equal(t, "diagnostic", dec.getFieldVal("OBX", 5, 0))
	assert.Equal(t, "2", dec.getFieldVal("OBX", 1, 1))
	assert.Equal(t, "more diagnostic", dec.getFieldVal("OBX", 5, 1))
	assert.Equal(t, "", dec.getFieldVal("OBX", 1, 2))
}
