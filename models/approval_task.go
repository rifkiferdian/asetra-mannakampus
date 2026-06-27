package models

type ApprovalTaskInboxItem struct {
	ID               int64
	ApprovalID       int64
	RefType          string
	RefID            int64
	DocumentNumber   string
	RequesterName    string
	StoreName        string
	RoleName         string
	StepOrder        int
	Scope            string
	Amount           float64
	AmountDisplay    string
	SpendType        string
	UrgentLevel      string
	NeededDate       string
	Status           string
	CreatedAtDisplay string
}

type ApprovalTaskInboxFilter struct {
	UserID         int
	Urgency        string
	SpendType      string
	NeededDateSort string
	Page           int
	PerPage        int
}

type ApprovalTaskInboxResult struct {
	Items      []ApprovalTaskInboxItem
	TotalRows  int
	QueueValue float64
}

type PaginationPage struct {
	Page   int
	URL    string
	Active bool
}

type PaginationView struct {
	Page       int
	PerPage    int
	TotalRows  int
	TotalPages int
	From       int
	To         int
	HasPrev    bool
	HasNext    bool
	PrevURL    string
	NextURL    string
	Pages      []PaginationPage
}

type ApprovalTaskDetail struct {
	ApprovalTaskInboxItem
	Justification         string
	ApprovalStatus        string
	DocumentStatus        string
	CurrentApprovalStep   string
	CurrentApprovalStatus string
}

type ApprovalActionInput struct {
	TaskID       int64
	ActorUserID  int
	Action       string
	Comment      string
	AuditContext AuditContext
}
