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
	Amount            Money
	AmountDisplay     string
	UsedAmount        Money
	UsedAmountDisplay string
	RemainingAmount   Money
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
	Amount      Money
}

type BudgetUpdateInput struct {
	ID          int64
	FiscalYear  int
	PeriodType  string
	PeriodKey   string
	StoreID     int
	DivisionID  int
	GLAccountID int
	Amount      Money
}
