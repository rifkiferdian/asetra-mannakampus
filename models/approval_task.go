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
	AmountDisplay    string
	UrgentLevel      string
	Status           string
	CreatedAtDisplay string
}

type ApprovalActionInput struct {
	TaskID       int64
	ActorUserID  int
	Action       string
	Comment      string
	AuditContext AuditContext
}
