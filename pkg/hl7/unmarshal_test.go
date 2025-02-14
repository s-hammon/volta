package hl7

// import (
// 	"reflect"
// 	"runtime"
// 	"testing"
// )

// var mockHL7 = Message{
// 	"EVN.1": "A01",
// 	"EVN.2": "20210101",
// 	"EVN.5": map[string]interface{}{
// 		"1": "a",
// 		"2": "b",
// 		"3": "c",
// 	},
// 	"NAF.1": []string{"d", "e", "f"},
// 	"INS.1": []map[string]interface{}{
// 		{"1": "g", "2": "h", "3": "i"},
// 		{"1": "j", "2": "k", "3": "l"},
// 	},
// }

// type opID struct {
// 	IDNumber string `hl7:"1"`
// 	Family   string `hl7:"2"`
// 	Given    string `hl7:"3"`
// }

// type opIgnoreID struct {
// 	IDNumber string `hl7:"1"`
// 	Family   string `hl7:"-"`
// 	Given    string `hl7:"3"`
// }

// type ins struct {
// 	IDNumber string `hl7:"1"`
// 	Family   string `hl7:"2"`
// 	Given    string `hl7:"3"`
// }

// type mockStruct struct {
// 	Code       string   `hl7:"EVN.1"`
// 	DT         string   `hl7:"EVN.2"`
// 	OperatorID opID     `hl7:"EVN.5"`
// 	NotAField  []string `hl7:"NAF.1"`
// 	Insured    []ins    `hl7:"INS.1"`
// }

// type mockIgnoreTag struct {
// 	Code       string     `hl7:"EVN.1"`
// 	DT         string     `hl7:"EVN.2"`
// 	OperatorID opIgnoreID `hl7:"EVN.5"`
// }

// func TestUnmarshal(t *testing.T) {
// 	tests := []struct {
// 		name    string
// 		v       interface{}
// 		want    interface{}
// 		wantErr bool
// 	}{
// 		{
// 			name: "simple",
// 			v:    &mockStruct{},
// 			want: &mockStruct{
// 				Code: "A01",
// 				DT:   "20210101",
// 				OperatorID: opID{
// 					IDNumber: "a",
// 					Family:   "b",
// 					Given:    "c",
// 				},
// 				NotAField: []string{"d", "e", "f"},
// 				Insured: []ins{
// 					{IDNumber: "g", Family: "h", Given: "i"},
// 					{IDNumber: "j", Family: "k", Given: "l"},
// 				},
// 			},
// 		},
// 		{
// 			name: "ignore tag",
// 			v:    &mockIgnoreTag{},
// 			want: &mockIgnoreTag{
// 				Code: "A01",
// 				DT:   "20210101",
// 				OperatorID: opIgnoreID{
// 					IDNumber: "a",
// 					Given:    "c",
// 				},
// 			},
// 		},
// 		{
// 			name:    "v is not a pointer",
// 			v:       struct{}{},
// 			wantErr: true,
// 		},
// 		{
// 			name:    "v is not a struct pointer",
// 			v:       new(int),
// 			wantErr: true,
// 		},
// 		{
// 			name:    "v is nil",
// 			v:       nil,
// 			wantErr: true,
// 		},
// 	}

// 	for _, tt := range tests {
// 		t.Run(tt.name, func(t *testing.T) {
// 			if err := Unmarshal(mockHL7, tt.v); err != nil {
// 				if tt.wantErr {
// 					return
// 				}
// 				t.Errorf("Unmarshal() error = %v", err)
// 				return
// 			}
// 			if !reflect.DeepEqual(tt.v, tt.want) {
// 				t.Errorf("Unmarshal() = got '%v', want '%v'", tt.v, tt.want)
// 			}
// 		})
// 	}
// }

// func BenchmarkUnmarshal(b *testing.B) {
// 	runtime.GOMAXPROCS(1)

// 	for i := 0; i < b.N; i++ {
// 		var result mockStruct
// 		if err := Unmarshal(mockHL7, &result); err != nil {
// 			b.Fatalf("Unmarshal() error = %v", err)
// 		}
// 	}
// }
