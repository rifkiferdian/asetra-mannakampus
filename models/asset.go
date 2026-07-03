package models

type AssetType struct {
	ID               int64
	Code             string
	Name             string
	Description      string
	IsActive         bool
	IsActiveLabel    string
	AssetCount       int
	CreatedAtDisplay string
	UpdatedAtDisplay string
}

type AssetTypeInput struct {
	ID          int64
	Code        string
	Name        string
	Description string
	IsActive    bool
}

type ComponentType struct {
	ID               int64
	Code             string
	Name             string
	Description      string
	IsActive         bool
	IsActiveLabel    string
	ComponentCount   int
	CreatedAtDisplay string
	UpdatedAtDisplay string
}

type ComponentTypeInput struct {
	ID          int64
	Code        string
	Name        string
	Description string
	IsActive    bool
}

type AssetLocation struct {
	ID               int64
	LocationCode     string
	LocationName     string
	StoreID          int
	StoreName        string
	ParentID         int64
	ParentName       string
	LocationType     string
	IsActive         bool
	IsActiveLabel    string
	CreatedAtDisplay string
	UpdatedAtDisplay string
}

type AssetLocationInput struct {
	ID           int64
	LocationCode string
	LocationName string
	StoreID      int
	ParentID     int64
	LocationType string
	IsActive     bool
}

type Asset struct {
	ID                       int64
	AssetCode                string
	AssetName                string
	AssetTypeID              int64
	AssetTypeName            string
	SerialNumber             string
	StoreID                  int
	StoreName                string
	LocationID               int64
	LocationName             string
	AssignedPersonNIP        string
	AssignedPersonName       string
	AssignedPersonDepartment string
	SourceGRItemID           int64
	AcquisitionDate          string
	AcquisitionValue         float64
	AcquisitionValueDisplay  string
	Status                   string
	StatusLabel              string
	Notes                    string
	CreatedAtDisplay         string
}

type AssetInput struct {
	ID                       int64
	AssetCode                string
	AssetName                string
	AssetTypeID              int64
	SerialNumber             string
	StoreID                  int
	LocationID               int64
	AssignedPersonNIP        string
	AssignedPersonName       string
	AssignedPersonDepartment string
	SourceGRItemID           int64
	AcquisitionDate          string
	AcquisitionValue         float64
	Status                   string
	Notes                    string
}

type AssetComponent struct {
	ID                      int64
	ComponentCode           string
	ComponentName           string
	ComponentTypeID         int64
	ComponentTypeName       string
	Brand                   string
	Model                   string
	Specification           string
	SerialNumber            string
	ParentAssetID           int64
	ParentAssetCode         string
	LocationID              int64
	LocationName            string
	SourceGRItemID          int64
	AcquisitionDate         string
	AcquisitionValue        float64
	AcquisitionValueDisplay string
	Status                  string
	StatusLabel             string
	Notes                   string
	CreatedAtDisplay        string
}

type AssetComponentInput struct {
	ID               int64
	ComponentCode    string
	ComponentName    string
	ComponentTypeID  int64
	Brand            string
	Model            string
	Specification    string
	SerialNumber     string
	ParentAssetID    int64
	LocationID       int64
	SourceGRItemID   int64
	AcquisitionDate  string
	AcquisitionValue float64
	Status           string
	Notes            string
}

type AssetMovement struct {
	ID               int64
	AssetID          int64
	AssetCode        string
	MovementType     string
	FromStoreName    string
	ToStoreName      string
	FromLocationName string
	ToLocationName   string
	FromUserName     string
	ToUserName       string
	ActedByName      string
	MovementDate     string
	Notes            string
}

type AssetMovementInput struct {
	AssetID      int64
	MovementType string
	ToStoreID    int
	ToLocationID int64
	ToUserID     int
	ActedBy      int
	Notes        string
}

type AssetComponentMovement struct {
	ID               int64
	ComponentID      int64
	ComponentCode    string
	MovementType     string
	FromAssetCode    string
	ToAssetCode      string
	FromLocationName string
	ToLocationName   string
	ActedByName      string
	MovementDate     string
	Notes            string
}

type AssetComponentMovementInput struct {
	ComponentID  int64
	MovementType string
	ToAssetID    int64
	ToLocationID int64
	ActedBy      int
	Notes        string
}

type AssetSelectOption struct {
	ID    int64
	Code  string
	Name  string
	Label string
}

type UserSelectOption struct {
	ID    int
	Name  string
	Label string
}
