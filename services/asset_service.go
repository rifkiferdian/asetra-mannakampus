package services

import (
	"errors"
	"gobase-app/models"
	"gobase-app/repositories"
	"strings"
)

type AssetService struct {
	Repo *repositories.AssetRepository
}

func (s *AssetService) GetAssetTypes() ([]models.AssetType, error) {
	return s.Repo.GetAssetTypes()
}

func (s *AssetService) SaveAssetType(input models.AssetTypeInput) error {
	input.Code = strings.ToUpper(strings.TrimSpace(input.Code))
	input.Name = strings.TrimSpace(input.Name)
	input.Description = strings.TrimSpace(input.Description)
	if input.Code == "" || input.Name == "" {
		return errors.New("kode dan nama asset type wajib diisi")
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

func (s *AssetService) GetAssetByID(id int64) (*models.Asset, error) {
	if id <= 0 {
		return nil, errors.New("asset tidak valid")
	}
	return s.Repo.GetAssetByID(id)
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

func (s *AssetService) SaveAsset(input models.AssetInput) error {
	input.AssetCode = strings.ToUpper(strings.TrimSpace(input.AssetCode))
	input.AssetName = strings.TrimSpace(input.AssetName)
	input.SerialNumber = strings.TrimSpace(input.SerialNumber)
	input.AssignedPersonNIP = strings.TrimSpace(input.AssignedPersonNIP)
	input.AssignedPersonName = strings.TrimSpace(input.AssignedPersonName)
	input.AssignedPersonDepartment = strings.TrimSpace(input.AssignedPersonDepartment)
	input.Status = strings.ToUpper(strings.TrimSpace(input.Status))
	input.Notes = strings.TrimSpace(input.Notes)
	if input.Status == "" {
		input.Status = "AVAILABLE"
	}
	if input.AssetCode == "" || input.AssetName == "" || input.AssetTypeID <= 0 {
		return errors.New("kode, nama, dan tipe asset wajib diisi")
	}
	if !allowedAssetStatus(input.Status) {
		return errors.New("status asset tidak valid")
	}
	if input.ID > 0 {
		return s.Repo.UpdateAsset(input)
	}
	return s.Repo.CreateAsset(input)
}

func (s *AssetService) DeleteAsset(id int64) error {
	if id <= 0 {
		return errors.New("asset tidak valid")
	}
	return s.Repo.DeleteAsset(id)
}

func (s *AssetService) GetComponents() ([]models.AssetComponent, error) {
	return s.Repo.GetComponents()
}

func (s *AssetService) SaveComponent(input models.AssetComponentInput) error {
	input.ComponentCode = strings.ToUpper(strings.TrimSpace(input.ComponentCode))
	input.ComponentName = strings.TrimSpace(input.ComponentName)
	input.Brand = strings.TrimSpace(input.Brand)
	input.Model = strings.TrimSpace(input.Model)
	input.Specification = strings.TrimSpace(input.Specification)
	input.SerialNumber = strings.TrimSpace(input.SerialNumber)
	input.Status = strings.ToUpper(strings.TrimSpace(input.Status))
	input.Notes = strings.TrimSpace(input.Notes)
	if input.Status == "" {
		input.Status = "IN_STORAGE"
	}
	if input.ComponentCode == "" || input.ComponentName == "" || input.ComponentTypeID <= 0 {
		return errors.New("kode, nama, dan tipe komponen wajib diisi")
	}
	if !allowedComponentStatus(input.Status) {
		return errors.New("status komponen tidak valid")
	}
	if input.ID > 0 {
		return s.Repo.UpdateComponent(input)
	}
	return s.Repo.CreateComponent(input)
}

func (s *AssetService) DeleteComponent(id int64) error {
	if id <= 0 {
		return errors.New("komponen tidak valid")
	}
	return s.Repo.DeleteComponent(id)
}

func (s *AssetService) GetAssetMovements() ([]models.AssetMovement, error) {
	return s.Repo.GetAssetMovements()
}

func (s *AssetService) CreateAssetMovement(input models.AssetMovementInput) error {
	input.MovementType = strings.ToUpper(strings.TrimSpace(input.MovementType))
	input.Notes = strings.TrimSpace(input.Notes)
	if input.AssetID <= 0 {
		return errors.New("asset wajib dipilih")
	}
	if !allowedAssetMovementType(input.MovementType) {
		return errors.New("tipe movement asset tidak valid")
	}
	return s.Repo.CreateAssetMovement(input)
}

func (s *AssetService) GetComponentMovements() ([]models.AssetComponentMovement, error) {
	return s.Repo.GetComponentMovements()
}

func (s *AssetService) CreateComponentMovement(input models.AssetComponentMovementInput) error {
	input.MovementType = strings.ToUpper(strings.TrimSpace(input.MovementType))
	input.Notes = strings.TrimSpace(input.Notes)
	if input.ComponentID <= 0 {
		return errors.New("komponen wajib dipilih")
	}
	if !allowedComponentMovementType(input.MovementType) {
		return errors.New("tipe movement komponen tidak valid")
	}
	if input.MovementType == "INSTALL" && input.ToAssetID <= 0 {
		return errors.New("asset tujuan wajib dipilih untuk install komponen")
	}
	return s.Repo.CreateComponentMovement(input)
}

func (s *AssetService) GetStoreOptions() ([]models.Store, error) {
	return s.Repo.GetStoreOptions()
}

func (s *AssetService) GetUserOptions() ([]models.UserSelectOption, error) {
	return s.Repo.GetUserOptions()
}

func (s *AssetService) GetAssetOptions() ([]models.AssetSelectOption, error) {
	return s.Repo.GetAssetOptions()
}

func (s *AssetService) GetComponentOptions() ([]models.AssetSelectOption, error) {
	return s.Repo.GetComponentOptions()
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
	case "AVAILABLE", "IN_USE", "MAINTENANCE", "BROKEN", "DISPOSED":
		return true
	default:
		return false
	}
}

func allowedComponentStatus(value string) bool {
	switch value {
	case "IN_STORAGE", "INSTALLED", "MAINTENANCE", "BROKEN", "DISPOSED":
		return true
	default:
		return false
	}
}

func allowedAssetMovementType(value string) bool {
	switch value {
	case "RECEIVE", "TRANSFER", "ASSIGN", "RETURN", "MAINTENANCE", "BROKEN", "DISPOSE":
		return true
	default:
		return false
	}
}

func allowedComponentMovementType(value string) bool {
	switch value {
	case "RECEIVE", "INSTALL", "UNINSTALL", "TRANSFER", "MAINTENANCE", "RETURN_TO_STORAGE", "BROKEN", "DISPOSE":
		return true
	default:
		return false
	}
}
