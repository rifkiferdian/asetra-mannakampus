package models

// Permission represents a permission record.
type Permission struct {
	ID        int64
	Name      string
	GroupName string
	GuardName string
}

// PermissionGroup bundles permissions under the same group name.
type PermissionGroup struct {
	Key         string
	Label       string
	Permissions []Permission
}
