package config

import "testing"

func TestPOVarianceToleranceBPS(t *testing.T) {
	t.Setenv("PO_VARIANCE_TOLERANCE_PERCENT", "7.25")
	if got := POVarianceToleranceBPS(); got != 725 {
		t.Fatalf("tolerance = %d bps, want 725", got)
	}
}

func TestPOVarianceToleranceDefaultsForInvalidValue(t *testing.T) {
	t.Setenv("PO_VARIANCE_TOLERANCE_PERCENT", "invalid")
	if got := POVarianceToleranceBPS(); got != 1_000 {
		t.Fatalf("tolerance = %d bps, want default 1000", got)
	}
}
