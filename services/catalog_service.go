package services

import (
	"errors"
	"gobase-app/models"
	"gobase-app/repositories"
	"regexp"
	"strings"
	"time"
)

var catalogCodePattern = regexp.MustCompile(`^[A-Z0-9][A-Z0-9_-]{0,49}$`)
var catalogCurrencyPattern = regexp.MustCompile(`^[A-Z]{3}$`)

type CatalogService struct {
	Repo *repositories.CatalogRepository
}

func (s *CatalogService) GetItemTypes() ([]models.CatalogItemType, error) {
	return s.Repo.GetItemTypes()
}

func (s *CatalogService) SaveItemType(input models.CatalogItemTypeInput) error {
	input.Code = strings.ToUpper(strings.TrimSpace(input.Code))
	input.Name = strings.TrimSpace(input.Name)
	input.Description = strings.TrimSpace(input.Description)
	if input.Code == "" || input.Name == "" {
		return errors.New("kode dan nama jenis item wajib diisi")
	}
	if len(input.Code) > 30 || !catalogCodePattern.MatchString(input.Code) {
		return errors.New("kode jenis item maksimal 30 karakter dan hanya boleh berisi huruf, angka, tanda hubung, atau underscore")
	}
	if input.ID > 0 {
		return s.Repo.UpdateItemType(input)
	}
	return s.Repo.CreateItemType(input)
}

func (s *CatalogService) DeleteItemType(id int64) error {
	if id <= 0 {
		return errors.New("jenis item tidak valid")
	}
	return s.Repo.DeleteItemType(id)
}

func (s *CatalogService) GetCategories() ([]models.CatalogItemCategory, error) {
	return s.Repo.GetCategories()
}

func (s *CatalogService) SaveCategory(input models.CatalogItemCategoryInput) error {
	input.Code = strings.ToUpper(strings.TrimSpace(input.Code))
	input.Name = strings.TrimSpace(input.Name)
	input.Description = strings.TrimSpace(input.Description)
	if input.Code == "" || input.Name == "" {
		return errors.New("kode dan nama kategori wajib diisi")
	}
	if !catalogCodePattern.MatchString(input.Code) {
		return errors.New("kode kategori hanya boleh berisi huruf, angka, tanda hubung, atau underscore")
	}
	if input.ID > 0 && input.ParentID == input.ID {
		return errors.New("parent kategori tidak boleh sama dengan kategori yang diedit")
	}
	if input.ID > 0 {
		return s.Repo.UpdateCategory(input)
	}
	return s.Repo.CreateCategory(input)
}

func (s *CatalogService) DeleteCategory(id int64) error {
	if id <= 0 {
		return errors.New("kategori tidak valid")
	}
	return s.Repo.DeleteCategory(id)
}

func (s *CatalogService) GetItems() ([]models.CatalogItem, error) {
	return s.Repo.GetItems()
}

func (s *CatalogService) SaveItem(input models.CatalogItemInput) error {
	input.ItemCode = strings.ToUpper(strings.TrimSpace(input.ItemCode))
	input.Name = strings.TrimSpace(input.Name)
	input.UOM = strings.TrimSpace(input.UOM)
	input.Description = strings.TrimSpace(input.Description)
	if input.ItemCode == "" || input.Name == "" || input.CategoryID <= 0 || input.ItemTypeID <= 0 || input.UOM == "" {
		return errors.New("kode, nama, kategori, jenis, dan satuan item wajib diisi")
	}
	if !catalogCodePattern.MatchString(input.ItemCode) {
		return errors.New("kode item hanya boleh berisi huruf, angka, tanda hubung, atau underscore")
	}
	if input.IsAssetCandidate && input.AssetTypeID <= 0 {
		return errors.New("asset type wajib dipilih untuk item kandidat aset")
	}
	if !input.IsAssetCandidate {
		input.AssetTypeID = 0
	}
	itemTypeCode, err := s.Repo.GetItemTypeCodeByID(input.ItemTypeID)
	if err != nil {
		return err
	}
	if itemTypeCode == "COMPOSITE" {
		input.IsPurchasable = false
	} else if !input.IsPurchasable {
		return errors.New("item barang atau jasa harus dapat dibeli langsung")
	}
	if input.ID > 0 {
		return s.Repo.UpdateItem(input)
	}
	return s.Repo.CreateItem(input)
}

func (s *CatalogService) DeleteItem(id int64) error {
	if id <= 0 {
		return errors.New("item tidak valid")
	}
	return s.Repo.DeleteItem(id)
}

func (s *CatalogService) GetDivisions() ([]models.Division, error) {
	return s.Repo.GetDivisions()
}

func (s *CatalogService) GetAssetTypes() ([]models.AssetType, error) {
	return s.Repo.GetAssetTypes()
}

func (s *CatalogService) GetComponentTypes() ([]models.ComponentType, error) {
	return s.Repo.GetComponentTypes()
}

func (s *CatalogService) GetPackages() ([]models.CatalogPackage, error) {
	return s.Repo.GetPackages()
}

func (s *CatalogService) SavePackage(input models.CatalogPackageInput) error {
	input.PackageCode = strings.ToUpper(strings.TrimSpace(input.PackageCode))
	input.Name = strings.TrimSpace(input.Name)
	input.Description = strings.TrimSpace(input.Description)
	if input.PackageCode == "" || input.Name == "" {
		return errors.New("kode dan nama paket wajib diisi")
	}
	if !catalogCodePattern.MatchString(input.PackageCode) {
		return errors.New("kode paket hanya boleh berisi huruf, angka, tanda hubung, atau underscore")
	}
	if input.ID > 0 {
		return s.Repo.UpdatePackage(input)
	}
	return s.Repo.CreatePackage(input)
}

func (s *CatalogService) DeletePackage(id int64) error {
	if id <= 0 {
		return errors.New("paket tidak valid")
	}
	return s.Repo.DeletePackage(id)
}

func (s *CatalogService) GetPackageItems() ([]models.CatalogPackageItem, error) {
	return s.Repo.GetPackageItems()
}

func (s *CatalogService) SavePackageItem(input models.CatalogPackageItemInput) error {
	input.Notes = strings.TrimSpace(input.Notes)
	if input.PackageID <= 0 || input.ItemID <= 0 {
		return errors.New("paket dan item wajib dipilih")
	}
	if input.Qty <= 0 {
		return errors.New("jumlah item harus lebih besar dari nol")
	}
	if input.SortOrder < 0 {
		return errors.New("urutan tidak boleh negatif")
	}
	itemTypeCode, err := s.Repo.GetItemTypeCode(input.ItemID)
	if err != nil {
		return err
	}
	if itemTypeCode == "COMPOSITE" {
		if input.BOMID <= 0 {
			return errors.New("konfigurasi BOM wajib dipilih untuk item rakitan")
		}
		ok, err := s.Repo.BOMBelongsToItem(input.BOMID, input.ItemID)
		if err != nil {
			return err
		}
		if !ok {
			return errors.New("BOM tidak sesuai dengan item rakitan")
		}
		input.VariantID = 0
		input.PreferredVendorID = 0
	} else {
		if input.VariantID <= 0 {
			return errors.New("varian wajib dipilih untuk item yang dibeli langsung")
		}
		if input.PreferredVendorID <= 0 {
			return errors.New("vendor pilihan wajib dipilih untuk item yang dibeli langsung")
		}
		ok, err := s.Repo.VariantBelongsToItem(input.VariantID, input.ItemID)
		if err != nil {
			return err
		}
		if !ok {
			return errors.New("varian tidak sesuai dengan item paket")
		}
		input.BOMID = 0
	}
	if input.ID > 0 {
		return s.Repo.UpdatePackageItem(input)
	}
	return s.Repo.CreatePackageItem(input)
}

func (s *CatalogService) DeletePackageItem(id int64) error {
	if id <= 0 {
		return errors.New("isi paket tidak valid")
	}
	return s.Repo.DeletePackageItem(id)
}

func (s *CatalogService) GetVendorItemPrices() ([]models.VendorItemPrice, error) {
	return s.Repo.GetVendorItemPrices()
}

func (s *CatalogService) GetVendors() ([]models.Vendor, error) {
	return s.Repo.GetVendors()
}

func (s *CatalogService) SaveVendorItemPrice(input models.VendorItemPriceInput) error {
	input.CurrencyCode = strings.ToUpper(strings.TrimSpace(input.CurrencyCode))
	input.ValidFrom = strings.TrimSpace(input.ValidFrom)
	input.ValidUntil = strings.TrimSpace(input.ValidUntil)
	input.QuotationReference = strings.TrimSpace(input.QuotationReference)
	input.Notes = strings.TrimSpace(input.Notes)
	if input.VendorID <= 0 || input.VariantID <= 0 {
		return errors.New("vendor dan varian item wajib dipilih")
	}
	if input.UnitPrice <= 0 {
		return errors.New("harga satuan harus lebih besar dari nol")
	}
	if input.MinimumQty <= 0 {
		return errors.New("minimum pembelian harus lebih besar dari nol")
	}
	if !catalogCurrencyPattern.MatchString(input.CurrencyCode) {
		return errors.New("kode mata uang wajib terdiri dari tiga huruf, contoh IDR")
	}
	validFrom, err := time.Parse("2006-01-02", input.ValidFrom)
	if err != nil {
		return errors.New("tanggal mulai berlaku wajib diisi dengan benar")
	}
	if input.ValidUntil != "" {
		validUntil, err := time.Parse("2006-01-02", input.ValidUntil)
		if err != nil {
			return errors.New("tanggal akhir berlaku tidak valid")
		}
		if validUntil.Before(validFrom) {
			return errors.New("tanggal akhir berlaku tidak boleh sebelum tanggal mulai")
		}
	}
	if input.LeadTimeDays < 0 {
		return errors.New("lead time tidak boleh negatif")
	}
	if input.ID > 0 {
		return s.Repo.UpdateVendorItemPrice(input)
	}
	return s.Repo.CreateVendorItemPrice(input)
}

func (s *CatalogService) DeleteVendorItemPrice(id int64) error {
	if id <= 0 {
		return errors.New("harga vendor tidak valid")
	}
	return s.Repo.DeleteVendorItemPrice(id)
}
