package services

import (
	"errors"
	"gobase-app/models"
	"strings"
)

func (s *CatalogService) GetBrands() ([]models.CatalogBrand, error) {
	return s.Repo.GetBrands()
}

func (s *CatalogService) SaveBrand(input models.CatalogBrandInput) error {
	input.Code = strings.ToUpper(strings.TrimSpace(input.Code))
	input.Name = strings.TrimSpace(input.Name)
	input.Description = strings.TrimSpace(input.Description)
	if input.Code == "" || input.Name == "" {
		return errors.New("kode dan nama merek wajib diisi")
	}
	if !catalogCodePattern.MatchString(input.Code) {
		return errors.New("kode merek hanya boleh berisi huruf, angka, tanda hubung, atau underscore")
	}
	if input.ID > 0 {
		return s.Repo.UpdateBrand(input)
	}
	return s.Repo.CreateBrand(input)
}

func (s *CatalogService) DeleteBrand(id int64) error {
	if id <= 0 {
		return errors.New("merek tidak valid")
	}
	return s.Repo.DeleteBrand(id)
}

func (s *CatalogService) GetItemVariants() ([]models.CatalogItemVariant, error) {
	return s.Repo.GetItemVariants()
}

func (s *CatalogService) SaveItemVariant(input models.CatalogItemVariantInput) error {
	input.VariantCode = strings.ToUpper(strings.TrimSpace(input.VariantCode))
	input.ModelName = strings.TrimSpace(input.ModelName)
	input.ManufacturerSKU = strings.TrimSpace(input.ManufacturerSKU)
	input.Specification = strings.TrimSpace(input.Specification)
	if input.ItemID <= 0 || input.VariantCode == "" || input.ModelName == "" {
		return errors.New("item, kode varian, dan model wajib diisi")
	}
	if !catalogCodePattern.MatchString(input.VariantCode) {
		return errors.New("kode varian hanya boleh berisi huruf, angka, tanda hubung, atau underscore")
	}
	itemTypeCode, err := s.Repo.GetItemTypeCode(input.ItemID)
	if err != nil {
		return err
	}
	if itemTypeCode == "COMPOSITE" {
		return errors.New("item rakitan menggunakan konfigurasi BOM dan tidak memiliki varian produk langsung")
	}
	if input.ID > 0 {
		return s.Repo.UpdateItemVariant(input)
	}
	return s.Repo.CreateItemVariant(input)
}

func (s *CatalogService) DeleteItemVariant(id int64) error {
	if id <= 0 {
		return errors.New("varian tidak valid")
	}
	return s.Repo.DeleteItemVariant(id)
}

func (s *CatalogService) GetItemBOMs() ([]models.CatalogItemBOM, error) {
	return s.Repo.GetItemBOMs()
}

func (s *CatalogService) SaveItemBOM(input models.CatalogItemBOMInput) error {
	input.BOMCode = strings.ToUpper(strings.TrimSpace(input.BOMCode))
	input.Name = strings.TrimSpace(input.Name)
	input.Description = strings.TrimSpace(input.Description)
	if input.ParentItemID <= 0 || input.BOMCode == "" || input.Name == "" {
		return errors.New("item parent, kode BOM, dan nama konfigurasi wajib diisi")
	}
	if !catalogCodePattern.MatchString(input.BOMCode) {
		return errors.New("kode BOM hanya boleh berisi huruf, angka, tanda hubung, atau underscore")
	}
	if input.VersionNo <= 0 {
		return errors.New("versi BOM harus lebih besar dari nol")
	}
	itemTypeCode, err := s.Repo.GetItemTypeCode(input.ParentItemID)
	if err != nil {
		return err
	}
	if itemTypeCode != "COMPOSITE" {
		return errors.New("item parent BOM harus memiliki jenis Rakitan/COMPOSITE")
	}
	if input.ID > 0 {
		return s.Repo.UpdateItemBOM(input)
	}
	return s.Repo.CreateItemBOM(input)
}

func (s *CatalogService) DeleteItemBOM(id int64) error {
	if id <= 0 {
		return errors.New("BOM tidak valid")
	}
	return s.Repo.DeleteItemBOM(id)
}

func (s *CatalogService) GetItemBOMItems() ([]models.CatalogItemBOMItem, error) {
	return s.Repo.GetItemBOMItems()
}

func (s *CatalogService) SaveItemBOMItem(input models.CatalogItemBOMItemInput) error {
	input.Notes = strings.TrimSpace(input.Notes)
	if input.BOMID <= 0 || input.ComponentItemID <= 0 ||
		input.VariantID <= 0 || input.PreferredVendorID <= 0 {
		return errors.New("BOM, komponen, varian, dan vendor pilihan wajib diisi")
	}
	if input.Qty <= 0 {
		return errors.New("jumlah komponen harus lebih besar dari nol")
	}
	if input.SortOrder < 0 {
		return errors.New("urutan tidak boleh negatif")
	}
	if err := s.Repo.ValidateBOMItemSelection(input); err != nil {
		return err
	}
	if input.ID > 0 {
		return s.Repo.UpdateItemBOMItem(input)
	}
	return s.Repo.CreateItemBOMItem(input)
}

func (s *CatalogService) DeleteItemBOMItem(id int64) error {
	if id <= 0 {
		return errors.New("komponen BOM tidak valid")
	}
	return s.Repo.DeleteItemBOMItem(id)
}
