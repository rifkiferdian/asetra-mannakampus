package models

type PurchaseRequest struct {
	ID                 int64
	PRNumber           string
	RequesterUserID    int
	RequesterName      string
	StoreID            int
	StoreCode          string
	StoreName          string
	DivisionID         int
	DivisionName       string
	GLAccountID        int
	GLAccountName      string
	SpendType          string
	UrgentLevel        string
	NeededDate         string
	Justification      string
	TotalAmount        float64
	TotalAmountDisplay string
	Status             string
	StatusLabel        string
	CreatedAtDisplay   string
}

type PurchaseRequestDetail struct {
	PurchaseRequest
	Items       []PurchaseRequestItem
	Attachments []Attachment
}

type PurchaseRequestItem struct {
	ID           int64
	PRID         int64
	ItemName     string
	Qty          float64
	UOM          string
	EstUnitPrice float64
	EstTotal     float64
	Notes        string
}

type PurchaseRequestCreateInput struct {
	RequesterUserID int
	StoreID         int
	DivisionID      int
	GLAccountID     int
	SpendType       string
	UrgentLevel     string
	NeededDate      string
	Justification   string
	Action          string
	Items           []PurchaseRequestItemInput
	Attachments     []AttachmentFileInput
	AuditContext    AuditContext
}

type PurchaseRequestItemInput struct {
	ItemName     string
	Qty          float64
	UOM          string
	EstUnitPrice float64
	Notes        string
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
