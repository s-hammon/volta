package hl7

import (
	"testing"

	"github.com/stretchr/testify/require"
)

var wholeShabang = []byte("MSH|^~\\&|LabSystem|Hospital|OrderingSystem|Clinic|202501140830||ORU^R01|MSG00002|P|2.3\rPID|1||123456^^^Hospital^MR||Doe^John^A~Doe^Johnny^B||19800101|M|||123 Main St^^Metropolis^NY^10001\rPV1|1|I|ICU&Room101^Hospital&BedA||||1234^Smith^John^A^^^Dr.|||Cardiology")

type mockModel struct {
	FieldDelimiter    string   `hl7:"MSH.1"`
	EcodingCharacters string   `hl7:"MSH.2"`
	SendingFacility   string   `hl7:"MSH.3"`
	MessageType       cm_msg   `hl7:"MSH.9"`
	Quantity          []string `hl7:"OBR.9"`
	Procedure         ce       `hl7:"OBX.3"`
}

type mockObservation struct {
	LineNo    string `hl7:"OBX.1"`
	Procedure ce     `hl7:"OBX.3"`
	Results   string `hl7:"OBX.5"`
}

type mockPatient struct {
	MRN      ce     `hl7:"PID.3"`
	Name     []xpn  `hl7:"PID.5"`
	DOB      string `hl7:"PID.7"`
	Location listPL `hl7:"PV1.3"`
}

type mockVisit struct {
	Class            string `hl7:"PV1.2"`
	AssignedLocation listPL `hl7:"PV1.3"`
}

type cm_msg struct {
	Name         string `hl7:"1"`
	TriggerEvent string `hl7:"2"`
}

type ce struct {
	Code               string `hl7:"1"`
	Description        string `hl7:"2"`
	AssigningAuthority string `hl7:"3"`
	IdentifierTypeCode string `hl7:"4"`
	AssigningFacility  string `hl7:"5"`
}

type cx struct {
	ID         string `hl7:"1"`
	CheckDigit string `hl7:"2"`
}

type listPL struct {
	FirstLocation  pl `hl7:"1"`
	SecondLocation pl `hl7:"2"`
}

type pl struct {
	Unit string `hl7:"1"`
	Room string `hl7:"2"`
}

type xpn struct {
	Last   string `hl7:"1"`
	First  string `hl7:"2"`
	Middle string `hl7:"3"`
}

type xcn struct {
	ID     string `hl7:"1"`
	Last   string `hl7:"1"`
	First  string `hl7:"2"`
	Middle string `hl7:"3"`
	Degree string `hl7:"7"`
}

var multipleOrders = []byte("MSH|^~\\&|LabSystem|Hospital|OrderingSystem|Clinic|202501140830||ORU^R01|MSG00002|P|2.3\rPID|1||123456^^^Hospital^MR||Doe^John^A~Doe^Johnny^B||19800101|M|||123 Main St^^Metropolis^NY^10001\rPV1|1|I|ICU&Room101^Hospital&BedA||||1234^Smith^John^A^^^Dr.|||Cardiology\rORC|CN|42069|96024||CM||20250115083500||20250115083500\rOBR|1|42069|96024|CXR^Chest X-Ray|S\rORC|RE|42070|07024||CM||20250115083500||20250115083500\rOBR|2|42070|07024|UDOP^US Doppler|S")

type orderGroup struct {
	Control   string `hl7:"ORC.1"`
	PlacerNo  string `hl7:"ORC.2"`
	FillerNo  string `hl7:"ORC.3"`
	Procedure ce     `hl7:"OBR.4"`
	Priority  string `hl7:"OBR.5"`
}

func TestUnmarshal(t *testing.T) {
	pid := &mockPatient{}
	err := Unmarshal(wholeShabang, pid)
	require.NoError(t, err)
	want := &mockPatient{
		MRN: ce{
			Code:               "123456",
			Description:        "",
			AssigningAuthority: "",
			IdentifierTypeCode: "Hospital",
			AssigningFacility:  "MR",
		},
		Name: []xpn{
			{"Doe", "John", "A"},
			{"Doe", "Johnny", "B"},
		},
		DOB: "19800101",
		Location: listPL{
			FirstLocation:  pl{"ICU", "Room101"},
			SecondLocation: pl{"Hospital", "BedA"},
		},
	}
	require.Equal(t, want, pid)

	obs := []mockObservation{}
	err = Unmarshal(validOBX, &obs)
	require.NoError(t, err)
	require.Equal(t, 2, len(obs))
	wantObs := []mockObservation{
		{"1", ce{Code: "CXR", Description: "Chest X-ray"}, "diagnostic"},
		{"2", ce{Code: "CXR", Description: "Chest X-ray"}, "more diagnostic"},
	}
	require.Equal(t, wantObs, obs)

	orders := []orderGroup{}
	err = Unmarshal(multipleOrders, &orders)
	require.NoError(t, err)
	require.Equal(t, 2, len(orders))
	wantOrders := []orderGroup{
		{
			"CN",
			"42069",
			"96024",
			ce{Code: "CXR", Description: "Chest X-Ray"},
			"S",
		},
		{
			"RE",
			"42070",
			"07024",
			ce{Code: "UDOP", Description: "US Doppler"},
			"S",
		},
	}
	require.Equal(t, wantOrders, orders)
}
