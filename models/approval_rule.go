package models

type ApprovalRule struct {
	ID              int64
	Name            string
	IsActive        bool
	IsActiveLabel   string
	MinAmount       float64
	MaxAmount       *float64
	MaxAmountLabel  string
	LocationScope   string
	SpendType       string
	UrgentLevel     string
	StepCount       int
	CreatedAt       string
	CreatedAtDisplay string
}

type ApprovalRuleStep struct {
	ID         int64
	RuleID     int64
	StepOrder  int
	RoleID     int64
	RoleName   string
	Scope      string
	IsParallel bool
	IsRequired bool
}

type ApprovalRuleDetail struct {
	ID            int64
	Name          string
	IsActive      bool
	MinAmount     float64
	MaxAmount     *float64
	LocationScope string
	SpendType     string
	UrgentLevel   string
	Steps         []ApprovalRuleStep
}

type ApprovalRuleCreateInput struct {
	Name          string
	IsActive      bool
	MinAmount     float64
	MaxAmount     *float64
	LocationScope string
	SpendType     string
	UrgentLevel   string
	Steps         []ApprovalRuleStepInput
}

type ApprovalRuleUpdateInput struct {
	ID            int64
	Name          string
	IsActive      bool
	MinAmount     float64
	MaxAmount     *float64
	LocationScope string
	SpendType     string
	UrgentLevel   string
	Steps         []ApprovalRuleStepInput
}

type ApprovalRuleStepInput struct {
	StepOrder  int
	RoleID     int64
	Scope      string
	IsParallel bool
	IsRequired bool
}
