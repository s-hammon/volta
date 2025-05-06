package models

import (
	"testing"

	"github.com/s-hammon/volta/pkg/hl7"
)

var fileName = "test_hl7/9.hl7"

func BenchmarkUnmarshalDecoderAll(b *testing.B) {
	data, err := hl7.HL7.ReadFile(fileName)
	if err != nil {
		b.Fatalf("failed to read test file: %v", err)
	}
	if len(data) == 0 {
		b.Fatalf("test file is empty!")
	}
	b.Run(fileName, func(b *testing.B) {
		d := hl7.NewDecoder(data)
		b.ReportAllocs()
		b.ResetTimer()
		for range b.N {
			patient := &PatientModel{}
			if err = d.Decode(patient); err != nil {
				b.Fatal(err)
			}
			visit := &VisitModel{}
			if err = d.Decode(visit); err != nil {
				b.Fatal(err)
			}
			exam := []ExamModel{}
			if err = d.Decode(&exam); err != nil {
				b.Fatal(err)
			}
			report := []ReportModel{}
			if err = d.Decode(&report); err != nil {
				b.Fatal(err)
			}
		}
	})
}

func BenchmarkUnmarshalExamDecoderOne(b *testing.B) {
	data, err := hl7.HL7.ReadFile(fileName)
	if err != nil {
		b.Fatalf("failed to read test file: %v", err)
	}
	if len(data) == 0 {
		b.Fatalf("test file is empty!")
	}

	b.Run(fileName, func(b *testing.B) {
		b.ReportAllocs()
		b.ResetTimer()
		for range b.N {
			d := hl7.NewDecoder(data)
			exam := &ExamModel{}
			if err = d.Decode(exam); err != nil {
				b.Fatalf("couldn't unmarshal: %v", err)
			}
		}
	})
}
