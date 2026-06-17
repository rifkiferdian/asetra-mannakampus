package models

type Budget struct {
	ID                int64
	FiscalYear        int
	PeriodType        string
	PeriodKey         string
	StoreID           int
	StoreName         string
	DivisionID        int
	DivisionName      string
	GLAccountID       int
	GLAccountName     string
	Amount            float64
	AmountDisplay     string
	UsedAmount        float64
	UsedAmountDisplay string
	RemainingAmount   float64
	RemainingDisplay  string
	CreatedAt         string
	CreatedAtDisplay  string
	UpdatedAt         string
	UpdatedAtDisplay  string
}

type BudgetCreateInput struct {
	FiscalYear  int
	PeriodType  string
	PeriodKey   string
	StoreID     int
	DivisionID  int
	GLAccountID int
	Amount      float64
}

type BudgetUpdateInput struct {
	ID          int64
	FiscalYear  int
	PeriodType  string
	PeriodKey   string
	StoreID     int
	DivisionID  int
	GLAccountID int
	Amount      float64
}
