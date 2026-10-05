package config

import (
	"gobase-app/models"
	"log"
	"os"
	"strings"
)

const defaultPOVarianceToleranceBPS int64 = 1_000 // 10.00%

func POVarianceToleranceBPS() int64 {
	raw := strings.TrimSpace(os.Getenv("PO_VARIANCE_TOLERANCE_PERCENT"))
	if raw == "" {
		return defaultPOVarianceToleranceBPS
	}
	percentage, err := models.ParseMoney(raw)
	bps := int64(percentage)
	if err != nil || bps < 0 || bps > 10_000 {
		log.Printf("WARNING: invalid PO_VARIANCE_TOLERANCE_PERCENT %q; using 10.00%%", raw)
		return defaultPOVarianceToleranceBPS
	}
	return bps
}
