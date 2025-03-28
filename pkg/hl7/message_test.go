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

var validHL7 = []byte("MSH|^~\\&||GA0000||VAERS PROCESSOR|20010331605||ORU^R01|20010422GA03|T|2.3.1|||AL\rPID|||1234^^^^SR~1234-12^^^^LR~00725^^^^MR||Doe^John^Fitzgerald^JR^^^L||20001007|M||2106-3^White^HL70005|123 Peachtree St^APT 3B^Atlanta^GA^30210^^M^^GA067||(678) 555-1212^^PRN\rNK1|1|Jones^Jane^Lee^^RN|VAB^Vaccine administered by (Name)^HL70063\rNK1|2|Jones^Jane^Lee^^RN|FVP^Form completed by (Name)-Vaccine provider^HL70063|101 Main Street^^Atlanta^GA^38765^^O^^GA121||(404) 554-9097^^WPN\rORC|CN|||||||||||1234567^Welby^Marcus^J^Jr^Dr.^MD^L|||||||||Peachtree Clinic|101 Main Street^^Atlanta^GA^38765^^O^^GA121|(404) 554-9097^^WPN|101 Main Street^^Atlanta^GA^38765^^O^^GA121\rOBR|1|||^CDC VAERS-1 (FDA) Report|||20010316\rOBX|1|NM|21612-7^Reported Patient Age^LN||05|mo^month^ANSI\rOBX|2|TS|30947-6^Date form completed^LN||20010316\rOBX|3|FT|30948-4^Vaccination adverse events and treatment, if any^LN|1|fever of 106F, with vomiting, seizures, persistent crying lasting over 3 hours, loss of appetite\rOBX|4|CE|30949-2^Vaccination adverse event outcome^LN|1|E^required emergency room/doctor visit^NIP005\rOBX|5|CE|30949-2^Vaccination adverse event outcome^LN|1|H^required hospitalization^NIP005\rOBX|6|NM|30950-0^Number of days hospitalized due to vaccination adverse event^LN|1|02|d^day^ANSI\rOBX|7|CE|30951-8^Patient recovered^LN||Y^Yes^ HL70239\rOBX|8|TS|30952-6^Date of vaccination^LN||20010216\rOBX|9|TS|30953-4^Adverse event onset date and time^LN||200102180900\rOBX|10|FT|30954-2^Relevant diagnostic tests/lab data^LN||Electrolytes, CBC, Blood culture")
var invalidLineEnding = bytes.Join([][]byte{validMSH, validPID, validPV1}, []byte{'\t'})
var invalidMSH = bytes.Join([][]byte{[]byte("MSH|"), validPID, validPV1}, []byte("\r"))

func TestNewMessage(t *testing.T) {
	got, err := NewMessage(validHL7)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	want := Message(`{"MSH":{"MSH.1":"|","MSH.2":"^~\\&","MSH.4":"GA0000","MSH.6":"VAERS PROCESSOR","MSH.7":"20010331605","MSH.9":{"MSH.9.1":"ORU","MSH.9.2":"R01"},"MSH.10":"20010422GA03","MSH.11":"T","MSH.12":"2.3.1","MSH.15":"AL"},"PID":{"PID.3":[{"PID.3.1":"1234","PID.3.5":"SR"},{"PID.3.1":"1234-12","PID.3.5":"LR"},{"PID.3.1":"00725","PID.3.5":"MR"}],"PID.5":{"PID.5.1":"Doe","PID.5.2":"John","PID.5.3":"Fitzgerald","PID.5.4":"JR","PID.5.7":"L"},"PID.7":"20001007","PID.8":"M","PID.10":{"PID.10.1":"2106-3","PID.10.2":"White","PID.10.3":"HL70005"},"PID.11":{"PID.11.1":"123 Peachtree St","PID.11.2":"APT 3B","PID.11.3":"Atlanta","PID.11.4":"GA","PID.11.5":"30210","PID.11.7":"M","PID.11.9":"GA067"},"PID.13":{"PID.13.1":"(678) 555-1212","PID.13.3":"PRN"}},"NK1":[{"NK1.1":"1","NK1.2":{"NK1.2.1":"Jones","NK1.2.2":"Jane","NK1.2.3":"Lee","NK1.2.5":"RN"},"NK1.3":{"NK1.3.1":"VAB","NK1.3.2":"Vaccine administered by (Name)","NK1.3.3":"HL70063"}},{"NK1.1":"2","NK1.2":{"NK1.2.1":"Jones","NK1.2.2":"Jane","NK1.2.3":"Lee","NK1.2.5":"RN"},"NK1.3":{"NK1.3.1":"FVP","NK1.3.2":"Form completed by (Name)-Vaccine provider","NK1.3.3":"HL70063"},"NK1.4":{"NK1.4.1":"101 Main Street","NK1.4.3":"Atlanta","NK1.4.4":"GA","NK1.4.5":"38765","NK1.4.7":"O","NK1.4.9":"GA121"},"NK1.6":{"NK1.6.1":"(404) 554-9097","NK1.6.3":"WPN"}}],"ORC":{"ORC.1":"CN","ORC.12":{"ORC.12.1":"1234567","ORC.12.2":"Welby","ORC.12.3":"Marcus","ORC.12.4":"J","ORC.12.5":"Jr","ORC.12.6":"Dr.","ORC.12.7":"MD","ORC.12.8":"L"},"ORC.21":"Peachtree Clinic","ORC.22":{"ORC.22.1":"101 Main Street","ORC.22.3":"Atlanta","ORC.22.4":"GA","ORC.22.5":"38765","ORC.22.7":"O","ORC.22.9":"GA121"},"ORC.23":{"ORC.23.1":"(404) 554-9097","ORC.23.3":"WPN"},"ORC.24":{"ORC.24.1":"101 Main Street","ORC.24.3":"Atlanta","ORC.24.4":"GA","ORC.24.5":"38765","ORC.24.7":"O","ORC.24.9":"GA121"}},"OBR":{"OBR.1":"1","OBR.4":{"OBR.4.2":"CDC VAERS-1 (FDA) Report"},"OBR.7":"20010316"},"OBX":[{"OBX.1":"1","OBX.2":"NM","OBX.3":{"OBX.3.1":"21612-7","OBX.3.2":"Reported Patient Age","OBX.3.3":"LN"},"OBX.5":"05","OBX.6":{"OBX.6.1":"mo","OBX.6.2":"month","OBX.6.3":"ANSI"}},{"OBX.1":"2","OBX.2":"TS","OBX.3":{"OBX.3.1":"30947-6","OBX.3.2":"Date form completed","OBX.3.3":"LN"},"OBX.5":"20010316"},{"OBX.1":"3","OBX.2":"FT","OBX.3":{"OBX.3.1":"30948-4","OBX.3.2":"Vaccination adverse events and treatment, if any","OBX.3.3":"LN"},"OBX.4":"1","OBX.5":"fever of 106F, with vomiting, seizures, persistent crying lasting over 3 hours, loss of appetite"},{"OBX.1":"4","OBX.2":"CE","OBX.3":{"OBX.3.1":"30949-2","OBX.3.2":"Vaccination adverse event outcome","OBX.3.3":"LN"},"OBX.4":"1","OBX.5":{"OBX.5.1":"E","OBX.5.2":"required emergency room/doctor visit","OBX.5.3":"NIP005"}},{"OBX.1":"5","OBX.2":"CE","OBX.3":{"OBX.3.1":"30949-2","OBX.3.2":"Vaccination adverse event outcome","OBX.3.3":"LN"},"OBX.4":"1","OBX.5":{"OBX.5.1":"H","OBX.5.2":"required hospitalization","OBX.5.3":"NIP005"}},{"OBX.1":"6","OBX.2":"NM","OBX.3":{"OBX.3.1":"30950-0","OBX.3.2":"Number of days hospitalized due to vaccination adverse event","OBX.3.3":"LN"},"OBX.4":"1","OBX.5":"02","OBX.6":{"OBX.6.1":"d","OBX.6.2":"day","OBX.6.3":"ANSI"}},{"OBX.1":"7","OBX.2":"CE","OBX.3":{"OBX.3.1":"30951-8","OBX.3.2":"Patient recovered","OBX.3.3":"LN"},"OBX.5":{"OBX.5.1":"Y","OBX.5.2":"Yes","OBX.5.3":" HL70239"}},{"OBX.1":"8","OBX.2":"TS","OBX.3":{"OBX.3.1":"30952-6","OBX.3.2":"Date of vaccination","OBX.3.3":"LN"},"OBX.5":"20010216"},{"OBX.1":"9","OBX.2":"TS","OBX.3":{"OBX.3.1":"30953-4","OBX.3.2":"Adverse event onset date and time","OBX.3.3":"LN"},"OBX.5":"200102180900"},{"OBX.1":"10","OBX.2":"FT","OBX.3":{"OBX.3.1":"30954-2","OBX.3.2":"Relevant diagnostic tests/lab data","OBX.3.3":"LN"},"OBX.5":"Electrolytes, CBC, Blood culture"}]}`)
	if !reflect.DeepEqual(got, want) {
		t.Errorf("got %s, want %s", got, want)
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
			if got, err := NewMessage(tt.msg); err == nil {
				t.Errorf("expected error, got nil\nresults: %v", got)
			}
		})
	}
}

func BenchmarkNewMessage(b *testing.B) {
	runtime.GOMAXPROCS(1)

	for i := 0; i < b.N; i++ {
		_, err := NewMessage(validHL7)
		if err != nil {
			b.Fatalf("unexpected error: %v", err)
		}
	}
}
