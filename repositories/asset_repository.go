package repositories

import (
	"database/sql"
	"math"
	"strconv"
	"strings"
)

type AssetRepository struct {
	DB *sql.DB
}

func assetNullableInt64(value int64) interface{} {
	if value <= 0 {
		return nil
	}
	return value
}

func nullInt64Value(value sql.NullInt64) interface{} {
	if !value.Valid {
		return nil
	}
	return value.Int64
}

func nullStringValue(value sql.NullString) interface{} {
	if !value.Valid || strings.TrimSpace(value.String) == "" {
		return nil
	}
	return value.String
}

func boolToInt(value bool) int {
	if value {
		return 1
	}
	return 0
}

func activeLabel(value bool) string {
	if value {
		return "Aktif"
	}
	return "Non Aktif"
}

func formatNullTime(value sql.NullTime) string {
	if value.Valid {
		return value.Time.Format("02 Jan 2006 15:04")
	}
	return "-"
}

func formatAssetAmountID(value float64) string {
	rounded := int64(math.Round(value))
	raw := strconv.FormatInt(rounded, 10)
	var parts []string
	for len(raw) > 3 {
		parts = append([]string{raw[len(raw)-3:]}, parts...)
		raw = raw[:len(raw)-3]
	}
	if raw != "" {
		parts = append([]string{raw}, parts...)
	}
	return "IDR " + strings.Join(parts, ",")
}

func formatNumberInput(value float64) string {
	return strconv.FormatFloat(value, 'f', -1, 64)
}

func formatAssetStatusLabel(status string) string {
	switch status {
	case "AVAILABLE":
		return "Available"
	case "IN_USE":
		return "In Use"
	case "MAINTENANCE":
		return "Maintenance"
	case "BROKEN":
		return "Broken"
	case "DISPOSED":
		return "Disposed"
	default:
		return status
	}
}

func formatComponentStatusLabel(status string) string {
	switch status {
	case "IN_STORAGE":
		return "In Storage"
	case "INSTALLED":
		return "Installed"
	case "MAINTENANCE":
		return "Maintenance"
	case "BROKEN":
		return "Broken"
	case "DISPOSED":
		return "Disposed"
	default:
		return status
	}
}
