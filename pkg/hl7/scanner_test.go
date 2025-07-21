package hl7

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

var validOBX = []byte("MSH|^~\\&|LabSystem|Hospital|OrderingSystem|Clinic|202501140830||ORU^R01|MSG00002|P|2.3\rOBX|1|FT|CXR^Chest X-ray||diagnostic\rOBX|2|FT|CXR^Chest X-ray||more diagnostic")

func TestScanner(t *testing.T) {
	segs, err := FastScan(validOBX, '\r', '|')
	require.NoError(t, err)
	require.Equal(t, 3, len(segs))

	msh := segs[0]
	assert.Equal(t, "MSH", msh.name)
	assert.GreaterOrEqual(t, len(msh.fields), 11)
	assert.Equal(t, "^~\\&", msh.GetField(validOBX, 1))
	assert.Equal(t, "ORU^R01", msh.GetField(validOBX, 8))

	obx1 := segs[1]
	assert.Equal(t, "OBX", obx1.name)
	assert.Equal(t, "1", obx1.GetField(validOBX, 1))
	assert.Equal(t, "diagnostic", obx1.GetField(validOBX, 5))

	obx2 := segs[2]
	assert.Equal(t, "OBX", obx2.name)
	assert.Equal(t, "2", obx2.GetField(validOBX, 1))
	assert.Equal(t, "more diagnostic", obx2.GetField(validOBX, 5))
}

func BenchmarkFastScan(b *testing.B) {
	data, err := HL7.ReadFile("test_hl7/9.hl7")
	if err != nil {
		b.Fatal(err)
	}
	segDelim := byte('\r')
	fieldDelim := data[3]

	b.ReportAllocs()
	b.ResetTimer()
	for b.Loop() {
		_, err = FastScan(data, segDelim, fieldDelim)
		if err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkGetField(b *testing.B) {
	data, err := HL7.ReadFile("test_hl7/9.hl7")
	if err != nil {
		b.Fatal(err)
	}
	segs, err := FastScan(data, '\r', data[3])
	if err != nil {
		b.Fatal(err)
	}
	var seg *segment
	for _, s := range segs {
		if s.name == "PV1" {
			seg = s
			break
		}
	}
	if seg == nil {
		b.Fatal("PV1 segment not found")
	}

	b.ReportAllocs()
	b.ResetTimer()
	for b.Loop() {
		_ = seg.GetField(data, 3)
	}
}
