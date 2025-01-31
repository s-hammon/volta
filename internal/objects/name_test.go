package objects

import "testing"

func TestName(t *testing.T) {
	name := NewName(
		"Smith",
		"John",
		"Q",
		"Jr",
		"Dr",
		"MD",
	)

	if name.Full() != "Dr John Q Smith Jr MD" {
		t.Errorf("Full() = %s; want Dr John Q Smith Jr MD", name.Full())
	}
	if name.Record() != "Smith, John Q" {
		t.Errorf("Record() = %s; want Smith, John Q", name.Record())
	}

	name.Coalesce(NewName("", "", "", "", "", ""))
	if name.Full() != "Dr John Q Smith Jr MD" {
		t.Errorf("Full() = %s; want Dr John Q Smith Jr MD", name.Full())
	}

	name.Coalesce(NewName("", "Adam", "", "", "", ""))
	if name.Full() != "Dr Adam Q Smith Jr MD" {
		t.Errorf("Full() = %s; want Dr Adam Q Smith Jr MD", name.Full())
	}
}
