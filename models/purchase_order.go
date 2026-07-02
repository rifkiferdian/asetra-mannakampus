package models

type PurchaseOrder struct {
	ID                 int64
	PONumber           string
	PRID               int64
	PRNumber           string
	VendorID           int64
	VendorName         string
	StoreID            int
	StoreCode          string
	StoreName          string
	DivisionID         int
	DivisionName       string
	TotalAmount        float64
	TotalAmountDisplay string
	Status             string
	StatusLabel        string
	CreatedAtDisplay   string
}

type PurchaseOrderDetail struct {
	PurchaseOrder
	Items []PurchaseOrderItem
}

type PurchaseOrderItem struct {
	ID               int64
	POID             int64
	ItemName         string
	Qty              float64
	QtyDisplay       string
	UOM              string
	UnitPrice        float64
	UnitPriceDisplay string
	Total            float64
	TotalDisplay     string
}

type ApprovedPRForPO struct {
	ID                 int64
	PRNumber           string
	RequesterName      string
	StoreName          string
	DivisionName       string
	GLAccountName      string
	SpendType          string
	TotalAmount        float64
	TotalAmountDisplay string
	ApprovedAtDisplay  string
}

type PurchaseOrderCreateForm struct {
	PR      PurchaseRequestDetail
	Vendors []Vendor
}

type PurchaseOrderCreateInput struct {
	PRID         int64
	VendorID     int64
	Items        []PurchaseOrderItemInput
	AuditContext AuditContext
}

type PurchaseOrderItemInput struct {
	PRItemID  int64
	ItemName  string
	Qty       float64
	UOM       string
	UnitPrice float64
}
