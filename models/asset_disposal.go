package models

type AssetDisposalType struct {
	ID               int64
	Code             string
	Name             string
	Description      string
	IsActive         bool
	IsActiveLabel    string
	DisposalCount    int
	CreatedAtDisplay string
	UpdatedAtDisplay string
}

type AssetDisposalTypeInput struct {
	ID          int64
	Code        string
	Name        string
	Description string
	IsActive    bool
}

type AssetDisposalFilter struct {
	Status  string
	Search  string
	Page    int
	PerPage int
}

type AssetDisposal struct {
	ID                             int64
	DisposalNumber                 string
	AssetID                        int64
	AssetCode                      string
	AssetName                      string
	AssetTypeName                  string
	DisposalTypeID                 int64
	DisposalTypeCode               string
	DisposalTypeName               string
	DisposalDate                   string
	DisposalDateDisplay            string
	DisposalValueInput             string
	DisposalValueDisplay           string
	BuyerName                      string
	DocumentReference              string
	Reason                         string
	Status                         string
	ProcessedByName                string
	ApprovedByName                 string
	SubmittedByName               string
	SubmittedAtDisplay            string
	RejectedByName                string
	RejectedAtDisplay             string
	RejectionReason               string
	PostedAtDisplay                string
	PostedByName                  string
	CancelledAtDisplay             string
	CancelledByName                string
	CancellationReason             string
	ReversedByName                string
	ReversedAtDisplay             string
	ReversalReason                string
	Notes                          string
	AcquisitionValueDisplay        string
	AccumulatedDepreciationDisplay string
	BookValueDisplay               string
	GainLossAmountDisplay          string
	GainLossLabel                  string
	CreatedAtDisplay               string
}

type AssetDisposalInput struct {
	ID                int64
	AssetID           int64
	DisposalTypeID    int64
	DisposalDate      string
	DisposalValue     float64
	BuyerName         string
	DocumentReference string
	Reason            string
	Notes             string
	AuditContext      AuditContext
}

type AssetDisposalStats struct {
	Total             int
	Draft             int
	InApproval        int
	Approved          int
	Rejected          int
	Posted            int
	Cancelled         int
	TotalValueDisplay string
}

type AssetDisposalResult struct {
	Items      []AssetDisposal
	Stats      AssetDisposalStats
	TotalRows  int
	TotalPages int
}

type AssetDisposalAssetOption struct {
	ID                    int64
	AssetCode             string
	AssetName             string
	AssetTypeName         string
	AssetStatus           string
	AcquisitionDate       string
	AcquisitionValueInput string
	ProfileStatus         string
	CurrentBookValueInput string
}
