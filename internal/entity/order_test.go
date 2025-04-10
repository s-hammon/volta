package entity

import "testing"

func TestOrderStatus(t *testing.T) {
	tests := []struct {
		name    string
		order   string
		want    orderStatus
		wantErr bool
	}{
		{"scheduled", "SC", OrderScheduled, false},
		{"in progress", "IP", OrderInProgress, false},
		{"complete", "CM", OrderComplete, false},
		{"cancelled", "CA", OrderCancelled, false},
		{"rescheduled", "RS", OrderRescheduled, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := newOrderStatus(tt.order)
			if got != tt.want {
				t.Errorf("want: '%s', got: '%s'", tt.want, got)
			}
		})
	}
}
