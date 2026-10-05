package services

import (
	"gobase-app/models"
	"testing"
)

func TestCalculatePRTotalRoundsEachLineToCurrencyScale(t *testing.T) {
	items := []models.PurchaseRequestItemInput{
		{Qty: 1.25, EstUnitPrice: models.MustMoney("10.05")},
		{Qty: 3, EstUnitPrice: models.MustMoney("0.10")},
	}

	got, err := calculatePRTotal(items)
	if err != nil {
		t.Fatal(err)
	}
	if want := models.MustMoney("12.86"); got != want {
		t.Fatalf("total = %s, want %s", got.DecimalString(), want.DecimalString())
	}
}

func TestCalculatePRTotalRejectsQuantityBeyondDatabaseScale(t *testing.T) {
	_, err := calculatePRTotal([]models.PurchaseRequestItemInput{
		{Qty: 1.234, EstUnitPrice: models.MustMoney("10.00")},
	})
	if err == nil {
		t.Fatal("expected quantity precision validation error")
	}
}
