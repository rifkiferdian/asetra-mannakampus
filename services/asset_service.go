package services

import (
	"errors"
	"gobase-app/models"
	"regexp"
	"strings"
	"time"
)

var assetPrefixPattern = regexp.MustCompile(`^[A-Z0-9]{1,10}$`)

type AssetService struct {
	Repo AssetRepositoryPort
}

// AssetRepositoryPort keeps business rules independent from the SQL adapter
// and allows lifecycle tests to use a deterministic in-memory fake.
type AssetRepositoryPort interface {
	GetAssetTypes() ([]models.AssetType, error)
	CreateAssetType(models.AssetTypeInput) error
	UpdateAssetType(models.AssetTypeInput) error
	DeleteAssetType(int64) error
	GetAssetCategories() ([]models.AssetCategory, error)
	CreateAssetCategory(models.AssetCategoryInput) error
	UpdateAssetCategory(models.AssetCategoryInput) error
	DeleteAssetCategory(int64) error
	AssetCategoryBelongsToType(int64, int64) (bool, error)
	GetComponentTypes() ([]models.ComponentType, error)
	CreateComponentType(models.ComponentTypeInput) error
	UpdateComponentType(models.ComponentTypeInput) error
	DeleteComponentType(int64) error
	GetLocations() ([]models.AssetLocation, error)
	CreateLocation(models.AssetLocationInput) error
	UpdateLocation(models.AssetLocationInput) error
	DeleteLocation(int64) error
	LocationBelongsToStore(int64, int) (bool, error)
	GetLocationStoreID(int64) (int, error)
	GetAssets() ([]models.Asset, error)
	GetAssetsPage(int, int, models.AccessScope) ([]models.Asset, error)
	CountAssets(models.AccessScope) (int, error)
	GetAssetStatusCounts(models.AccessScope) (map[string]int, error)
	GetAssetByID(int64, ...models.AccessScope) (*models.Asset, error)
	CreateAsset(models.AssetInput) error
	UpdateAsset(models.AssetInput) error
	DeleteAsset(int64) error
	GetComponents(...models.AccessScope) ([]models.AssetComponent, error)
	GetComponentByID(int64, ...models.AccessScope) (*models.AssetComponent, error)
	CreateComponent(models.AssetComponentInput) error
	UpdateComponent(models.AssetComponentInput) error
	DeleteComponent(int64) error
	GetComponentsByAssetID(int64) ([]models.AssetComponent, error)
	GetAssetMovements(...models.AccessScope) ([]models.AssetMovement, error)
	GetAssetMovementsByAssetID(int64, int) ([]models.AssetMovement, error)
	CreateAssetMovement(models.AssetMovementInput) error
	GetComponentMovements(...models.AccessScope) ([]models.AssetComponentMovement, error)
	GetComponentMovementsByComponentID(int64) ([]models.AssetComponentMovement, error)
	CreateComponentMovement(models.AssetComponentMovementInput) error
	GetStoreOptions(...models.AccessScope) ([]models.Store, error)
	GetUserOptions() ([]models.UserSelectOption, error)
	GetAssetOptions(...models.AccessScope) ([]models.AssetSelectOption, error)
	GetComponentOptions(...models.AccessScope) ([]models.AssetSelectOption, error)
}

func NewAssetService(repo AssetRepositoryPort) *AssetService {
	return &AssetService{Repo: repo}
}

func (s *AssetService) GetAssetTypes() ([]models.AssetType, error) {
	return s.Repo.GetAssetTypes()
}

func (s *AssetService) SaveAssetType(input models.AssetTypeInput) error {
	input.Code = strings.ToUpper(strings.TrimSpace(input.Code))
	input.AssetPrefix = strings.ToUpper(strings.TrimSpace(input.AssetPrefix))
	input.Name = strings.TrimSpace(input.Name)
	input.Description = strings.TrimSpace(input.Description)
	if input.Code == "" || input.AssetPrefix == "" || input.Name == "" {
		return errors.New("kode, prefix, dan nama asset type wajib diisi")
	}
	if !assetPrefixPattern.MatchString(input.AssetPrefix) {
		return errors.New("prefix asset type hanya boleh berisi huruf dan angka, maksimal 10 karakter")
	}
	if input.ID > 0 {
		return s.Repo.UpdateAssetType(input)
	}
	return s.Repo.CreateAssetType(input)
}

func (s *AssetService) DeleteAssetType(id int64) error {
	if id <= 0 {
		return errors.New("asset type tidak valid")
	}
	return s.Repo.DeleteAssetType(id)
}

func (s *AssetService) GetAssetCategories() ([]models.AssetCategory, error) {
	return s.Repo.GetAssetCategories()
}

func (s *AssetService) SaveAssetCategory(input models.AssetCategoryInput) error {
	input.Code = strings.ToUpper(strings.TrimSpace(input.Code))
	input.CategoryPrefix = strings.ToUpper(strings.TrimSpace(input.CategoryPrefix))
	input.Name = strings.TrimSpace(input.Name)
	input.Description = strings.TrimSpace(input.Description)
	if input.AssetTypeID <= 0 || input.Code == "" || input.CategoryPrefix == "" || input.Name == "" {
		return errors.New("tipe aset, kode, prefix, dan nama kategori wajib diisi")
	}
	if !assetPrefixPattern.MatchString(input.CategoryPrefix) {
		return errors.New("prefix kategori hanya boleh berisi huruf dan angka, maksimal 10 karakter")
	}
	if input.ID > 0 {
		return s.Repo.UpdateAssetCategory(input)
	}
	return s.Repo.CreateAssetCategory(input)
}

func (s *AssetService) DeleteAssetCategory(id int64) error {
	if id <= 0 {
		return errors.New("kategori aset tidak valid")
	}
	return s.Repo.DeleteAssetCategory(id)
}

func (s *AssetService) GetComponentTypes() ([]models.ComponentType, error) {
	return s.Repo.GetComponentTypes()
}

func (s *AssetService) SaveComponentType(input models.ComponentTypeInput) error {
	input.Code = strings.ToUpper(strings.TrimSpace(input.Code))
	input.Name = strings.TrimSpace(input.Name)
	input.Description = strings.TrimSpace(input.Description)
	if input.Code == "" || input.Name == "" {
		return errors.New("kode dan nama component type wajib diisi")
	}
	if input.ID > 0 {
		return s.Repo.UpdateComponentType(input)
	}
	return s.Repo.CreateComponentType(input)
}

func (s *AssetService) DeleteComponentType(id int64) error {
	if id <= 0 {
		return errors.New("component type tidak valid")
	}
	return s.Repo.DeleteComponentType(id)
}

func (s *AssetService) GetLocations() ([]models.AssetLocation, error) {
	return s.Repo.GetLocations()
}

func (s *AssetService) SaveLocation(input models.AssetLocationInput) error {
	input.LocationCode = strings.ToUpper(strings.TrimSpace(input.LocationCode))
	input.LocationName = strings.TrimSpace(input.LocationName)
	input.LocationType = strings.ToUpper(strings.TrimSpace(input.LocationType))
	if input.LocationType == "" {
		input.LocationType = "STORE"
	}
	if input.LocationCode == "" || input.LocationName == "" {
		return errors.New("kode dan nama lokasi wajib diisi")
	}
	if input.ID > 0 && input.ParentID == input.ID {
		return errors.New("parent location tidak boleh sama dengan lokasi yang diedit")
	}
	if !allowedAssetLocationType(input.LocationType) {
		return errors.New("tipe lokasi tidak valid")
	}
	if input.ID > 0 {
		return s.Repo.UpdateLocation(input)
	}
	return s.Repo.CreateLocation(input)
}

func (s *AssetService) DeleteLocation(id int64) error {
	if id <= 0 {
		return errors.New("lokasi asset tidak valid")
	}
	return s.Repo.DeleteLocation(id)
}

func (s *AssetService) GetAssets() ([]models.Asset, error) {
	return s.Repo.GetAssets()
}

func (s *AssetService) GetAssetsPage(page, pageSize int, scope models.AccessScope) ([]models.Asset, int, map[string]int, error) {
	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 200 {
		return nil, 0, nil, errors.New("ukuran halaman aset tidak valid")
	}
	total, err := s.Repo.CountAssets(scope)
	if err != nil {
		return nil, 0, nil, err
	}
	totalPages := 1
	if total > 0 {
		totalPages = (total + pageSize - 1) / pageSize
	}
	if page > totalPages {
		page = totalPages
	}
	items, err := s.Repo.GetAssetsPage(pageSize, (page-1)*pageSize, scope)
	if err != nil {
		return nil, 0, nil, err
	}
	counts, err := s.Repo.GetAssetStatusCounts(scope)
	if err != nil {
		return nil, 0, nil, err
	}
	return items, total, counts, nil
}

func (s *AssetService) GetAssetByID(id int64, scope models.AccessScope) (*models.Asset, error) {
	if id <= 0 {
		return nil, errors.New("asset tidak valid")
	}
	return s.Repo.GetAssetByID(id, scope)
}

func (s *AssetService) GetComponentsByAssetID(assetID int64) ([]models.AssetComponent, error) {
	if assetID <= 0 {
		return nil, errors.New("asset tidak valid")
	}
	return s.Repo.GetComponentsByAssetID(assetID)
}

func (s *AssetService) GetAssetMovementsByAssetID(assetID int64, limit int) ([]models.AssetMovement, error) {
	if assetID <= 0 {
		return nil, errors.New("asset tidak valid")
	}
	return s.Repo.GetAssetMovementsByAssetID(assetID, limit)
}

func (s *AssetService) SaveAsset(input models.AssetInput, scopes ...models.AccessScope) error {
	input.AssetName = strings.TrimSpace(input.AssetName)
	input.SerialNumber = strings.TrimSpace(input.SerialNumber)
	input.AssignedPersonNIP = strings.TrimSpace(input.AssignedPersonNIP)
	input.AssignedPersonName = strings.TrimSpace(input.AssignedPersonName)
	input.AssignedPersonDepartment = strings.TrimSpace(input.AssignedPersonDepartment)
	input.Status = strings.ToUpper(strings.TrimSpace(input.Status))
	input.Notes = strings.TrimSpace(input.Notes)
	if input.Status == "" {
		input.Status = models.AssetStatusAvailable
	}
	if input.AssetName == "" || input.AssetTypeID <= 0 || input.AssetCategoryID <= 0 {
		return errors.New("nama, tipe, dan kategori aset wajib diisi")
	}
	if input.AcquisitionDate == "" {
		return errors.New("tanggal perolehan wajib diisi untuk membuat kode aset")
	}
	if !allowedAssetStatus(input.Status) {
		return errors.New("status asset tidak valid")
	}
	if _, err := time.Parse("2006-01-02", input.AcquisitionDate); err != nil {
		return errors.New("tanggal perolehan tidak valid")
	}
	if input.AcquisitionValue.IsNegative() {
		return errors.New("nilai perolehan tidak valid")
	}
	if input.Status == models.AssetStatusDisposed {
		return errors.New("status disposed hanya dapat dibuat melalui proses disposal")
	}
	validCategory, err := s.Repo.AssetCategoryBelongsToType(input.AssetCategoryID, input.AssetTypeID)
	if err != nil {
		return err
	}
	if !validCategory {
		return errors.New("kategori aset tidak sesuai dengan tipe aset")
	}
	validLocation, err := s.Repo.LocationBelongsToStore(input.LocationID, input.StoreID)
	if err != nil {
		return err
	}
	if !validLocation {
		return errors.New("lokasi aset tidak sesuai dengan store yang dipilih")
	}
	if len(scopes) > 0 && !scopes[0].AllowsStore(input.StoreID) {
		return errors.New("anda tidak memiliki akses ke store aset")
	}
	if input.ID > 0 {
		if len(scopes) > 0 {
			if _, err := s.Repo.GetAssetByID(input.ID, scopes[0]); err != nil {
				return errors.New("aset tidak ditemukan atau berada di luar akses store")
			}
		}
		return s.Repo.UpdateAsset(input)
	}
	return s.Repo.CreateAsset(input)
}

func (s *AssetService) DeleteAsset(id int64, scopes ...models.AccessScope) error {
	if id <= 0 {
		return errors.New("asset tidak valid")
	}
	if len(scopes) > 0 {
		if _, err := s.Repo.GetAssetByID(id, scopes[0]); err != nil {
			return errors.New("aset tidak ditemukan atau berada di luar akses store")
		}
	}
	return s.Repo.DeleteAsset(id)
}

func (s *AssetService) GetComponents(scope models.AccessScope) ([]models.AssetComponent, error) {
	return s.Repo.GetComponents(scope)
}

func (s *AssetService) GetComponentByID(id int64, scope models.AccessScope) (*models.AssetComponent, error) {
	if id <= 0 {
		return nil, errors.New("komponen tidak valid")
	}
	return s.Repo.GetComponentByID(id, scope)
}

func (s *AssetService) SaveComponent(input models.AssetComponentInput, scopes ...models.AccessScope) error {
	input.ComponentCode = strings.ToUpper(strings.TrimSpace(input.ComponentCode))
	input.ComponentName = strings.TrimSpace(input.ComponentName)
	input.Brand = strings.TrimSpace(input.Brand)
	input.Model = strings.TrimSpace(input.Model)
	input.Specification = strings.TrimSpace(input.Specification)
	input.SerialNumber = strings.TrimSpace(input.SerialNumber)
	input.Status = strings.ToUpper(strings.TrimSpace(input.Status))
	input.Notes = strings.TrimSpace(input.Notes)
	if input.Status == "" {
		input.Status = models.ComponentStatusInStorage
	}
	if input.ComponentCode == "" || input.ComponentName == "" || input.ComponentTypeID <= 0 {
		return errors.New("kode, nama, dan tipe komponen wajib diisi")
	}
	if !allowedComponentStatus(input.Status) {
		return errors.New("status komponen tidak valid")
	}
	if input.AcquisitionValue.IsNegative() {
		return errors.New("nilai perolehan komponen tidak valid")
	}
	if input.AcquisitionDate != "" {
		if _, err := time.Parse("2006-01-02", input.AcquisitionDate); err != nil {
			return errors.New("tanggal perolehan komponen tidak valid")
		}
	}
	if input.Status == models.ComponentStatusDisposed {
		return errors.New("status disposed hanya dapat dibuat melalui proses disposal")
	}
	if len(scopes) > 0 {
		if input.ID > 0 {
			if _, err := s.Repo.GetComponentByID(input.ID, scopes[0]); err != nil {
				return errors.New("komponen tidak ditemukan atau berada di luar akses store")
			}
		}
		if input.ParentAssetID > 0 {
			if _, err := s.Repo.GetAssetByID(input.ParentAssetID, scopes[0]); err != nil {
				return errors.New("aset induk tidak ditemukan atau berada di luar akses store")
			}
		} else if input.LocationID > 0 {
			storeID, err := s.Repo.GetLocationStoreID(input.LocationID)
			if err != nil {
				return err
			}
			if !scopes[0].AllowsStore(storeID) {
				return errors.New("lokasi komponen berada di luar akses store")
			}
		}
	}
	if input.ID > 0 {
		return s.Repo.UpdateComponent(input)
	}
	return s.Repo.CreateComponent(input)
}

func (s *AssetService) DeleteComponent(id int64, scopes ...models.AccessScope) error {
	if id <= 0 {
		return errors.New("komponen tidak valid")
	}
	if len(scopes) > 0 {
		if _, err := s.Repo.GetComponentByID(id, scopes[0]); err != nil {
			return errors.New("komponen tidak ditemukan atau berada di luar akses store")
		}
	}
	return s.Repo.DeleteComponent(id)
}

func (s *AssetService) GetAssetMovements(scope models.AccessScope) ([]models.AssetMovement, error) {
	return s.Repo.GetAssetMovements(scope)
}

func (s *AssetService) CreateAssetMovement(input models.AssetMovementInput, scopes ...models.AccessScope) error {
	input.MovementType = strings.ToUpper(strings.TrimSpace(input.MovementType))
	input.Notes = strings.TrimSpace(input.Notes)
	if input.AssetID <= 0 {
		return errors.New("asset wajib dipilih")
	}
	if !allowedAssetMovementType(input.MovementType) {
		return errors.New("tipe movement asset tidak valid")
	}
	var current *models.Asset
	var err error
	if len(scopes) > 0 {
		current, err = s.Repo.GetAssetByID(input.AssetID, scopes[0])
		if err != nil {
			return errors.New("aset tidak ditemukan atau berada di luar akses store")
		}
	} else {
		current, err = s.Repo.GetAssetByID(input.AssetID)
		if err != nil {
			return errors.New("aset tidak ditemukan")
		}
	}
	if input.ToStoreID <= 0 {
		input.ToStoreID = current.StoreID
	}
	if input.ToLocationID <= 0 {
		input.ToLocationID = current.LocationID
	}
	if len(scopes) > 0 {
		if !scopes[0].AllowsStore(input.ToStoreID) {
			return errors.New("anda tidak memiliki akses ke store tujuan")
		}
	}
	validLocation, err := s.Repo.LocationBelongsToStore(input.ToLocationID, input.ToStoreID)
	if err != nil {
		return err
	}
	if !validLocation {
		return errors.New("lokasi tujuan tidak sesuai dengan store tujuan")
	}
	input.ResultStatus = nextAssetStatus(input.MovementType, input.ToUserID)
	return s.Repo.CreateAssetMovement(input)
}

func (s *AssetService) GetComponentMovements(scope models.AccessScope) ([]models.AssetComponentMovement, error) {
	return s.Repo.GetComponentMovements(scope)
}

func (s *AssetService) GetComponentMovementsByComponentID(componentID int64) ([]models.AssetComponentMovement, error) {
	if componentID <= 0 {
		return nil, errors.New("komponen tidak valid")
	}
	return s.Repo.GetComponentMovementsByComponentID(componentID)
}

func (s *AssetService) CreateComponentMovement(input models.AssetComponentMovementInput, scopes ...models.AccessScope) error {
	input.MovementType = strings.ToUpper(strings.TrimSpace(input.MovementType))
	input.Notes = strings.TrimSpace(input.Notes)
	if input.ComponentID <= 0 {
		return errors.New("komponen wajib dipilih")
	}
	if !allowedComponentMovementType(input.MovementType) {
		return errors.New("tipe movement komponen tidak valid")
	}
	if input.MovementType == models.ComponentMovementInstall && input.ToAssetID <= 0 {
		return errors.New("asset tujuan wajib dipilih untuk install komponen")
	}
	if len(scopes) > 0 {
		if _, err := s.Repo.GetComponentByID(input.ComponentID, scopes[0]); err != nil {
			return errors.New("komponen tidak ditemukan atau berada di luar akses store")
		}
		if input.ToAssetID > 0 {
			if _, err := s.Repo.GetAssetByID(input.ToAssetID, scopes[0]); err != nil {
				return errors.New("aset tujuan tidak ditemukan atau berada di luar akses store")
			}
		} else if input.ToLocationID > 0 {
			storeID, err := s.Repo.GetLocationStoreID(input.ToLocationID)
			if err != nil {
				return err
			}
			if !scopes[0].AllowsStore(storeID) {
				return errors.New("lokasi tujuan berada di luar akses store")
			}
		}
	}
	input.ResultAssetID = input.ToAssetID
	if input.MovementType == models.ComponentMovementUninstall || input.MovementType == models.ComponentMovementReturnToStorage || input.MovementType == models.ComponentMovementTransfer {
		input.ResultAssetID = 0
	}
	input.ResultStatus = nextComponentStatus(input.MovementType, input.ResultAssetID)
	return s.Repo.CreateComponentMovement(input)
}

func (s *AssetService) GetStoreOptions(scopes ...models.AccessScope) ([]models.Store, error) {
	return s.Repo.GetStoreOptions(scopes...)
}

func (s *AssetService) GetUserOptions() ([]models.UserSelectOption, error) {
	return s.Repo.GetUserOptions()
}

func (s *AssetService) GetAssetOptions(scope models.AccessScope) ([]models.AssetSelectOption, error) {
	return s.Repo.GetAssetOptions(scope)
}

func (s *AssetService) GetComponentOptions(scope models.AccessScope) ([]models.AssetSelectOption, error) {
	return s.Repo.GetComponentOptions(scope)
}

func allowedAssetLocationType(value string) bool {
	switch value {
	case "STORE", "WAREHOUSE", "ROOM", "SERVICE_CENTER", "OTHER":
		return true
	default:
		return false
	}
}

func allowedAssetStatus(value string) bool {
	switch value {
	case models.AssetStatusAvailable, models.AssetStatusInUse, models.AssetStatusMaintenance, models.AssetStatusBroken, models.AssetStatusDisposed:
		return true
	default:
		return false
	}
}

func allowedComponentStatus(value string) bool {
	switch value {
	case models.ComponentStatusInStorage, models.ComponentStatusInstalled, models.ComponentStatusMaintenance, models.ComponentStatusBroken, models.ComponentStatusDisposed:
		return true
	default:
		return false
	}
}

func allowedAssetMovementType(value string) bool {
	switch value {
	case models.AssetMovementReceive, models.AssetMovementTransfer, models.AssetMovementAssign, models.AssetMovementReturn, models.AssetMovementMaintenance, models.AssetMovementBroken:
		return true
	default:
		return false
	}
}

func allowedComponentMovementType(value string) bool {
	switch value {
	case models.ComponentMovementReceive, models.ComponentMovementInstall, models.ComponentMovementUninstall, models.ComponentMovementTransfer, models.ComponentMovementMaintenance, models.ComponentMovementReturnToStorage, models.ComponentMovementBroken, models.ComponentMovementDispose:
		return true
	default:
		return false
	}
}

func nextAssetStatus(movementType string, toUserID int) string {
	switch movementType {
	case models.AssetMovementMaintenance:
		return models.AssetStatusMaintenance
	case models.AssetMovementBroken:
		return models.AssetStatusBroken
	case models.AssetMovementReturn:
		return models.AssetStatusAvailable
	default:
		if toUserID > 0 || movementType == models.AssetMovementAssign || movementType == models.AssetMovementTransfer || movementType == models.AssetMovementReceive {
			return models.AssetStatusInUse
		}
		return models.AssetStatusAvailable
	}
}

func nextComponentStatus(movementType string, toAssetID int64) string {
	switch movementType {
	case models.ComponentMovementInstall:
		return models.ComponentStatusInstalled
	case models.ComponentMovementMaintenance:
		return models.ComponentStatusMaintenance
	case models.ComponentMovementBroken:
		return models.ComponentStatusBroken
	case models.ComponentMovementDispose:
		return models.ComponentStatusDisposed
	default:
		if toAssetID > 0 {
			return models.ComponentStatusInstalled
		}
		return models.ComponentStatusInStorage
	}
}
