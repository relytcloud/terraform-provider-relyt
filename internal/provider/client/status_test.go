package client

import "testing"

func TestIsProvisionFailed(t *testing.T) {
	tests := []struct {
		status string
		want   bool
	}{
		// Both spellings occur: the state machine writes one, the API returns
		// the other. Matching only one lets failures slip through as "pending".
		{STATUS_PROVISION_FAILED, true},
		{STATUS_PROVISIONING_FAILED, true},
		{DPS_STATUS_READY, false},
		{"PROVISIONING", false},
		// The delete paths wait for DROPPED, so it must not read as a failure.
		{DPS_STATUS_DROPPED, false},
		{"", false},
	}
	for _, tt := range tests {
		if got := IsProvisionFailed(tt.status); got != tt.want {
			t.Errorf("IsProvisionFailed(%q) = %v, want %v", tt.status, got, tt.want)
		}
	}
}
