package models

type AssetDisposalApprovalRule struct {
	ID                    int64
	Name                  string
	DisposalTypeID        int64
	DisposalTypeName      string
	AssetTypeID           int64
	AssetTypeName         string
	MinBookValue          float64
	MinBookValueInput     string
	MinBookValueDisplay   string
	MaxBookValue          *float64
	MaxBookValueInput     string
	MaxBookValueDisplay   string
	Priority              int
	IsActive              bool
	EffectiveFrom         string
	EffectiveUntil        string
	EffectivePeriodLabel  string
	Steps                 []AssetDisposalApprovalRuleStep
	StepSummary           string
	ApprovalCount         int
}

type AssetDisposalApprovalRuleStep struct {
	ID         int64
	StepOrder  int
	RoleID     int64
	RoleName   string
	Scope      string
	IsParallel bool
	IsRequired bool
}

type AssetDisposalApprovalRuleInput struct {
	ID               int64
	Name             string
	DisposalTypeID   int64
	AssetTypeID      int64
	MinBookValue     float64
	MaxBookValue     *float64
	Priority         int
	IsActive         bool
	EffectiveFrom    string
	EffectiveUntil   string
	Steps            []AssetDisposalApprovalRuleStep
}

type AssetDisposalApprover struct {
	ID         int64
	Scope      string
	StoreID    int
	StoreName  string
	RoleID     int64
	RoleName   string
	UserID     int
	UserName   string
	IsActive   bool
	UpdatedAt  string
}

type AssetDisposalApproverInput struct {
	ID       int64
	Scope    string
	StoreID  int
	RoleID   int64
	UserID   int
	IsActive bool
}

type AssetDisposalApprovalInboxFilter struct {
	Status  string
	Search  string
	Page    int
	PerPage int
}

type AssetDisposalApprovalTask struct {
	ID                   int64
	ApprovalID           int64
	DisposalID           int64
	DisposalNumber       string
	AssetID               int64
	AssetCode             string
	AssetName             string
	StoreName             string
	DisposalTypeName      string
	DisposalDateDisplay   string
	DisposalValueDisplay  string
	BookValueDisplay      string
	RuleName              string
	AttemptNo             int
	StepOrder             int
	RoleName              string
	Scope                 string
	AssignedUserName      string
	Status                string
	Comment               string
	SubmittedByName       string
	SubmittedAtDisplay    string
	ActedAtDisplay        string
}

type AssetDisposalApprovalInboxStats struct {
	Waiting  int
	Pending  int
	Approved int
	Rejected int
}

type AssetDisposalApprovalInboxResult struct {
	Items      []AssetDisposalApprovalTask
	Stats      AssetDisposalApprovalInboxStats
	TotalRows  int
	TotalPages int
}
