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
	MethodCode                     string
	MethodName                     string
	UsefulLifeMonths               int
	PeriodDate                     string
	VersionNo                      int
	OriginalScheduleID             int64
	IsCorrection                   bool
	CorrectionReason               string
	AcquisitionValueDisplay        string
	OpeningBookValueDisplay        string
	DepreciationAmountDisplay      string
	AccumulatedDepreciationDisplay string
	ClosingBookValueDisplay        string
	Status                         string
	ActionAtDisplay                string
	ActionByName                   string
	SkipReason                     string
	ReversalReason                 string
	DepreciationAmountInput        string
}

type MonthlyDepreciationStats struct {
	TotalAssets               int
	DraftCount                int
	PostedCount               int
	SkippedCount              int
	ReversedCount             int
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

type DepreciationPeriod struct {
	ID                 int64
	Year               int
	Month              int
	Status             string
	CanClose           bool
	GeneratedAtDisplay string
	PostedAtDisplay    string
	ClosedAtDisplay    string
	ClosedByName       string
	ClosingNotes       string
	ReopenedAtDisplay  string
	ReopenedByName     string
	ReopenReason       string
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

type DepreciationProfileFilter struct {
	Status  string
	Search  string
	Page    int
	PerPage int
}

type AssetDepreciationProfile struct {
	ID                         int64
	AssetID                    int64
	AssetCode                  string
	AssetName                  string
	AssetTypeName              string
	MethodID                   int64
	MethodCode                 string
	MethodName                 string
	UsefulLifeMonths           int
	SalvageValueInput          string
	SalvageValueDisplay        string
	DepreciableBasisInput      string
	DepreciableBasisDisplay    string
	MonthlyDepreciationDisplay string
	PostedDepreciationDisplay  string
	CurrentBookValueDisplay    string
	StartDate                  string
	Status                     string
	Notes                      string
	PostedScheduleCount        int
	DraftScheduleCount         int
	LastPostedPeriodDisplay    string
	ConfigurationLocked        bool
}

type DepreciationProfileStats struct {
	TotalProfiles      int
	ActiveProfiles     int
	PausedProfiles     int
	FinishedProfiles   int
	UnconfiguredAssets int
}

type DepreciationProfileResult struct {
	Items      []AssetDepreciationProfile
	Stats      DepreciationProfileStats
	TotalRows  int
	TotalPages int
}

type DepreciationMethodOption struct {
	ID   int64
	Code string
	Name string
}

type DepreciationAssetOption struct {
	ID                    int64
	AssetCode             string
	AssetName             string
	AssetTypeName         string
	AcquisitionDate       string
	AcquisitionValueInput string
	HasProfile            bool
}

type DepreciationProfileInput struct {
	ID               int64
	AssetID          int64
	MethodID         int64
	UsefulLifeMonths int
	SalvageValue     float64
	DepreciableBasis float64
	StartDate        string
	Status           string
	Notes            string
	AuditContext     AuditContext
}

type DepreciationPostingHistoryFilter struct {
	Year    int
	Month   int
	Search  string
	Page    int
	PerPage int
}

type DepreciationPostingHistoryItem struct {
	ID                             int64
	AssetID                        int64
	AssetCode                      string
	AssetName                      string
	AssetTypeName                  string
	MethodCode                     string
	MethodName                     string
	VersionNo                      int
	OriginalScheduleID             int64
	Status                         string
	PeriodStatus                   string
	PeriodYear                     int
	PeriodMonth                    int
	PeriodDisplay                  string
	OpeningBookValueDisplay        string
	DepreciationAmountDisplay      string
	AccumulatedDepreciationDisplay string
	ClosingBookValueDisplay        string
	PostedAtDisplay                string
	PostedByName                   string
	ReversedAtDisplay              string
	ReversedByName                 string
	ReversalReason                 string
}

type DepreciationCorrectionInput struct {
	ScheduleID        int64
	DepreciationValue float64
	Reason            string
	AuditContext      AuditContext
}

type DepreciationPostingHistoryStats struct {
	TotalPostings        int
	TotalAssets          int
	TotalAmountDisplay   string
	LatestPostingDisplay string
}

type DepreciationPostingHistoryResult struct {
	Items      []DepreciationPostingHistoryItem
	Stats      DepreciationPostingHistoryStats
	TotalRows  int
	TotalPages int
}
