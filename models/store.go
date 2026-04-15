package models

// Store merepresentasikan data toko.
type Store struct {
	StoreID          int
	StoreCode        string
	StoreName        string
	StoreAddress     string
	IsActive         bool
	IsActiveLabel    string
	CreatedAt        string
	CreatedAtDisplay string
	UpdatedAt        string
	UpdatedAtDisplay string
}

// StoreCreateInput menampung data form untuk membuat store.
type StoreCreateInput struct {
	StoreID      int
	StoreCode    string
	StoreName    string
	StoreAddress string
	IsActive     bool
}

// StoreUpdateInput menampung data form untuk memperbarui store.
type StoreUpdateInput struct {
	StoreID      int
	StoreCode    string
	StoreName    string
	StoreAddress string
	IsActive     bool
}
