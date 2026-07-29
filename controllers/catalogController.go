package controllers

import (
	"fmt"
	"gobase-app/config"
	"gobase-app/models"
	"gobase-app/repositories"
	"gobase-app/services"
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
)

func CatalogItemTypeIndex(c *gin.Context) {
	renderCatalogItemTypePage(c, catalogService(), "")
}

func CatalogItemTypeStore(c *gin.Context) {
	if err := catalogService().SaveItemType(bindCatalogItemTypeInput(c)); err != nil {
		renderCatalogItemTypePage(c, catalogService(), err.Error())
		return
	}
	c.Redirect(http.StatusSeeOther, "/catalog/item-types")
}

func CatalogItemTypeUpdate(c *gin.Context) {
	input := bindCatalogItemTypeInput(c)
	input.ID = parseInt64Form(c, "id")
	if err := catalogService().SaveItemType(input); err != nil {
		renderCatalogItemTypePage(c, catalogService(), err.Error())
		return
	}
	c.Redirect(http.StatusSeeOther, "/catalog/item-types")
}

func CatalogItemTypeDelete(c *gin.Context) {
	if err := catalogService().DeleteItemType(parseInt64Param(c, "id")); err != nil {
		renderCatalogItemTypePage(c, catalogService(), err.Error())
		return
	}
	c.Redirect(http.StatusSeeOther, "/catalog/item-types")
}

func CatalogCategoryIndex(c *gin.Context) {
	renderCatalogCategoryPage(c, catalogService(), "")
}

func CatalogCategoryStore(c *gin.Context) {
	if err := catalogService().SaveCategory(bindCatalogCategoryInput(c)); err != nil {
		renderCatalogCategoryPage(c, catalogService(), err.Error())
		return
	}
	c.Redirect(http.StatusSeeOther, "/catalog/categories")
}

func CatalogCategoryUpdate(c *gin.Context) {
	input := bindCatalogCategoryInput(c)
	input.ID = parseInt64Form(c, "id")
	if err := catalogService().SaveCategory(input); err != nil {
		renderCatalogCategoryPage(c, catalogService(), err.Error())
		return
	}
	c.Redirect(http.StatusSeeOther, "/catalog/categories")
}

func CatalogCategoryDelete(c *gin.Context) {
	if err := catalogService().DeleteCategory(parseInt64Param(c, "id")); err != nil {
		renderCatalogCategoryPage(c, catalogService(), err.Error())
		return
	}
	c.Redirect(http.StatusSeeOther, "/catalog/categories")
}

func CatalogItemIndex(c *gin.Context) {
	renderCatalogItemPage(c, catalogService(), "")
}

func CatalogItemStore(c *gin.Context) {
	if err := catalogService().SaveItem(bindCatalogItemInput(c)); err != nil {
		renderCatalogItemPage(c, catalogService(), err.Error())
		return
	}
	c.Redirect(http.StatusSeeOther, "/catalog/items")
}

func CatalogItemUpdate(c *gin.Context) {
	input := bindCatalogItemInput(c)
	input.ID = parseInt64Form(c, "id")
	if err := catalogService().SaveItem(input); err != nil {
		renderCatalogItemPage(c, catalogService(), err.Error())
		return
	}
	c.Redirect(http.StatusSeeOther, "/catalog/items")
}

func CatalogItemDelete(c *gin.Context) {
	if err := catalogService().DeleteItem(parseInt64Param(c, "id")); err != nil {
		renderCatalogItemPage(c, catalogService(), err.Error())
		return
	}
	c.Redirect(http.StatusSeeOther, "/catalog/items")
}

func CatalogItemDetailIndex(c *gin.Context) {
	renderCatalogItemDetailPage(c, catalogService(), "")
}

func CatalogItemDetailStore(c *gin.Context) {
	if err := catalogService().SaveItemDetail(bindCatalogItemDetailInput(c)); err != nil {
		renderCatalogItemDetailPage(c, catalogService(), err.Error())
		return
	}
	redirectCatalogItemDetails(c)
}

func CatalogItemDetailUpdate(c *gin.Context) {
	input := bindCatalogItemDetailInput(c)
	input.ID = parseInt64Form(c, "id")
	if err := catalogService().SaveItemDetail(input); err != nil {
		renderCatalogItemDetailPage(c, catalogService(), err.Error())
		return
	}
	redirectCatalogItemDetails(c)
}

func CatalogItemDetailDelete(c *gin.Context) {
	if err := catalogService().DeleteItemDetail(parseInt64Param(c, "id")); err != nil {
		renderCatalogItemDetailPage(c, catalogService(), err.Error())
		return
	}
	redirectCatalogItemDetails(c)
}

func CatalogPackageIndex(c *gin.Context) {
	renderCatalogPackagePage(c, catalogService(), "")
}

func CatalogPackageStore(c *gin.Context) {
	if err := catalogService().SavePackage(bindCatalogPackageInput(c)); err != nil {
		renderCatalogPackagePage(c, catalogService(), err.Error())
		return
	}
	c.Redirect(http.StatusSeeOther, "/catalog/packages")
}

func CatalogPackageUpdate(c *gin.Context) {
	input := bindCatalogPackageInput(c)
	input.ID = parseInt64Form(c, "id")
	if err := catalogService().SavePackage(input); err != nil {
		renderCatalogPackagePage(c, catalogService(), err.Error())
		return
	}
	c.Redirect(http.StatusSeeOther, "/catalog/packages")
}

func CatalogPackageDelete(c *gin.Context) {
	if err := catalogService().DeletePackage(parseInt64Param(c, "id")); err != nil {
		renderCatalogPackagePage(c, catalogService(), err.Error())
		return
	}
	c.Redirect(http.StatusSeeOther, "/catalog/packages")
}

func CatalogPackageItemIndex(c *gin.Context) {
	renderCatalogPackageItemPage(c, catalogService(), "")
}

func CatalogPackageItemStore(c *gin.Context) {
	if err := catalogService().SavePackageItem(bindCatalogPackageItemInput(c)); err != nil {
		renderCatalogPackageItemPage(c, catalogService(), err.Error())
		return
	}
	c.Redirect(http.StatusSeeOther, "/catalog/package-items")
}

func CatalogPackageItemUpdate(c *gin.Context) {
	input := bindCatalogPackageItemInput(c)
	input.ID = parseInt64Form(c, "id")
	if err := catalogService().SavePackageItem(input); err != nil {
		renderCatalogPackageItemPage(c, catalogService(), err.Error())
		return
	}
	c.Redirect(http.StatusSeeOther, "/catalog/package-items")
}

func CatalogPackageItemDelete(c *gin.Context) {
	if err := catalogService().DeletePackageItem(parseInt64Param(c, "id")); err != nil {
		renderCatalogPackageItemPage(c, catalogService(), err.Error())
		return
	}
	c.Redirect(http.StatusSeeOther, "/catalog/package-items")
}

func VendorItemPriceIndex(c *gin.Context) {
	renderVendorItemPricePage(c, catalogService(), "")
}

func VendorItemPriceStore(c *gin.Context) {
	if err := catalogService().SaveVendorItemPrice(bindVendorItemPriceInput(c)); err != nil {
		renderVendorItemPricePage(c, catalogService(), err.Error())
		return
	}
	c.Redirect(http.StatusSeeOther, "/catalog/vendor-item-prices")
}

func VendorItemPriceUpdate(c *gin.Context) {
	input := bindVendorItemPriceInput(c)
	input.ID = parseInt64Form(c, "id")
	if err := catalogService().SaveVendorItemPrice(input); err != nil {
		renderVendorItemPricePage(c, catalogService(), err.Error())
		return
	}
	c.Redirect(http.StatusSeeOther, "/catalog/vendor-item-prices")
}

func VendorItemPriceDelete(c *gin.Context) {
	if err := catalogService().DeleteVendorItemPrice(parseInt64Param(c, "id")); err != nil {
		renderVendorItemPricePage(c, catalogService(), err.Error())
		return
	}
	c.Redirect(http.StatusSeeOther, "/catalog/vendor-item-prices")
}

func renderCatalogItemTypePage(c *gin.Context, service *services.CatalogService, message string) {
	items, err := service.GetItemTypes()
	if err != nil {
		c.String(http.StatusInternalServerError, err.Error())
		return
	}
	activeCount, totalItems, unusedCount := 0, 0, 0
	for _, item := range items {
		if item.IsActive {
			activeCount++
		}
		if item.ItemCount == 0 {
			unusedCount++
		}
		totalItems += item.ItemCount
	}
	pageItems, pagination := paginateAssetSlice(c, items)
	Render(c, "catalog_item_type.html", gin.H{
		"Title": "Jenis Item", "Page": "catalog_item_type", "Items": pageItems,
		"Pagination": pagination, "Error": message, "TotalTypes": len(items),
		"ActiveCount": activeCount, "TotalItems": totalItems, "UnusedCount": unusedCount,
	})
}

func renderCatalogCategoryPage(c *gin.Context, service *services.CatalogService, message string) {
	items, err := service.GetCategories()
	if err != nil {
		c.String(http.StatusInternalServerError, err.Error())
		return
	}
	divisions, err := service.GetDivisions()
	if err != nil {
		c.String(http.StatusInternalServerError, err.Error())
		return
	}
	activeCount, rootCount, totalItems := 0, 0, 0
	for _, item := range items {
		if item.IsActive {
			activeCount++
		}
		if item.ParentID == 0 {
			rootCount++
		}
		totalItems += item.ItemCount
	}
	pageItems, pagination := paginateAssetSlice(c, items)
	Render(c, "catalog_category.html", gin.H{
		"Title": "Kategori Item", "Page": "catalog_category", "Items": pageItems,
		"AllCategories": items, "Divisions": divisions, "Pagination": pagination,
		"Error": message, "TotalCategories": len(items), "ActiveCount": activeCount,
		"RootCount": rootCount, "TotalItems": totalItems,
	})
}

func renderCatalogItemPage(c *gin.Context, service *services.CatalogService, message string) {
	items, err := service.GetItems()
	if err != nil {
		c.String(http.StatusInternalServerError, err.Error())
		return
	}
	categories, err := service.GetCategories()
	if err != nil {
		c.String(http.StatusInternalServerError, err.Error())
		return
	}
	itemTypes, err := service.GetItemTypes()
	if err != nil {
		c.String(http.StatusInternalServerError, err.Error())
		return
	}
	assetTypes, err := service.GetAssetTypes()
	if err != nil {
		c.String(http.StatusInternalServerError, err.Error())
		return
	}
	activeCount, goodsCount, serviceCount, totalDetails := 0, 0, 0, 0
	for _, item := range items {
		if item.IsActive {
			activeCount++
		}
		if item.ItemTypeCode == "GOODS" {
			goodsCount++
		}
		if item.ItemTypeCode == "SERVICE" {
			serviceCount++
		}
		totalDetails += item.DetailCount
	}
	pageItems, pagination := paginateAssetSlice(c, items)
	Render(c, "catalog_item.html", gin.H{
		"Title": "Master Item", "Page": "catalog_item", "Items": pageItems,
		"Categories": categories, "ItemTypes": itemTypes, "AssetTypes": assetTypes,
		"Pagination": pagination, "Error": message, "TotalItems": len(items),
		"ActiveCount": activeCount, "GoodsCount": goodsCount,
		"ServiceCount": serviceCount, "TotalDetails": totalDetails,
	})
}

func renderCatalogItemDetailPage(c *gin.Context, service *services.CatalogService, message string) {
	details, err := service.GetItemDetails()
	if err != nil {
		c.String(http.StatusInternalServerError, err.Error())
		return
	}
	items, err := service.GetItems()
	if err != nil {
		c.String(http.StatusInternalServerError, err.Error())
		return
	}

	selectedItemID, _ := strconv.ParseInt(c.Query("item_id"), 10, 64)
	if selectedItemID <= 0 {
		selectedItemID = parseInt64Form(c, "return_item_id")
	}
	selectedItemName := ""
	if selectedItemID > 0 {
		filtered := make([]models.CatalogItemDetail, 0)
		for _, detail := range details {
			if detail.ItemID == selectedItemID {
				filtered = append(filtered, detail)
			}
		}
		details = filtered
		for _, item := range items {
			if item.ID == selectedItemID {
				selectedItemName = item.ItemCode + " - " + item.Name
				break
			}
		}
	}

	distinctItems := map[int64]bool{}
	withUnit := 0
	for _, detail := range details {
		distinctItems[detail.ItemID] = true
		if detail.Unit != "" {
			withUnit++
		}
	}
	pageItems, pagination := paginateAssetSlice(c, details)
	paginationQuery := ""
	if selectedItemID > 0 {
		paginationQuery = fmt.Sprintf("&item_id=%d", selectedItemID)
	}
	Render(c, "catalog_item_detail.html", gin.H{
		"Title": "Detail Item", "Page": "catalog_item_detail", "Items": pageItems,
		"CatalogItems": items, "Pagination": pagination, "Error": message,
		"TotalDetails": len(details), "ConfiguredItems": len(distinctItems),
		"WithUnit": withUnit, "AvailableItems": len(items),
		"SelectedItemID": selectedItemID, "SelectedItemName": selectedItemName,
		"PaginationQuery": paginationQuery,
	})
}

func renderCatalogPackagePage(c *gin.Context, service *services.CatalogService, message string) {
	items, err := service.GetPackages()
	if err != nil {
		c.String(http.StatusInternalServerError, err.Error())
		return
	}
	categories, err := service.GetCategories()
	if err != nil {
		c.String(http.StatusInternalServerError, err.Error())
		return
	}
	divisions, err := service.GetDivisions()
	if err != nil {
		c.String(http.StatusInternalServerError, err.Error())
		return
	}
	activeCount, totalLines, optionalLines := 0, 0, 0
	for _, item := range items {
		if item.IsActive {
			activeCount++
		}
		totalLines += item.ItemCount
		optionalLines += item.OptionalCount
	}
	pageItems, pagination := paginateAssetSlice(c, items)
	Render(c, "catalog_package.html", gin.H{
		"Title": "Paket Item", "Page": "catalog_package", "Items": pageItems,
		"Categories": categories, "Divisions": divisions, "Pagination": pagination,
		"Error": message, "TotalPackages": len(items), "ActiveCount": activeCount,
		"TotalLines": totalLines, "OptionalLines": optionalLines,
	})
}

func renderCatalogPackageItemPage(c *gin.Context, service *services.CatalogService, message string) {
	items, err := service.GetPackageItems()
	if err != nil {
		c.String(http.StatusInternalServerError, err.Error())
		return
	}
	packages, err := service.GetPackages()
	if err != nil {
		c.String(http.StatusInternalServerError, err.Error())
		return
	}
	catalogItems, err := service.GetItems()
	if err != nil {
		c.String(http.StatusInternalServerError, err.Error())
		return
	}
	configuredPackages := map[int64]bool{}
	requiredCount, optionalCount := 0, 0
	for _, item := range items {
		configuredPackages[item.PackageID] = true
		if item.IsOptional {
			optionalCount++
		} else {
			requiredCount++
		}
	}
	pageItems, pagination := paginateAssetSlice(c, items)
	Render(c, "catalog_package_item.html", gin.H{
		"Title": "Isi Paket", "Page": "catalog_package_item", "Items": pageItems,
		"Packages": packages, "CatalogItems": catalogItems, "Pagination": pagination,
		"Error": message, "TotalLines": len(items), "ConfiguredPackages": len(configuredPackages),
		"RequiredCount": requiredCount, "OptionalCount": optionalCount,
	})
}

func renderVendorItemPricePage(c *gin.Context, service *services.CatalogService, message string) {
	prices, err := service.GetVendorItemPrices()
	if err != nil {
		c.String(http.StatusInternalServerError, err.Error())
		return
	}
	vendors, err := service.GetVendors()
	if err != nil {
		c.String(http.StatusInternalServerError, err.Error())
		return
	}
	items, err := service.GetItems()
	if err != nil {
		c.String(http.StatusInternalServerError, err.Error())
		return
	}

	activeCount, preferredCount := 0, 0
	distinctVendors := map[int64]bool{}
	for _, price := range prices {
		distinctVendors[price.VendorID] = true
		if price.IsCurrentlyValid {
			activeCount++
		}
		if price.IsPreferred {
			preferredCount++
		}
	}
	pageItems, pagination := paginateAssetSlice(c, prices)
	Render(c, "vendor_item_price.html", gin.H{
		"Title": "Harga Vendor", "Page": "vendor_item_price", "Items": pageItems,
		"Vendors": vendors, "CatalogItems": items, "Pagination": pagination,
		"Error": message, "TotalPrices": len(prices), "ActiveCount": activeCount,
		"PreferredCount": preferredCount, "VendorCount": len(distinctVendors),
		"Today": time.Now().Format("2006-01-02"),
	})
}

func bindCatalogItemTypeInput(c *gin.Context) models.CatalogItemTypeInput {
	return models.CatalogItemTypeInput{
		Code: c.PostForm("code"), Name: c.PostForm("name"),
		Description: c.PostForm("description"), IsActive: c.PostForm("is_active") != "0",
	}
}

func bindCatalogCategoryInput(c *gin.Context) models.CatalogItemCategoryInput {
	return models.CatalogItemCategoryInput{
		Code: c.PostForm("code"), Name: c.PostForm("name"),
		ParentID:        parseInt64Form(c, "parent_id"),
		OwnerDivisionID: parseIntForm(c, "owner_division_id"),
		Description:     c.PostForm("description"), IsActive: c.PostForm("is_active") != "0",
	}
}

func bindCatalogItemInput(c *gin.Context) models.CatalogItemInput {
	return models.CatalogItemInput{
		ItemCode: c.PostForm("item_code"), CategoryID: parseInt64Form(c, "category_id"),
		ItemTypeID: parseInt64Form(c, "item_type_id"), AssetTypeID: parseInt64Form(c, "asset_type_id"),
		Name: c.PostForm("name"), UOM: c.PostForm("uom"), Description: c.PostForm("description"),
		IsAssetCandidate: c.PostForm("is_asset_candidate") == "1",
		IsActive:         c.PostForm("is_active") != "0",
	}
}

func bindCatalogItemDetailInput(c *gin.Context) models.CatalogItemDetailInput {
	return models.CatalogItemDetailInput{
		ItemID: parseInt64Form(c, "item_id"), DetailName: c.PostForm("detail_name"),
		DetailValue: c.PostForm("detail_value"), Unit: c.PostForm("unit"),
		SortOrder: parseIntForm(c, "sort_order"),
	}
}

func bindCatalogPackageInput(c *gin.Context) models.CatalogPackageInput {
	return models.CatalogPackageInput{
		PackageCode: c.PostForm("package_code"), Name: c.PostForm("name"),
		CategoryID:      parseInt64Form(c, "category_id"),
		OwnerDivisionID: parseIntForm(c, "owner_division_id"),
		Description:     c.PostForm("description"), IsActive: c.PostForm("is_active") != "0",
	}
}

func redirectCatalogItemDetails(c *gin.Context) {
	itemID := parseInt64Form(c, "return_item_id")
	if itemID > 0 {
		c.Redirect(http.StatusSeeOther, fmt.Sprintf("/catalog/item-details?item_id=%d", itemID))
		return
	}
	c.Redirect(http.StatusSeeOther, "/catalog/item-details")
}

func bindCatalogPackageItemInput(c *gin.Context) models.CatalogPackageItemInput {
	return models.CatalogPackageItemInput{
		PackageID: parseInt64Form(c, "package_id"), ItemID: parseInt64Form(c, "item_id"),
		Qty: parseFloatForm(c, "qty"), IsOptional: c.PostForm("is_optional") == "1",
		SortOrder: parseIntForm(c, "sort_order"), Notes: c.PostForm("notes"),
	}
}

func bindVendorItemPriceInput(c *gin.Context) models.VendorItemPriceInput {
	return models.VendorItemPriceInput{
		VendorID: parseInt64Form(c, "vendor_id"), ItemID: parseInt64Form(c, "item_id"),
		UnitPrice: parseFloatForm(c, "unit_price"), CurrencyCode: c.PostForm("currency_code"),
		MinimumQty: parseFloatForm(c, "minimum_qty"), ValidFrom: c.PostForm("valid_from"),
		ValidUntil: c.PostForm("valid_until"), LeadTimeDays: parseIntForm(c, "lead_time_days"),
		QuotationReference: c.PostForm("quotation_reference"),
		IsPreferred:        c.PostForm("is_preferred") == "1",
		IsActive:           c.PostForm("is_active") != "0",
		Notes:              c.PostForm("notes"),
	}
}

func catalogService() *services.CatalogService {
	return &services.CatalogService{Repo: &repositories.CatalogRepository{DB: config.DB}}
}
