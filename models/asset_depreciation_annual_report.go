package models

type AnnualDepreciationReportFilter struct {
	YearFrom    int
	YearTo      int
	Mode        string
	AssetTypeID int64
	StoreID     int
	LocationID  int64
	AssetStatus string
	Search      string
}

type AnnualDepreciationYear struct {
	Year      int
	AsOfLabel string
}

type AnnualDepreciationAmount struct {
	Year                           int
	Depreciation                   float64
	DepreciationDisplay            string
	AccumulatedDepreciation        float64
	AccumulatedDepreciationDisplay string
	BookValue                      float64
	BookValueDisplay               string
}

type AnnualDepreciationRow struct {
	Sequence                int
	AssetID                 int64
	AssetCode               string
	AssetName               string
	AssetTypeID             int64
	AssetTypeCode           string
	AssetTypeName           string
	AcquisitionDate         string
	AcquisitionDateDisplay  string
	AcquisitionYear         int
	StoreName               string
	LocationName            string
	AssetStatus             string
	ProfileStatus           string
	AcquisitionValue        float64
	AcquisitionValueDisplay string
	DepreciableBasis        float64
	SalvageValue            float64
	YearAmounts             []AnnualDepreciationAmount
}

type AnnualDepreciationGroup struct {
	AssetTypeID             int64
	AssetTypeCode           string
	AssetTypeName           string
	Rows                    []AnnualDepreciationRow
	AssetCount              int
	AcquisitionValue        float64
	AcquisitionValueDisplay string
	YearTotals              []AnnualDepreciationAmount
}

type AnnualDepreciationReportResult struct {
	Years                     []AnnualDepreciationYear
	Groups                    []AnnualDepreciationGroup
	AssetCount                int
	AcquisitionValue          float64
	AcquisitionValueDisplay   string
	YearTotals                []AnnualDepreciationAmount
	LatestBookValueDisplay    string
	LatestDepreciationDisplay string
	LatestAccumulatedDisplay  string
	LatestYear                int
}
