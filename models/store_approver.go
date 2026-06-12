package models

type StoreApprover struct {
	ID            int64
	StoreID       int
	StoreName     string
	RoleID        int64
	RoleName      string
	UserID        int
	UserName      string
	Username      string
	IsActive      bool
	IsActiveLabel string
}

type StoreApproverCreateInput struct {
	StoreID  int
	RoleID   int64
	UserID   int
	IsActive bool
}

type StoreApproverUpdateInput struct {
	ID       int64
	StoreID  int
	RoleID   int64
	UserID   int
	IsActive bool
}
