package hl7

import (
	"bytes"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

var validMSH = []byte("MSH|^~\\&|LabSystem|Hospital|OrderingSystem|Clinic|202501140830||ORU^R01|MSG00002|P|2.3")
var validPV1 = []byte("PV1|1|I|ICU^Room101^BedA^^Hospital||||1234^Smith^John^A^^^Dr.|||Cardiology")
var validPID = []byte("PID|1||123456^^^Hospital^MR||Doe^John^A~Doe^Johnny^B||19800101|M|||123 Main St^^Metropolis^NY^10001")

var invalidLineEnding = bytes.Join([][]byte{validMSH, validPID, validPV1}, []byte{'\t'})
var invalidMSH = bytes.Join([][]byte{[]byte("MSH"), validPID, validPV1}, []byte{CR})
var firstNotMSH = bytes.Join([][]byte{validPID, validMSH}, []byte{CR})
var invalidSegmentName = []byte("ID|1||123456^^^Hospital^MR||Doe^John^A~Doe^Johnny^B||19800101|M|||123 Main St^^Metropolis^NY^10001")

func BenchmarkNewMessageAll(b *testing.B) {
	entries, err := HL7.ReadDir("test_hl7")
	if err != nil {
		b.Fatalf("failed to read embedded test directory: %v", err)
	}

	var messages [][]byte
	for _, entry := range entries {
		data, err := HL7.ReadFile(filepath.Join("test_hl7", entry.Name()))
		if err != nil {
			b.Fatalf("failed to read test file %s: %v", entry.Name(), err)
		}
		if len(data) == 0 {
			b.Fatalf("file %s is empty", entry.Name())
		}
		messages = append(messages, data)
	}

	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			for _, msg := range messages {
				if _, err := NewMessage(msg); err != nil {
					b.Fatalf("unexpected error in parsing: %v", err)
				}
			}
		}
	})
}

func BenchmarkNewMessageEach(b *testing.B) {
	entries, err := HL7.ReadDir("test_hl7")
	if err != nil {
		b.Fatalf("failed to read embedded test directory: %v", err)
	}

	for _, entry := range entries {
		data, err := HL7.ReadFile(filepath.Join("test_hl7", entry.Name()))
		if err != nil {
			b.Fatalf("failed to read test file %s: %v", entry.Name(), err)
		}
		if len(data) == 0 {
			b.Fatalf("file %s is empty", entry.Name())
		}

		b.Run(entry.Name(), func(b *testing.B) {
			b.ReportAllocs()
			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				if _, err := NewMessage(data); err != nil {
					b.Fatalf("unexpected error in parsing: %v", err)
				}
			}
		})
	}
}

func TestHL7Files(t *testing.T) {
	t.Parallel()

	entries, err := HL7.ReadDir("test_hl7")
	if err != nil {
		t.Fatalf("failed to read embedded test directory: %v", err)
	}

	for i, entry := range entries {
		if entry.IsDir() || !strings.HasSuffix(entry.Name(), ".hl7") {
			continue
		}
		t.Run(entry.Name(), func(t *testing.T) {
			fPath := filepath.Join("test_hl7", entry.Name())
			data, err := HL7.ReadFile(fPath)
			if err != nil {
				t.Fatalf("couldn't read test file at %s: %v", fPath, err)
			}
			if len(data) == 0 {
				t.Fatalf("file %s is empty", fPath)
			}
			msg, err := NewMessage(data)
			if err != nil {
				t.Fatalf("failed to parse HL7 message from file %s: %v", fPath, err)
			}
			if !testValidHL7(msg) {
				t.Fatalf("parsed HL7 message from file %s is not valid JSON:\n%s", fPath, msg)
			}

			wantFileName := strings.TrimSuffix(entry.Name(), ".hl7")
			wantFilePath := filepath.Join("test_hl7", fmt.Sprintf("%s.json", wantFileName))
			if _, err := os.Stat(wantFilePath); err != nil {
				if !os.IsNotExist(err) {
					t.Fatalf("failed to check expected file %s: %v", wantFilePath, err)
				}
				savePath := filepath.Join("test_hl7", fmt.Sprintf("%d.json", i+1))
				if err := os.WriteFile(savePath, msg, 0644); err != nil {
					t.Fatalf("failed to write expected file %s: %v", savePath, err)
				}
				t.Logf("⚠️  Expected file not found. Wrote generated output to: %s. Please review this JSON file and resubmit tests.\n", savePath)
			}
			want, err := os.ReadFile(wantFilePath)
			if err != nil {
				t.Fatalf("couldn't read expected file at %s: %v", wantFilePath, err)
			}

			if !bytes.Equal(msg, want) {
				t.Fatalf("parsed HL7 message from file %s does not match expected JSON file %s:\n%s\n", fPath, wantFilePath, msg)
			}
		})
	}
}

func testValidHL7(data []byte) bool {
	var t any
	return json.Unmarshal(data, &t) == nil
}

func TestNewMessageError(t *testing.T) {
	tests := []struct {
		name string
		msg  []byte
	}{
		{"message too short", []byte("MSH")},
		{"invalid line ending", invalidLineEnding},
		{"invalid MSH segment", invalidMSH},
		{"first segment not MSH", firstNotMSH},
		{"invalid segment name", invalidSegmentName},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got, err := NewMessage(tt.msg); err == nil {
				t.Fatalf("expected error, got nil\nresults: %s", got)
			}
		})
	}
}
