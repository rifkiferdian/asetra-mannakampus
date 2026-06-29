package models

// User merepresentasikan data pada tabel users.
// Field StoreIDs berisi daftar id toko dari tabel user_stores.
type User struct {
	ID               int
	NIP              int
	Username         string
	Name             string
	Email            string
	Status           string
	StatusLabel      string
	StoreIDs         []int
	StoreDisplay     string
	RoleDisplay      string
	RoleNames        []string
	CreatedAt        string
	CreatedAtDisplay string
}

// UserCreateInput menampung data yang dikirimkan dari form create user.
type UserCreateInput struct {
	NIP       int
	Username  string
	Password  string
	Name      string
	Email     string
	Status    string
	StoreIDs  []int
	RoleNames []string
}

// UserUpdateInput menampung data yang dikirimkan dari form edit user.
type UserUpdateInput struct {
	ID        int
	NIP       int
	Username  string
	Password  string
	Name      string
	Email     string
	Status    string
	StoreIDs  []int
	RoleNames []string
}

// UserProfileUpdateInput menampung perubahan data diri oleh user login.
type UserProfileUpdateInput struct {
	ID       int
	Username string
	Name     string
	Email    string
}

// UserPasswordUpdateInput menampung perubahan password oleh user login.
type UserPasswordUpdateInput struct {
	ID              int
	CurrentPassword string
	NewPassword     string
	ConfirmPassword string
}
