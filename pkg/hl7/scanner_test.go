package hl7

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

var (
	validMSH         = []byte("MSH|^~\\&|LabSystem|Hospital|OrderingSystem|Clinic|202501140830||ORU^R01|MSG00002|P|2.3")
	validOBX         = []byte("MSH|^~\\&|LabSystem|Hospital|OrderingSystem|Clinic|202501140830||ORU^R01|MSG00002|P|2.3\rOBX|1|FT|CXR^Chest X-ray||diagnostic\rOBX|2|FT|CXR^Chest X-ray||more diagnostic")
	validOBXRepField = []byte("MSH|^~\\&|LabSystem|Hospital|OrderingSystem|Clinic|202501140830||ORU^R01|MSG00002|P|2.3\rOBR|1|1234|1234|CXR^Chest X-ray|S|20250114082000|||12~mg/mL\rOBX|1|FT|CXR^Chest X-ray||diagnostic\rOBX|2|FT|CXR^Chest X-ray||more diagnostic")
)

func TestScanner(t *testing.T) {
	segs, err := FastScan(validOBX, '\r', '|')
	require.NoError(t, err)
	require.Equal(t, 3, len(segs))

	msh := segs[0]
	assert.Equal(t, "MSH", msh.name)
	assert.GreaterOrEqual(t, len(msh.fields), 12)
	assert.Equal(t, "^~\\&", msh.GetField(validOBX, 2))
	assert.Equal(t, "ORU^R01", msh.GetField(validOBX, 9))

	obx1 := segs[1]
	assert.Equal(t, "OBX", obx1.name)
	assert.Equal(t, "1", obx1.GetField(validOBX, 1))
	assert.Equal(t, "diagnostic", obx1.GetField(validOBX, 5))

	obx2 := segs[2]
	assert.Equal(t, "OBX", obx1.name)
	assert.Equal(t, "2", obx2.GetField(validOBX, 1))
	assert.Equal(t, "more diagnostic", obx1.GetField(validOBX, 5))
}
