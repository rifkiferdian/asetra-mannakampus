package models

type PurchaseRequest struct {
	ID                         int64
	PRNumber                   string
	RequesterUserID            int
	RequesterName              string
	StoreID                    int
	StoreCode                  string
	StoreName                  string
	DivisionID                 int
	DivisionName               string
	GLAccountID                int
	GLAccountName              string
	SpendType                  string
	UrgentLevel                string
	NeededDate                 string
	Justification              string
	RequestTitle               string
	RequestCategory            string
	DeliveryLocation           string
	ImpactIfNotApproved        string
	UrgencyReason              string
	RecommendedVendorID        int64
	RecommendedVendorName      string
	VendorRecommendationReason string
	IsSingleSource             bool
	SingleSourceReason         string
	BudgetExceptionReason      string
	AssetRequestType           string
	ExistingAssetCode          string
	AssetLocation              string
	AssetPIC                   string
	TotalAmount                float64
	TotalAmountDisplay         string
	Status                     string
	StatusLabel                string
	CurrentStep                string
	SLALabel                   string
	SLAState                   string
	CreatedAtDisplay           string
}

type PurchaseRequestDetail struct {
	PurchaseRequest
	Items             []PurchaseRequestItem
	Attachments       []Attachment
	ApprovalSteps     []PurchaseRequestApprovalStep
	CurrentUserTaskID int64
	BudgetImpactLabel string
	BudgetUtilizedPct int
	BudgetMessage     string
}

type PurchaseRequestItem struct {
	ID                  int64
	PRID                int64
	ItemName            string
	Qty                 float64
	QtyDisplay          string
	UOM                 string
	EstUnitPrice        float64
	EstUnitPriceDisplay string
	EstTotal            float64
	EstTotalDisplay     string
	Notes               string
	Specification       string
	PriceSource         string
}

type PurchaseRequestApprovalStep struct {
	TaskID           int64
	StepOrder        int
	RoleName         string
	AssignedUserName string
	Status           string
	StatusLabel      string
	ActedAtDisplay   string
	CreatedAtDisplay string
}

type PurchaseRequestCreateInput struct {
	RequesterUserID            int
	StoreID                    int
	DivisionID                 int
	GLAccountID                int
	SpendType                  string
	UrgentLevel                string
	NeededDate                 string
	Justification              string
	RequestTitle               string
	RequestCategory            string
	DeliveryLocation           string
	ImpactIfNotApproved        string
	UrgencyReason              string
	RecommendedVendorID        int64
	VendorRecommendationReason string
	IsSingleSource             bool
	SingleSourceReason         string
	BudgetExceptionReason      string
	AssetRequestType           string
	ExistingAssetCode          string
	AssetLocation              string
	AssetPIC                   string
	Action                     string
	Items                      []PurchaseRequestItemInput
	Attachments                []AttachmentFileInput
	AuditContext               AuditContext
}

type PurchaseRequestUpdateInput struct {
	ID                         int64
	StoreID                    int
	DivisionID                 int
	GLAccountID                int
	SpendType                  string
	UrgentLevel                string
	NeededDate                 string
	Justification              string
	RequestTitle               string
	RequestCategory            string
	DeliveryLocation           string
	ImpactIfNotApproved        string
	UrgencyReason              string
	RecommendedVendorID        int64
	VendorRecommendationReason string
	IsSingleSource             bool
	SingleSourceReason         string
	BudgetExceptionReason      string
	AssetRequestType           string
	ExistingAssetCode          string
	AssetLocation              string
	AssetPIC                   string
	Items                      []PurchaseRequestItemInput
	AuditContext               AuditContext
}

type PurchaseRequestItemInput struct {
	ItemName      string
	Qty           float64
	UOM           string
	EstUnitPrice  float64
	Notes         string
	Specification string
	PriceSource   string
}

type PurchaseRequestBudgetCheck struct {
	BudgetID        int64   `json:"budget_id"`
	PeriodLabel     string  `json:"period_label"`
	Amount          float64 `json:"amount"`
	UsedAmount      float64 `json:"used_amount"`
	RemainingAmount float64 `json:"remaining_amount"`
	PRAmount        float64 `json:"pr_amount"`
	AfterPRAmount   float64 `json:"after_pr_amount"`
	UtilizedPct     int     `json:"utilized_pct"`
	Status          string  `json:"status"`
	Message         string  `json:"message"`
}

type PurchaseRequestApprovalPreviewStep struct {
	StepOrder    int    `json:"step_order"`
	RoleName     string `json:"role_name"`
	ApproverName string `json:"approver_name"`
}

type PurchaseRequestFormCheck struct {
	SpendType     string                               `json:"spend_type"`
	Budget        PurchaseRequestBudgetCheck           `json:"budget"`
	ApprovalRule  string                               `json:"approval_rule"`
	ApprovalSteps []PurchaseRequestApprovalPreviewStep `json:"approval_steps"`
}

type Attachment struct {
	ID               int64
	RefType          string
	RefID            int64
	FilePath         string
	FileName         string
	MimeType         string
	FileSize         int64
	UploadedBy       int
	CreatedAtDisplay string
}

type AttachmentFileInput struct {
	FileName string
	FilePath string
	MimeType string
	FileSize int64
}

type AuditContext struct {
	ActorUserID int
	IPAddress   string
	UserAgent   string
}

type Division struct {
	ID           int
	DivisionCode string
	DivisionName string
}

type DivisionCreateInput struct {
	DivisionCode string
	DivisionName string
}

type DivisionUpdateInput struct {
	ID           int
	DivisionCode string
	DivisionName string
}
