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
	ID                int64
	ItemCode          string
	CategoryID        int64
	CategoryName      string
	ItemTypeID        int64
	ItemTypeCode      string
	ItemTypeName      string
	AssetTypeID       int64
	AssetTypeName     string
	ComponentTypeID   int64
	ComponentTypeName string
	Name              string
	UOM               string
	Description       string
	IsAssetCandidate  bool
	IsPurchasable     bool
	IsActive          bool
	VariantCount      int
	BOMCount          int
	PackageCount      int
	PriceCount        int
	UpdatedAtDisplay  string
}

type CatalogItemInput struct {
	ID               int64
	ItemCode         string
	CategoryID       int64
	ItemTypeID       int64
	AssetTypeID      int64
	ComponentTypeID  int64
	Name             string
	UOM              string
	Description      string
	IsAssetCandidate bool
	IsPurchasable    bool
	IsActive         bool
}

type CatalogBrand struct {
	ID               int64
	Code             string
	Name             string
	Description      string
	IsActive         bool
	VariantCount     int
	UpdatedAtDisplay string
}

type CatalogBrandInput struct {
	ID          int64
	Code        string
	Name        string
	Description string
	IsActive    bool
}

type CatalogItemVariant struct {
	ID               int64
	ItemID           int64
	ItemCode         string
	ItemName         string
	BrandID          int64
	BrandName        string
	VariantCode      string
	ModelName        string
	ManufacturerSKU  string
	Specification    string
	IsActive         bool
	PriceCount       int
	BOMUsageCount    int
	PackageCount     int
	UpdatedAtDisplay string
}

type CatalogItemVariantInput struct {
	ID              int64
	ItemID          int64
	BrandID         int64
	VariantCode     string
	ModelName       string
	ManufacturerSKU string
	Specification   string
	IsActive        bool
}

type CatalogItemBOM struct {
	ID                    int64
	ParentItemID          int64
	ParentItemCode        string
	ParentItemName        string
	BOMCode               string
	Name                  string
	VersionNo             int
	Description           string
	IsActive              bool
	ComponentCount        int
	MissingPriceCount     int
	EstimatedTotal        float64
	EstimatedTotalDisplay string
	PackageCount          int
	UpdatedAtDisplay      string
}

type CatalogItemBOMInput struct {
	ID           int64
	ParentItemID int64
	BOMCode      string
	Name         string
	VersionNo    int
	Description  string
	IsActive     bool
}

type CatalogItemBOMItem struct {
	ID                  int64
	BOMID               int64
	BOMCode             string
	BOMName             string
	ParentItemCode      string
	ParentItemName      string
	ComponentItemID     int64
	ComponentItemCode   string
	ComponentItemName   string
	VariantID           int64
	VariantCode         string
	VariantName         string
	BrandName           string
	PreferredVendorID   int64
	PreferredVendorName string
	Qty                 float64
	QtyDisplay          string
	IsRequired          bool
	SortOrder           int
	Notes               string
	UnitPrice           float64
	UnitPriceDisplay    string
	LineTotal           float64
	LineTotalDisplay    string
	HasCurrentPrice     bool
	UpdatedAtDisplay    string
}

type CatalogItemBOMItemInput struct {
	ID                int64
	BOMID             int64
	ComponentItemID   int64
	VariantID         int64
	PreferredVendorID int64
	Qty               float64
	IsRequired        bool
	SortOrder         int
	Notes             string
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
	ID                        int64
	PackageID                 int64
	PackageCode               string
	PackageName               string
	ItemID                    int64
	ItemCode                  string
	ItemName                  string
	ItemTypeCode              string
	CategoryName              string
	BOMID                     int64
	BOMCode                   string
	BOMName                   string
	VariantID                 int64
	VariantCode               string
	VariantName               string
	BrandName                 string
	PreferredVendorID         int64
	PreferredVendorName       string
	EstimatedUnitPrice        float64
	EstimatedUnitPriceDisplay string
	HasCurrentPrice           bool
	Qty                       float64
	QtyDisplay                string
	UOM                       string
	IsOptional                bool
	SortOrder                 int
	Notes                     string
	UpdatedAtDisplay          string
}

type CatalogPackageItemInput struct {
	ID                int64
	PackageID         int64
	ItemID            int64
	BOMID             int64
	VariantID         int64
	PreferredVendorID int64
	Qty               float64
	IsOptional        bool
	SortOrder         int
	Notes             string
}

type VendorItemPrice struct {
	ID                 int64
	VendorID           int64
	VendorName         string
	ItemID             int64
	ItemCode           string
	ItemName           string
	VariantID          int64
	VariantCode        string
	VariantName        string
	BrandName          string
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
	VariantID          int64
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
