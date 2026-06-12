package models

type GLAccount struct {
	ID               int
	GLCode           string
	GLName           string
	SpendType        string
	IsActive         bool
	IsActiveLabel    string
	CreatedAt        string
	CreatedAtDisplay string
	UpdatedAt        string
	UpdatedAtDisplay string
}

type GLAccountCreateInput struct {
	GLCode    string
	GLName    string
	SpendType string
	IsActive  bool
}

type GLAccountUpdateInput struct {
	ID        int
	GLCode    string
	GLName    string
	SpendType string
	IsActive  bool
}
