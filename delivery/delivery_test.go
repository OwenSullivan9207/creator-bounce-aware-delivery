package delivery

import "testing"

func TestDecideSuppressionAndBounce(t *testing.T) {
	tests := []struct {
		name               string
		bounce, suppressed bool
		status, reason     string
	}{
		{"hard bounce wins", true, false, "blocked", "hard bounce"},
		{"suppression blocks", false, true, "blocked", "suppressed subscriber"},
		{"clean subscriber", false, false, "ready", ""},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got := Decide(tc.bounce, tc.suppressed)
			if got.Status != tc.status || got.Reason != tc.reason {
				t.Fatalf("got %#v, want status=%q reason=%q", got, tc.status, tc.reason)
			}
		})
	}
}
