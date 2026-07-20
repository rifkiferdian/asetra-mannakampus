package models

type AssetDisposalReportFilter struct {
	DateFrom       string
	DateTo         string
	DisposalTypeID int64
	AssetTypeID    int64
	StoreID        int
	Result         string
	Status         string
	Search         string
	Page           int
	PerPage        int
}

type AssetDisposalReportRow struct {
	ID                             int64
	DisposalNumber                 string
	DisposalDate                   string
	DisposalDateDisplay            string
	AssetID                        int64
	AssetCode                      string
	AssetName                      string
	AssetTypeName                  string
	StoreName                      string
	DisposalTypeName               string
	BuyerName                      string
	DocumentReference              string
	AcquisitionValue               float64
	AcquisitionValueDisplay        string
	AccumulatedDepreciation        float64
	AccumulatedDepreciationDisplay string
	BookValue                      float64
	BookValueDisplay               string
	DisposalValue                  float64
	DisposalValueDisplay           string
	GainLossAmount                 float64
	GainLossAmountDisplay          string
	GainLossLabel                  string
	Status                         string
	PostedAtDisplay                string
	PostedByName                   string
	ReversalReason                 string
}

type AssetDisposalReportSummary struct {
	TransactionCount               int
	PostedCount                    int
	ReversedCount                  int
	AcquisitionValue               float64
	AcquisitionValueDisplay        string
	AccumulatedDepreciation        float64
	AccumulatedDepreciationDisplay string
	BookValue                      float64
	BookValueDisplay               string
	DisposalValue                  float64
	DisposalValueDisplay           string
	ProfitAmount                   float64
	ProfitAmountDisplay            string
	LossAmount                     float64
	LossAmountDisplay              string
	NetGainLoss                    float64
	NetGainLossDisplay             string
	NetGainLossLabel               string
}

type AssetDisposalReportResult struct {
	Items      []AssetDisposalReportRow
	Summary    AssetDisposalReportSummary
	TotalRows  int
	TotalPages int
}
