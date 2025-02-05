package api

import (
	"bytes"
	"reflect"
	"testing"
)

var mockPubSubMessage = []byte(`{"Message": {"Data": "cHJvamVjdHMvUFJPSkVDVF9JRC9sb2NhdGlvbnMvTE9DQVRJT05fSUQvZGF0YXNldHMvREFUQVNFVF9JRC9obDdWMlN0b3Jlcy9ITDdWMlNUT1JFX0lEL21lc3NhZ2VzL01FU1NBR0VfSUQK", "Attributes": {"Type": "ORM"}}, "Subscription": "test"}`)

func TestNewPubSubMessage(t *testing.T) {
	buf := bytes.NewReader(mockPubSubMessage)
	want := &pubSubMessage{
		Message: message{
			Data: []byte("projects/PROJECT_ID/locations/LOCATION_ID/datasets/DATASET_ID/hl7V2Stores/HL7V2STORE_ID/messages/MESSAGE_ID\n"),
			Attributes: attributes{
				Type: "ORM",
			},
		},
		Subscription: "test",
	}
	got, err := NewPubSubMessage(buf)
	if err != nil {
		t.Fatalf("NewPubSubMessage() unexpected error = %v", err)
	}

	if !reflect.DeepEqual(got, want) {
		t.Errorf("NewPubSubMessage() = got '%s', want '%s'", got, want)
	}
}

func BenchmarkNewPubSubMessage(b *testing.B) {
	r := bytes.NewReader(mockPubSubMessage)
	for i := 0; i < b.N; i++ {
		r.Reset(mockPubSubMessage)
		_, err := NewPubSubMessage(r)
		if err != nil {
			b.Fatalf("NewPubSubMessage() error = %v", err)
		}
	}
}
