package models

// Role mewakili data role beserta jumlah permission dan user yang terkait.
type Role struct {
	ID              int
	Name            string
	PermissionCount int
	UserCount       int
	UpdatedAt       string
}

// RoleCreateInput mewakili payload untuk membuat role baru.
type RoleCreateInput struct {
	Name          string
	GuardName     string
	PermissionIDs []int64
}

// RoleDetail mewakili detail role beserta daftar permission yang dimiliki.
type RoleDetail struct {
	ID            int
	Name          string
	GuardName     string
	IsAdmin       bool
	PermissionIDs []int64
}

// RoleUpdateInput mewakili payload untuk memperbarui role yang ada.
type RoleUpdateInput struct {
	ID            int
	Name          string
	GuardName     string
	PermissionIDs []int64
}
