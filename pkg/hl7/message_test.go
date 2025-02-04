package hl7

import (
	"bytes"
	"reflect"
	"runtime"
	"testing"
)

var validMSH = []byte("MSH|^~\\&|LabSystem|Hospital|OrderingSystem|Clinic|202501140830||ORU^R01|MSG00002|P|2.3")
var validPV1 = []byte("PV1|1|I|ICU^Room101^BedA^^Hospital||||1234^Smith^John^A^^^Dr.|||Cardiology")
var validPID = []byte("PID|1||123456^^^Hospital^MR||Doe^John^A~Doe^Johnny^B||19800101|M|||123 Main St^^Metropolis^NY^10001")

var testSegDelim = byte('\r')

var validHL7 = bytes.Join([][]byte{validMSH, validPID, validPV1}, []byte{testSegDelim})
var invalidLineEnding = bytes.Join([][]byte{validMSH, validPID, validPV1}, []byte{'\t'})
var invalidMSH = bytes.Join([][]byte{[]byte("MSH|"), validPID, validPV1}, []byte("\r"))

func TestNewMessage(t *testing.T) {
	got, err := NewMessage(validHL7, testSegDelim)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	want := Message{
		"MSH": map[string]interface{}{
			"MSH.1": "|",
			"MSH.2": "^~\\&",
			"MSH.3": "LabSystem",
			"MSH.4": "Hospital",
			"MSH.5": "OrderingSystem",
			"MSH.6": "Clinic",
			"MSH.7": "202501140830",
			"MSH.9": map[string]interface{}{
				"MSH.9.1": "ORU",
				"MSH.9.2": "R01",
			},
			"MSH.10": "MSG00002",
			"MSH.11": "P",
			"MSH.12": "2.3",
		},
		"PID": map[string]interface{}{
			"PID.1": "1",
			"PID.3": map[string]interface{}{
				"PID.3.1": "123456",
				"PID.3.4": "Hospital",
				"PID.3.5": "MR",
			},
			"PID.5": []map[string]interface{}{
				{
					"PID.5.1": "Doe",
					"PID.5.2": "John",
					"PID.5.3": "A",
				},
				{
					"PID.5.1": "Doe",
					"PID.5.2": "Johnny",
					"PID.5.3": "B",
				},
			},
			"PID.7": "19800101",
			"PID.8": "M",
			"PID.11": map[string]interface{}{
				"PID.11.1": "123 Main St",
				"PID.11.3": "Metropolis",
				"PID.11.4": "NY",
				"PID.11.5": "10001",
			},
		},
		"PV1": map[string]interface{}{
			"PV1.1": "1",
			"PV1.2": "I",
			"PV1.3": map[string]interface{}{
				"PV1.3.1": "ICU",
				"PV1.3.2": "Room101",
				"PV1.3.3": "BedA",
				"PV1.3.5": "Hospital",
			},
			"PV1.7": map[string]interface{}{
				"PV1.7.1": "1234",
				"PV1.7.2": "Smith",
				"PV1.7.3": "John",
				"PV1.7.4": "A",
				"PV1.7.7": "Dr.",
			},
			"PV1.10": "Cardiology",
		},
	}

	if !reflect.DeepEqual(got, want) {
		t.Errorf("got %v, want %v", got, want)
	}
}

func TestNewMessageError(t *testing.T) {
	tests := []struct {
		name string
		msg  []byte
	}{
		{"invalid line ending", invalidLineEnding},
		{"invalid MSH segment", invalidMSH},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got, err := NewMessage(tt.msg, testSegDelim); err == nil {
				t.Errorf("expected error, got nil\nresults: %v", got)
			}
		})
	}
}

func BenchmarkNewMessage(b *testing.B) {
	runtime.GOMAXPROCS(1)

	for i := 0; i < b.N; i++ {
		NewMessage(validHL7, testSegDelim)
	}
}
