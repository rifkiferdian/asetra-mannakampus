package helpers

import (
	"fmt"
	"strings"
	"time"
)

var monthsID = map[time.Month]string{
	time.January:   "Januari",
	time.February:  "Februari",
	time.March:     "Maret",
	time.April:     "April",
	time.May:       "Mei",
	time.June:      "Juni",
	time.July:      "Juli",
	time.August:    "Agustus",
	time.September: "September",
	time.October:   "Oktober",
	time.November:  "November",
	time.December:  "Desember",
}

// DateNowID mengembalikan tanggal sekarang format: 02 Desember 2025
func DateNowID() string {
	now := time.Now()
	return fmt.Sprintf("%02d %s %d",
		now.Day(),
		monthsID[now.Month()],
		now.Year(),
	)
}

// FormatDateID untuk format waktu tertentu
func FormatDateID(t time.Time) string {
	return fmt.Sprintf("%02d %s %d",
		t.Day(),
		monthsID[t.Month()],
		t.Year(),
	)
}

func Initials(s string) string {
	words := strings.Fields(s)
	var result string

	for _, w := range words {
		r := []rune(w)
		if len(r) > 0 {
			result += string(r[0])
		}
	}

	return result
}
