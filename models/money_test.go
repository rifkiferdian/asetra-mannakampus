package models

import (
	"encoding/json"
	"testing"
)

func TestParseMoneyIsExact(t *testing.T) {
	tests := map[string]Money{
		"0":            0,
		"0.01":         1,
		"1,250,000.50": 125000050,
		"-10.25":       -1025,
	}
	for input, want := range tests {
		got, err := ParseMoney(input)
		if err != nil {
			t.Fatalf("ParseMoney(%q): %v", input, err)
		}
		if got != want {
			t.Fatalf("ParseMoney(%q) = %d, want %d", input, got, want)
		}
	}
}

func TestParseMoneyRejectsExcessPrecision(t *testing.T) {
	if _, err := ParseMoney("10.001"); err == nil {
		t.Fatal("expected precision validation error")
	}
}

func TestParseMoneyRejectsValueBeyondDatabaseColumn(t *testing.T) {
	if _, err := ParseMoney("1000000000000000.00"); err == nil {
		t.Fatal("expected DECIMAL(18,2) range validation error")
	}
}

func TestMultiplyMoneyByQuantityRoundsToDatabaseScale(t *testing.T) {
	got, err := MultiplyMoneyByQuantity(MustMoney("10.05"), 1.25)
	if err != nil {
		t.Fatal(err)
	}
	if want := MustMoney("12.56"); got != want {
		t.Fatalf("total = %s, want %s", got.DecimalString(), want.DecimalString())
	}
}

func TestMoneyJSONUsesExactDecimalString(t *testing.T) {
	encoded, err := json.Marshal(MustMoney("123.45"))
	if err != nil {
		t.Fatal(err)
	}
	if string(encoded) != `"123.45"` {
		t.Fatalf("json = %s", encoded)
	}
}

func TestExceedsVariance(t *testing.T) {
	base := MustMoney("100.00")
	if ExceedsVariance(MustMoney("110.00"), base, 1_000) {
		t.Fatal("amount exactly at 10% tolerance must be accepted")
	}
	if !ExceedsVariance(MustMoney("110.01"), base, 1_000) {
		t.Fatal("amount above 10% tolerance must be rejected")
	}
}
