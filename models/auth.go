package models

// AuthAccount contains only the user fields required during authentication.
// It deliberately excludes the password from the session payload.
type AuthAccount struct {
	UserID       int
	Username     string
	Name         string
	PasswordHash string
	NIP          string
	Role         string
	StoreID      string
}
