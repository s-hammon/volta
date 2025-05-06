package models

import (
	"testing"

	"github.com/s-hammon/volta/pkg/hl7"
)

func BenchmarkUnmarshalReportDecoderOne(b *testing.B) {
	data, err := hl7.HL7.ReadFile("test_hl7/9.hl7")
	if err != nil {
		b.Fatalf("failed to read test file: %v", err)
	}
	if len(data) == 0 {
		b.Fatalf("test file is empty!")
	}

	b.Run("9.hl7", func(b *testing.B) {
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
