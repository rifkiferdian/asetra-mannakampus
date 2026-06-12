package models

type Vendor struct {
	ID               int64
	Name             string
	Phone            string
	Email            string
	Address          string
	IsActive         bool
	IsActiveLabel    string
	CreatedAt        string
	CreatedAtDisplay string
}

type VendorCreateInput struct {
	Name     string
	Phone    string
	Email    string
	Address  string
	IsActive bool
}

type VendorUpdateInput struct {
	ID       int64
	Name     string
	Phone    string
	Email    string
	Address  string
	IsActive bool
}
