package models

import (
	"database/sql/driver"
	"encoding/json"
	"errors"
	"fmt"
	"math"
	"math/big"
	"strconv"
	"strings"
)

const (
	moneyScale    int64 = 100
	moneyMaxCents int64 = 99_999_999_999_999_999 // DECIMAL(18,2)
)

// Money stores a DECIMAL(18,2) value as integer cents. It avoids binary
// floating-point arithmetic in purchasing and budget decisions.
type Money int64

type NullMoney struct {
	Money Money
	Valid bool
}

func (n *NullMoney) Scan(value interface{}) error {
	if value == nil {
		n.Money = 0
		n.Valid = false
		return nil
	}
	if err := n.Money.Scan(value); err != nil {
		return err
	}
	n.Valid = true
	return nil
}

func ParseMoney(value string) (Money, error) {
	normalized := strings.TrimSpace(value)
	normalized = strings.TrimSpace(strings.TrimPrefix(strings.TrimPrefix(normalized, "Rp"), "IDR"))
	normalized = strings.ReplaceAll(normalized, ",", "")
	if normalized == "" {
		return 0, errors.New("nilai uang wajib diisi")
	}

	sign := int64(1)
	if normalized[0] == '-' || normalized[0] == '+' {
		if normalized[0] == '-' {
			sign = -1
		}
		normalized = normalized[1:]
	}
	if normalized == "" {
		return 0, errors.New("nilai uang tidak valid")
	}

	parts := strings.Split(normalized, ".")
	if len(parts) > 2 || parts[0] == "" {
		return 0, errors.New("nilai uang tidak valid")
	}
	whole, err := strconv.ParseInt(parts[0], 10, 64)
	if err != nil || whole < 0 || whole > math.MaxInt64/moneyScale {
		return 0, errors.New("nilai uang di luar batas")
	}

	fraction := int64(0)
	if len(parts) == 2 {
		if len(parts[1]) > 2 {
			return 0, errors.New("nilai uang maksimal dua angka desimal")
		}
		fractionText := parts[1] + strings.Repeat("0", 2-len(parts[1]))
		if fractionText != "" {
			fraction, err = strconv.ParseInt(fractionText, 10, 64)
			if err != nil {
				return 0, errors.New("nilai uang tidak valid")
			}
		}
	}
	if whole == math.MaxInt64/moneyScale && fraction > math.MaxInt64-whole*moneyScale {
		return 0, errors.New("nilai uang di luar batas")
	}
	cents := whole*moneyScale + fraction
	if cents > moneyMaxCents {
		return 0, errors.New("nilai uang di luar batas DECIMAL(18,2)")
	}
	return Money(sign * cents), nil
}

func MustMoney(value string) Money {
	amount, err := ParseMoney(value)
	if err != nil {
		panic(err)
	}
	return amount
}

func (m *Money) Scan(value interface{}) error {
	if value == nil {
		*m = 0
		return nil
	}
	var raw string
	switch typed := value.(type) {
	case []byte:
		raw = string(typed)
	case string:
		raw = typed
	case int64:
		raw = strconv.FormatInt(typed, 10)
	case float64:
		raw = strconv.FormatFloat(typed, 'f', 2, 64)
	default:
		return fmt.Errorf("unsupported money scan type %T", value)
	}
	parsed, err := ParseMoney(raw)
	if err != nil {
		return err
	}
	*m = parsed
	return nil
}

func (m Money) Value() (driver.Value, error) {
	return m.DecimalString(), nil
}

func (m Money) MarshalJSON() ([]byte, error) {
	return json.Marshal(m.DecimalString())
}

func (m Money) DecimalString() string {
	value := int64(m)
	sign := ""
	if value < 0 {
		sign = "-"
		value = -value
	}
	return fmt.Sprintf("%s%d.%02d", sign, value/moneyScale, value%moneyScale)
}

func (m Money) FormatIDR() string {
	value := int64(m)
	sign := ""
	if value < 0 {
		sign = "-"
		value = -value
	}
	whole := strconv.FormatInt(value/moneyScale, 10)
	for index := len(whole) - 3; index > 0; index -= 3 {
		whole = whole[:index] + "." + whole[index:]
	}
	if fraction := value % moneyScale; fraction != 0 {
		return fmt.Sprintf("%sRp %s,%02d", sign, whole, fraction)
	}
	return sign + "Rp " + whole
}

func (m Money) Add(other Money) Money { return m + other }
func (m Money) Sub(other Money) Money { return m - other }
func (m Money) IsNegative() bool      { return m < 0 }
func (m Money) IsPositive() bool      { return m > 0 }

func MultiplyMoneyByQuantity(unitPrice Money, quantity float64) (Money, error) {
	if math.IsNaN(quantity) || math.IsInf(quantity, 0) || quantity < 0 {
		return 0, errors.New("qty tidak valid")
	}
	scaledFloat := quantity * 100
	scaledQuantity := math.Round(scaledFloat)
	if math.Abs(scaledFloat-scaledQuantity) > 0.0000001 || scaledQuantity > math.MaxInt64 {
		return 0, errors.New("qty maksimal dua angka desimal")
	}

	product := new(big.Int).Mul(big.NewInt(int64(unitPrice)), big.NewInt(int64(scaledQuantity)))
	half := big.NewInt(50)
	if product.Sign() < 0 {
		product.Sub(product, half)
	} else {
		product.Add(product, half)
	}
	product.Quo(product, big.NewInt(100))
	if !product.IsInt64() {
		return 0, errors.New("total nilai uang di luar batas")
	}
	return Money(product.Int64()), nil
}

func PercentageRounded(numerator, denominator Money) int {
	if denominator <= 0 {
		return 0
	}
	value := new(big.Int).Mul(big.NewInt(int64(numerator)), big.NewInt(100))
	value.Add(value, big.NewInt(int64(denominator)/2))
	value.Quo(value, big.NewInt(int64(denominator)))
	if !value.IsInt64() {
		return math.MaxInt
	}
	return int(value.Int64())
}

// ExceedsVariance reports whether actual is above base by more than the
// configured percentage in basis points (100 basis points = 1%).
func ExceedsVariance(actual, base Money, toleranceBPS int64) bool {
	if actual <= base {
		return false
	}
	if base < 0 || toleranceBPS < 0 {
		return true
	}
	excess := new(big.Int).Sub(big.NewInt(int64(actual)), big.NewInt(int64(base)))
	excess.Mul(excess, big.NewInt(10_000))
	allowed := new(big.Int).Mul(big.NewInt(int64(base)), big.NewInt(toleranceBPS))
	return excess.Cmp(allowed) > 0
}
