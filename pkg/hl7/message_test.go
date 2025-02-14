package hl7

import (
	"bytes"
	"fmt"
	"runtime"
	"testing"
)

var validMSH = []byte("MSH|^~\\&|LabSystem|Hospital|OrderingSystem|Clinic|202501140830||ORU^R01|MSG00002|P|2.3")
var validPV1 = []byte("PV1|1|I|ICU^Room101^BedA^^Hospital||||1234^Smith^John^A^^^Dr.|||Cardiology")
var validPID = []byte("PID|1||123456^^^Hospital^MR||Doe^John^A~Doe^Johnny^B||19800101|M|||123 Main St^^Metropolis^NY^10001")

var testSegDelim = byte('\r')

var validHL7 = bytes.Join([][]byte{validMSH, validPID, validPV1}, []byte{testSegDelim})
var validHL72 = []byte("MSH|^~\\&||GA0000||VAERS PROCESSOR|20010331605||ORU^R01|20010422GA03|T|2.3.1|||AL\rPID|||1234^^^^SR~1234-12^^^^LR~00725^^^^MR||Doe^John^Fitzgerald^JR^^^L||20001007|M||2106-3^White^HL70005|123 Peachtree St^APT 3B^Atlanta^GA^30210^^M^^GA067||(678) 555-1212^^PRN\rNK1|1|Jones^Jane^Lee^^RN|VAB^Vaccine administered by (Name)^HL70063\rNK1|2|Jones^Jane^Lee^^RN|FVP^Form completed by (Name)-Vaccine provider^HL70063|101 Main Street^^Atlanta^GA^38765^^O^^GA121||(404) 554-9097^^WPN\rORC|CN|||||||||||1234567^Welby^Marcus^J^Jr^Dr.^MD^L|||||||||Peachtree Clinic|101 Main Street^^Atlanta^GA^38765^^O^^GA121|(404) 554-9097^^WPN|101 Main Street^^Atlanta^GA^38765^^O^^GA121\rOBR|1|||^CDC VAERS-1 (FDA) Report|||20010316\rOBX|1|NM|21612-7^Reported Patient Age^LN||05|mo^month^ANSI\rOBX|1|TS|30947-6^Date form completed^LN||20010316\rOBX|2|FT|30948-4^Vaccination adverse events and treatment, if any^LN|1|fever of 106F, with vomiting, seizures, persistent crying lasting over 3 hours, loss of appetite\rOBX|3|CE|30949-2^Vaccination adverse event outcome^LN|1|E^required emergency room/doctor visit^NIP005\rOBX|4|CE|30949-2^Vaccination adverse event outcome^LN|1|H^required hospitalization^NIP005\rOBX|5|NM|30950-0^Number of days hospitalized due to vaccination adverse event^LN|1|02|d^day^ANSI\rOBX|6|CE|30951-8^Patient recovered^LN||Y^Yes^ HL70239\rOBX|7|TS|30952-6^Date of vaccination^LN||20010216\rOBX|8|TS|30953-4^Adverse event onset date and time^LN||200102180900\rOBX|9|FT|30954-2^Relevant diagnostic tests/lab data^LN||Electrolytes, CBC, Blood culture")
var invalidLineEnding = bytes.Join([][]byte{validMSH, validPID, validPV1}, []byte{'\t'})
var invalidMSH = bytes.Join([][]byte{[]byte("MSH|"), validPID, validPV1}, []byte("\r"))

var sampleFile = "test.hl7"

func TestNewMessage(t *testing.T) {
	got, err := NewMessage(validHL7)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	fmt.Println(string(got))
	// want := Message{
	// 	"MSH": map[string]interface{}{
	// 		"MSH.1": "|",
	// 		"MSH.2": "^~\\&",
	// 		"MSH.3": "LabSystem",
	// 		"MSH.4": "Hospital",
	// 		"MSH.5": "OrderingSystem",
	// 		"MSH.6": "Clinic",
	// 		"MSH.7": "202501140830",
	// 		"MSH.9": map[string]interface{}{
	// 			"MSH.9.1": "ORU",
	// 			"MSH.9.2": "R01",
	// 		},
	// 		"MSH.10": "MSG00002",
	// 		"MSH.11": "P",
	// 		"MSH.12": "2.3",
	// 	},
	// 	"PID": map[string]interface{}{
	// 		"PID.1": "1",
	// 		"PID.3": map[string]interface{}{
	// 			"PID.3.1": "123456",
	// 			"PID.3.4": "Hospital",
	// 			"PID.3.5": "MR",
	// 		},
	// 		"PID.5": []map[string]interface{}{
	// 			{
	// 				"PID.5.1": "Doe",
	// 				"PID.5.2": "John",
	// 				"PID.5.3": "A",
	// 			},
	// 			{
	// 				"PID.5.1": "Doe",
	// 				"PID.5.2": "Johnny",
	// 				"PID.5.3": "B",
	// 			},
	// 		},
	// 		"PID.7": "19800101",
	// 		"PID.8": "M",
	// 		"PID.11": map[string]interface{}{
	// 			"PID.11.1": "123 Main St",
	// 			"PID.11.3": "Metropolis",
	// 			"PID.11.4": "NY",
	// 			"PID.11.5": "10001",
	// 		},
	// 	},
	// 	"PV1": map[string]interface{}{
	// 		"PV1.1": "1",
	// 		"PV1.2": "I",
	// 		"PV1.3": map[string]interface{}{
	// 			"PV1.3.1": "ICU",
	// 			"PV1.3.2": "Room101",
	// 			"PV1.3.3": "BedA",
	// 			"PV1.3.5": "Hospital",
	// 		},
	// 		"PV1.7": map[string]interface{}{
	// 			"PV1.7.1": "1234",
	// 			"PV1.7.2": "Smith",
	// 			"PV1.7.3": "John",
	// 			"PV1.7.4": "A",
	// 			"PV1.7.7": "Dr.",
	// 		},
	// 		"PV1.10": "Cardiology",
	// 	},
	// }

	// if !reflect.DeepEqual(got, want) {
	// 	t.Errorf("got %v, want %v", got, want)
	// }
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
			if got, err := NewMessage(tt.msg); err == nil {
				t.Errorf("expected error, got nil\nresults: %v", got)
			}
		})
	}
}

func BenchmarkNewMessage(b *testing.B) {
	runtime.GOMAXPROCS(1)

	for i := 0; i < b.N; i++ {
		_, err := NewMessage(validHL72)
		if err != nil {
			b.Fatalf("unexpected error: %v", err)
		}
	}
}
