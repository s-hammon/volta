package jsonhl7

import (
	"bytes"
	"reflect"
	"testing"
)

var validMSH = []byte("MSH|^~\\&|LabSystem|Hospital|OrderingSystem|Clinic|202501140830||ORU^R01|MSG00002|P|2.3|")
var validPV1 = []byte("PV1|1|I|ICU^Room101^BedA^^Hospital||||1234^Smith^John^A^^^Dr.|||Cardiology|")
var validPID = []byte("PID|1||123456^^^Hospital^MR||Doe^John^A~Doe^Johnny^B||19800101|M|||123 Main St^^Metropolis^NY^10001|")
var validEsc = []byte("AL1|1|Penicillin \\T\\ Amoxicillin|P\\S\\A|Severe rash \\F\\ Anaphylaxis||2024\\E\\01\\E\\01|")
var testDelims = Delimiters{0: '|', 1: '~', 2: '^', 3: '&', 4: '\\'}

var validHL7 = bytes.Join([][]byte{validMSH, validPID, validPV1}, []byte("\r"))
var invalidLineEnding = bytes.Join([][]byte{validMSH, validPID, validPV1}, []byte{'\t'})
var invalidMSH = bytes.Join([][]byte{[]byte("MSH|"), validPID, validPV1}, []byte("\r"))

func TestNewMessage(t *testing.T) {
	got, err := NewMessage(validHL7)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	want := Message{
		Segments: []Segment{
			{
				Name: "MSH",
				Fields: []Hl7Obj{
					{Name: "MSH.1", Value: "|"},
					{Name: "MSH.2", Value: "^~\\&"},
					{Name: "MSH.3", Value: "LabSystem"},
					{Name: "MSH.4", Value: "Hospital"},
					{Name: "MSH.5", Value: "OrderingSystem"},
					{Name: "MSH.6", Value: "Clinic"},
					{Name: "MSH.7", Value: "202501140830"},
					{Name: "MSH.9", Value: []Hl7Obj{
						{Name: "MSH.9.1", Value: "ORU"},
						{Name: "MSH.9.2", Value: "R01"},
					}},
					{Name: "MSH.10", Value: "MSG00002"},
					{Name: "MSH.11", Value: "P"},
					{Name: "MSH.12", Value: "2.3"},
				},
			},
			{
				Name: "PID",
				Fields: []Hl7Obj{
					{Name: "PID.1", Value: "1"},
					{Name: "PID.3", Value: []Hl7Obj{
						{Name: "PID.3.1", Value: "123456"},
						{Name: "PID.3.4", Value: "Hospital"},
						{Name: "PID.3.5", Value: "MR"},
					}},
					{Name: "PID.5", Value: []Hl7Obj{
						{Name: "PID.5(1)", Value: []Hl7Obj{
							{Name: "PID.5(1).1", Value: "Doe"},
							{Name: "PID.5(1).2", Value: "John"},
							{Name: "PID.5(1).3", Value: "A"},
						}},
						{Name: "PID.5(2)", Value: []Hl7Obj{
							{Name: "PID.5(2).1", Value: "Doe"},
							{Name: "PID.5(2).2", Value: "Johnny"},
							{Name: "PID.5(2).3", Value: "B"},
						}},
					}},
					{Name: "PID.7", Value: "19800101"},
					{Name: "PID.8", Value: "M"},
					{Name: "PID.11", Value: []Hl7Obj{
						{Name: "PID.11.1", Value: "123 Main St"},
						{Name: "PID.11.3", Value: "Metropolis"},
						{Name: "PID.11.4", Value: "NY"},
						{Name: "PID.11.5", Value: "10001"},
					}},
				},
			},
			{
				Name: "PV1",
				Fields: []Hl7Obj{
					{Name: "PV1.1", Value: "1"},
					{Name: "PV1.2", Value: "I"},
					{Name: "PV1.3", Value: []Hl7Obj{
						{Name: "PV1.3.1", Value: "ICU"},
						{Name: "PV1.3.2", Value: "Room101"},
						{Name: "PV1.3.3", Value: "BedA"},
						{Name: "PV1.3.5", Value: "Hospital"},
					}},
					{Name: "PV1.7", Value: []Hl7Obj{
						{Name: "PV1.7.1", Value: "1234"},
						{Name: "PV1.7.2", Value: "Smith"},
						{Name: "PV1.7.3", Value: "John"},
						{Name: "PV1.7.4", Value: "A"},
						{Name: "PV1.7.7", Value: "Dr."},
					}},
					{Name: "PV1.10", Value: "Cardiology"},
				},
			},
		},
	}

	if !reflect.DeepEqual(want, got) {
		t.Errorf("want\n%+v\ngot\n%+v", want, got)
	}
}

func TestNewMessageError(t *testing.T) {
	tests := []struct {
		name string
		msg  []byte
	}{
		{
			name: "invalid line ending",
			msg:  invalidLineEnding,
		},
		{
			name: "invalid MSH segment",
			msg:  invalidMSH,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if _, err := NewMessage(tt.msg); err == nil {
				t.Fatalf("expected error, got nil")
			}
		})
	}
}

func TestNewSegment(t *testing.T) {
	tests := []struct {
		name string
		seg  []byte
		want Segment
	}{
		{
			name: "valid MSH segment",
			seg:  validMSH,
			want: Segment{
				Name: "MSH",
				Fields: []Hl7Obj{
					{Name: "MSH.1", Value: "|"},
					{Name: "MSH.2", Value: "^~\\&"},
					{Name: "MSH.3", Value: "LabSystem"},
					{Name: "MSH.4", Value: "Hospital"},
					{Name: "MSH.5", Value: "OrderingSystem"},
					{Name: "MSH.6", Value: "Clinic"},
					{Name: "MSH.7", Value: "202501140830"},
					{Name: "MSH.9", Value: []Hl7Obj{
						{Name: "MSH.9.1", Value: "ORU"},
						{Name: "MSH.9.2", Value: "R01"},
					}},
					{Name: "MSH.10", Value: "MSG00002"},
					{Name: "MSH.11", Value: "P"},
					{Name: "MSH.12", Value: "2.3"},
				},
			},
		},
		{
			name: "valid PV1 segment (contains components)",
			seg:  validPV1,
			want: Segment{
				Name: "PV1",
				Fields: []Hl7Obj{
					{Name: "PV1.1", Value: "1"},
					{Name: "PV1.2", Value: "I"},
					{Name: "PV1.3", Value: []Hl7Obj{
						{Name: "PV1.3.1", Value: "ICU"},
						{Name: "PV1.3.2", Value: "Room101"},
						{Name: "PV1.3.3", Value: "BedA"},
						{Name: "PV1.3.5", Value: "Hospital"},
					}},
					{Name: "PV1.7", Value: []Hl7Obj{
						{Name: "PV1.7.1", Value: "1234"},
						{Name: "PV1.7.2", Value: "Smith"},
						{Name: "PV1.7.3", Value: "John"},
						{Name: "PV1.7.4", Value: "A"},
						{Name: "PV1.7.7", Value: "Dr."},
					}},
					{Name: "PV1.10", Value: "Cardiology"},
				},
			},
		},
		{
			name: "valid PID segment (contains field repeats)",
			seg:  validPID,
			want: Segment{
				Name: "PID",
				Fields: []Hl7Obj{
					{Name: "PID.1", Value: "1"},
					{Name: "PID.3", Value: []Hl7Obj{
						{Name: "PID.3.1", Value: "123456"},
						{Name: "PID.3.4", Value: "Hospital"},
						{Name: "PID.3.5", Value: "MR"},
					}},
					{Name: "PID.5", Value: []Hl7Obj{
						{Name: "PID.5(1)", Value: []Hl7Obj{
							{Name: "PID.5(1).1", Value: "Doe"},
							{Name: "PID.5(1).2", Value: "John"},
							{Name: "PID.5(1).3", Value: "A"},
						}},
						{Name: "PID.5(2)", Value: []Hl7Obj{
							{Name: "PID.5(2).1", Value: "Doe"},
							{Name: "PID.5(2).2", Value: "Johnny"},
							{Name: "PID.5(2).3", Value: "B"},
						}},
					}},
					{Name: "PID.7", Value: "19800101"},
					{Name: "PID.8", Value: "M"},
					{Name: "PID.11", Value: []Hl7Obj{
						{Name: "PID.11.1", Value: "123 Main St"},
						{Name: "PID.11.3", Value: "Metropolis"},
						{Name: "PID.11.4", Value: "NY"},
						{Name: "PID.11.5", Value: "10001"},
					}},
				},
			},
		},
		{
			name: "valid segment with escape sequences",
			seg:  validEsc,
			want: Segment{
				Name: "AL1",
				Fields: []Hl7Obj{
					{Name: "AL1.1", Value: "1"},
					{Name: "AL1.2", Value: "Penicillin & Amoxicillin"},
					{Name: "AL1.3", Value: "P^A"},
					{Name: "AL1.4", Value: "Severe rash | Anaphylaxis"},
					{Name: "AL1.6", Value: "2024\\01\\01"},
				},
			},
		},
		// TODO: add an example with subcomponents
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := NewSegment(tt.seg, testDelims)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			if !reflect.DeepEqual(tt.want, got) {
				t.Errorf("want\n%+v\ngot\n%+v", tt.want, got)
			}
		})
	}
}

func TestNewSegmentError(t *testing.T) {
	tests := []struct {
		name string
		seg  []byte
	}{
		{
			name: "empty segment",
			seg:  []byte(""),
		},
		{
			name: "segment with no fields",
			seg:  []byte("PID"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if _, err := NewSegment(tt.seg, testDelims); err == nil {
				t.Fatalf("expected error, got nil")
			}
		})
	}
}
