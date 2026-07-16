package models

type MonthlyDepreciationFilter struct {
	Year    int
	Month   int
	Status  string
	Search  string
	Page    int
	PerPage int
}

type MonthlyDepreciationItem struct {
	ID                             int64
	AssetID                        int64
	AssetCode                      string
	AssetName                      string
	AssetTypeName                  string
	MethodName                     string
	UsefulLifeMonths               int
	PeriodDate                     string
	AcquisitionValueDisplay        string
	OpeningBookValueDisplay        string
	DepreciationAmountDisplay      string
	AccumulatedDepreciationDisplay string
	ClosingBookValueDisplay        string
	Status                         string
	PostedAtDisplay                string
}

type MonthlyDepreciationStats struct {
	TotalAssets               int
	DraftCount                int
	PostedCount               int
	SkippedCount              int
	TotalDepreciationDisplay  string
	DraftDepreciationDisplay  string
	PostedDepreciationDisplay string
}

type MonthlyDepreciationResult struct {
	Items      []MonthlyDepreciationItem
	Stats      MonthlyDepreciationStats
	TotalRows  int
	TotalPages int
}

type DepreciationMonthOption struct {
	Value int
	Label string
}

type DepreciationPagination struct {
	CurrentPage int
	TotalPages  int
	PageStart   int
	PageEnd     int
	TotalRows   int
	PageSize    int
	HasPrev     bool
	HasNext     bool
	PrevURL     string
	NextURL     string
}

type AssetDepreciationDetail struct {
	Configured                 bool
	ProfileID                  int64
	MethodCode                 string
	MethodName                 string
	UsefulLifeMonths           int
	StartDateDisplay           string
	ProfileStatus              string
	SalvageValueDisplay        string
	DepreciableBasisDisplay    string
	MonthlyDepreciationDisplay string
	PostedDepreciationDisplay  string
	CurrentBookValueDisplay    string
	ProgressPercent            float64
	ProgressPercentDisplay     string
	PostedScheduleCount        int
	DraftScheduleCount         int
	LastPostedPeriodDisplay    string
	NextDraftPeriodDisplay     string
}

type AssetDepreciationPosting struct {
	ID                             int64
	PeriodDisplay                  string
	OpeningBookValueDisplay        string
	DepreciationAmountDisplay      string
	AccumulatedDepreciationDisplay string
	ClosingBookValueDisplay        string
	PostedAtDisplay                string
}
