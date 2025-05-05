package hl7

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

var validOBX = []byte("MSH|^~\\&|LabSystem|Hospital|OrderingSystem|Clinic|202501140830||ORU^R01|MSG00002|P|2.3\rOBX|1|FT|CXR^Chest X-ray||diagnostic\rOBX|2|FT|CXR^Chest X-ray||more diagnostic")
var validOBXRepField = []byte("MSH|^~\\&|LabSystem|Hospital|OrderingSystem|Clinic|202501140830||ORU^R01|MSG00002|P|2.3\rOBR|1|1234|1234|CXR^Chest X-ray|S|20250114082000|||12~mg/mL\rOBX|1|FT|CXR^Chest X-ray||diagnostic\rOBX|2|FT|CXR^Chest X-ray||more diagnostic")

func TestScanner(t *testing.T) {
	scan := NewScanner('\r', '|')
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
	require.Equal(t, stateErr, state)
}

func TestDecoder(t *testing.T) {
	dec := newDecoder()
	dec.init(validMSH, '\r')
	require.NoError(t, dec.savedError)
	require.True(t, dec.scan.fldDelim == '|')
	require.True(t, dec.scan.comDelim == '^')
	require.True(t, dec.scan.repDelim == '~')
	require.True(t, dec.scan.escDelim == '\\')
	require.True(t, dec.scan.subDelim == '&')

	require.Equal(t, 1, len(dec.segments))
	segIdc, exists := dec.segMap.getSegmentIndices("MSH")
	require.True(t, exists)
	require.NotNil(t, segIdc)
	require.IsType(t, []int{}, segIdc)

	segment, ok := dec.segments[segIdc[0]]
	require.True(t, ok)
	require.NotNil(t, segment)

	assert.Equal(t, 12, len(segment.fields))

	assert.Equal(t, "|", dec.getFieldVal("MSH", 1, 0))
	assert.Equal(t, "^~\\&", dec.getFieldVal("MSH", 2, 0))
	assert.Equal(t, "LabSystem", dec.getFieldVal("MSH", 3, 0))
	assert.Equal(t, "ORU^R01", dec.getFieldVal("MSH", 9, 0))
	assert.Equal(t, "2.3", dec.getFieldVal("MSH", 12, 0))
	assert.Equal(t, "", dec.getFieldVal("PID", 2, 0))
}

func TestDecoderRepeatSegments(t *testing.T) {
	dec := newDecoder()
	dec.init(validOBX, '\r')
	require.NoError(t, dec.savedError)

	require.Equal(t, 3, len(dec.segments))
	indices, exists := dec.segMap.getSegmentIndices("OBX")
	require.True(t, exists)
	require.NotNil(t, indices)
	require.IsType(t, []int{}, indices)
	require.Equal(t, 2, len(indices))

	assert.Equal(t, "1", dec.getFieldVal("OBX", 1, 0))
	assert.Equal(t, "diagnostic", dec.getFieldVal("OBX", 5, 0))
	assert.Equal(t, "2", dec.getFieldVal("OBX", 1, 1))
	assert.Equal(t, "more diagnostic", dec.getFieldVal("OBX", 5, 1))
	assert.Equal(t, "", dec.getFieldVal("OBX", 1, 2))
}
