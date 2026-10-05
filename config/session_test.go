package config

import "testing"

func TestSessionSecretRequiresStrongProductionValue(t *testing.T) {
	t.Setenv("APP_ENV", "production")
	t.Setenv("SESSION_SECRET", "short")
	if _, err := SessionSecret(); err == nil {
		t.Fatal("expected production validation error for short SESSION_SECRET")
	}
}

func TestSessionSecretUsesConfiguredValue(t *testing.T) {
	t.Setenv("APP_ENV", "production")
	want := "a-production-session-secret-that-is-long-enough"
	t.Setenv("SESSION_SECRET", want)
	got, err := SessionSecret()
	if err != nil {
		t.Fatalf("SessionSecret returned error: %v", err)
	}
	if string(got) != want {
		t.Fatalf("SessionSecret = %q, want configured value", string(got))
	}
}
