package models

type CatalogItemType struct {
	ID               int64
	Code             string
	Name             string
	Description      string
	IsActive         bool
	ItemCount        int
	UpdatedAtDisplay string
}

type CatalogItemTypeInput struct {
	ID          int64
	Code        string
	Name        string
	Description string
	IsActive    bool
}

type CatalogItemCategory struct {
	ID                int64
	Code              string
	Name              string
	ParentID          int64
	ParentName        string
	OwnerDivisionID   int
	OwnerDivisionName string
	Description       string
	IsActive          bool
	ItemCount         int
	ChildCount        int
	UpdatedAtDisplay  string
}

type CatalogItemCategoryInput struct {
	ID              int64
	Code            string
	Name            string
	ParentID        int64
	OwnerDivisionID int
	Description     string
	IsActive        bool
}

type CatalogItem struct {
	ID               int64
	ItemCode         string
	CategoryID       int64
	CategoryName     string
	ItemTypeID       int64
	ItemTypeCode     string
	ItemTypeName     string
	AssetTypeID      int64
	AssetTypeName    string
	Name             string
	UOM              string
	Description      string
	IsAssetCandidate bool
	IsActive         bool
	DetailCount      int
	PackageCount     int
	PriceCount       int
	UpdatedAtDisplay string
}

type CatalogItemInput struct {
	ID               int64
	ItemCode         string
	CategoryID       int64
	ItemTypeID       int64
	AssetTypeID      int64
	Name             string
	UOM              string
	Description      string
	IsAssetCandidate bool
	IsActive         bool
}

type CatalogItemDetail struct {
	ID               int64
	ItemID           int64
	ItemCode         string
	ItemName         string
	CategoryName     string
	DetailName       string
	DetailValue      string
	Unit             string
	SortOrder        int
	UpdatedAtDisplay string
}

type CatalogItemDetailInput struct {
	ID          int64
	ItemID      int64
	DetailName  string
	DetailValue string
	Unit        string
	SortOrder   int
}

type CatalogPackage struct {
	ID                int64
	PackageCode       string
	Name              string
	CategoryID        int64
	CategoryName      string
	OwnerDivisionID   int
	OwnerDivisionName string
	Description       string
	IsActive          bool
	ItemCount         int
	RequiredCount     int
	OptionalCount     int
	UpdatedAtDisplay  string
}

type CatalogPackageInput struct {
	ID              int64
	PackageCode     string
	Name            string
	CategoryID      int64
	OwnerDivisionID int
	Description     string
	IsActive        bool
}

type CatalogPackageItem struct {
	ID               int64
	PackageID        int64
	PackageCode      string
	PackageName      string
	ItemID           int64
	ItemCode         string
	ItemName         string
	ItemTypeCode     string
	CategoryName     string
	Qty              float64
	QtyDisplay       string
	UOM              string
	IsOptional       bool
	SortOrder        int
	Notes            string
	UpdatedAtDisplay string
}

type CatalogPackageItemInput struct {
	ID         int64
	PackageID  int64
	ItemID     int64
	Qty        float64
	IsOptional bool
	SortOrder  int
	Notes      string
}

type VendorItemPrice struct {
	ID                 int64
	VendorID           int64
	VendorName         string
	ItemID             int64
	ItemCode           string
	ItemName           string
	CategoryName       string
	UOM                string
	UnitPrice          float64
	UnitPriceInput     string
	UnitPriceDisplay   string
	CurrencyCode       string
	MinimumQty         float64
	MinimumQtyDisplay  string
	ValidFrom          string
	ValidFromDisplay   string
	ValidUntil         string
	ValidUntilDisplay  string
	LeadTimeDays       int
	QuotationReference string
	IsPreferred        bool
	IsActive           bool
	IsCurrentlyValid   bool
	Notes              string
	UpdatedAtDisplay   string
}

type VendorItemPriceInput struct {
	ID                 int64
	VendorID           int64
	ItemID             int64
	UnitPrice          float64
	CurrencyCode       string
	MinimumQty         float64
	ValidFrom          string
	ValidUntil         string
	LeadTimeDays       int
	QuotationReference string
	IsPreferred        bool
	IsActive           bool
	Notes              string
}
