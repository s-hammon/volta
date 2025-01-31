package models

import (
	"encoding/json"
	"reflect"
	"testing"
)

func TestMessageModelUnmarshal(t *testing.T) {
	msgMap := map[string]interface{}{
		"MSH.1": "field separator",
		"MSH.2": "encoding chars",
		"MSH.3": "sending app",
		"MSH.4": "sending fac",
		"MSH.5": "receiving app",
		"MSH.6": "receiving fac",
		"MSH.7": "20210101000000",
		"MSH.9": map[string]interface{}{
			"MSH.9.1": "ADT",
			"MSH.9.2": "A01",
		},
		"MSH.10": "control id",
		"MSH.11": "processing id",
		"MSH.12": "version",
	}
	b, err := json.Marshal(msgMap)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	want := MessageModel{
		FieldSeparator: "field separator",
		EncodingChars:  "encoding chars",
		SendingApp:     "sending app",
		SendingFac:     "sending fac",
		ReceivingApp:   "receiving app",
		ReceivingFac:   "receiving fac",
		DateTime:       "20210101000000",
		Type: CM_MSG{
			Type:         "ADT",
			TriggerEvent: "A01",
		},
		ControlID:    "control id",
		ProcessingID: "processing id",
		Version:      "version",
	}

	got := MessageModel{}
	if err = json.Unmarshal(b, &got); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !reflect.DeepEqual(got, want) {
		t.Errorf("got '%v', want '%v'", got, want)
	}
}
