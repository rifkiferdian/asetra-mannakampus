package controllers

import (
	"fmt"
	"gobase-app/models"
	"gobase-app/services"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
)

func CatalogBrandIndex(c *gin.Context) {
	renderCatalogBrandPage(c, catalogService(), "")
}

func CatalogBrandStore(c *gin.Context) {
	if err := catalogService().SaveBrand(bindCatalogBrandInput(c)); err != nil {
		renderCatalogBrandPage(c, catalogService(), err.Error())
		return
	}
	c.Redirect(http.StatusSeeOther, "/catalog/brands")
}

func CatalogBrandUpdate(c *gin.Context) {
	input := bindCatalogBrandInput(c)
	input.ID = parseInt64Form(c, "id")
	if err := catalogService().SaveBrand(input); err != nil {
		renderCatalogBrandPage(c, catalogService(), err.Error())
		return
	}
	c.Redirect(http.StatusSeeOther, "/catalog/brands")
}

func CatalogBrandDelete(c *gin.Context) {
	if err := catalogService().DeleteBrand(parseInt64Param(c, "id")); err != nil {
		renderCatalogBrandPage(c, catalogService(), err.Error())
		return
	}
	c.Redirect(http.StatusSeeOther, "/catalog/brands")
}

func CatalogItemVariantIndex(c *gin.Context) {
	renderCatalogItemVariantPage(c, catalogService(), "")
}

func CatalogItemVariantStore(c *gin.Context) {
	if err := catalogService().SaveItemVariant(bindCatalogItemVariantInput(c)); err != nil {
		renderCatalogItemVariantPage(c, catalogService(), err.Error())
		return
	}
	redirectCatalogItemVariants(c)
}

func CatalogItemVariantUpdate(c *gin.Context) {
	input := bindCatalogItemVariantInput(c)
	input.ID = parseInt64Form(c, "id")
	if err := catalogService().SaveItemVariant(input); err != nil {
		renderCatalogItemVariantPage(c, catalogService(), err.Error())
		return
	}
	redirectCatalogItemVariants(c)
}

func CatalogItemVariantDelete(c *gin.Context) {
	if err := catalogService().DeleteItemVariant(parseInt64Param(c, "id")); err != nil {
		renderCatalogItemVariantPage(c, catalogService(), err.Error())
		return
	}
	redirectCatalogItemVariants(c)
}

func CatalogItemBOMIndex(c *gin.Context) {
	renderCatalogItemBOMPage(c, catalogService(), "")
}

func CatalogItemBOMStore(c *gin.Context) {
	if err := catalogService().SaveItemBOM(bindCatalogItemBOMInput(c)); err != nil {
		renderCatalogItemBOMPage(c, catalogService(), err.Error())
		return
	}
	c.Redirect(http.StatusSeeOther, "/catalog/item-boms")
}

func CatalogItemBOMUpdate(c *gin.Context) {
	input := bindCatalogItemBOMInput(c)
	input.ID = parseInt64Form(c, "id")
	if err := catalogService().SaveItemBOM(input); err != nil {
		renderCatalogItemBOMPage(c, catalogService(), err.Error())
		return
	}
	c.Redirect(http.StatusSeeOther, "/catalog/item-boms")
}

func CatalogItemBOMDelete(c *gin.Context) {
	if err := catalogService().DeleteItemBOM(parseInt64Param(c, "id")); err != nil {
		renderCatalogItemBOMPage(c, catalogService(), err.Error())
		return
	}
	c.Redirect(http.StatusSeeOther, "/catalog/item-boms")
}

func CatalogItemBOMItemIndex(c *gin.Context) {
	renderCatalogItemBOMItemPage(c, catalogService(), "")
}

func CatalogItemBOMItemStore(c *gin.Context) {
	if err := catalogService().SaveItemBOMItem(bindCatalogItemBOMItemInput(c)); err != nil {
		renderCatalogItemBOMItemPage(c, catalogService(), err.Error())
		return
	}
	redirectCatalogItemBOMItems(c)
}

func CatalogItemBOMItemUpdate(c *gin.Context) {
	input := bindCatalogItemBOMItemInput(c)
	input.ID = parseInt64Form(c, "id")
	if err := catalogService().SaveItemBOMItem(input); err != nil {
		renderCatalogItemBOMItemPage(c, catalogService(), err.Error())
		return
	}
	redirectCatalogItemBOMItems(c)
}

func CatalogItemBOMItemDelete(c *gin.Context) {
	if err := catalogService().DeleteItemBOMItem(parseInt64Param(c, "id")); err != nil {
		renderCatalogItemBOMItemPage(c, catalogService(), err.Error())
		return
	}
	redirectCatalogItemBOMItems(c)
}

func renderCatalogBrandPage(c *gin.Context, service *services.CatalogService, message string) {
	items, err := service.GetBrands()
	if err != nil {
		c.String(http.StatusInternalServerError, err.Error())
		return
	}
	activeCount, usedCount, totalVariants := 0, 0, 0
	for _, item := range items {
		if item.IsActive {
			activeCount++
		}
		if item.VariantCount > 0 {
			usedCount++
		}
		totalVariants += item.VariantCount
	}
	pageItems, pagination := paginateAssetSlice(c, items)
	Render(c, "catalog_brand.html", gin.H{
		"Title": "Merek Item", "Page": "catalog_brand", "Items": pageItems,
		"Pagination": pagination, "Error": message, "TotalBrands": len(items),
		"ActiveCount": activeCount, "UsedCount": usedCount, "TotalVariants": totalVariants,
	})
}

func renderCatalogItemVariantPage(c *gin.Context, service *services.CatalogService, message string) {
	variants, err := service.GetItemVariants()
	if err != nil {
		c.String(http.StatusInternalServerError, err.Error())
		return
	}
	items, err := service.GetItems()
	if err != nil {
		c.String(http.StatusInternalServerError, err.Error())
		return
	}
	brands, err := service.GetBrands()
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
		filtered := make([]models.CatalogItemVariant, 0)
		for _, variant := range variants {
			if variant.ItemID == selectedItemID {
				filtered = append(filtered, variant)
			}
		}
		variants = filtered
		for _, item := range items {
			if item.ID == selectedItemID {
				selectedItemName = item.ItemCode + " - " + item.Name
				break
			}
		}
	}

	activeCount, withPriceCount, usedInBOMCount := 0, 0, 0
	for _, variant := range variants {
		if variant.IsActive {
			activeCount++
		}
		if variant.PriceCount > 0 {
			withPriceCount++
		}
		if variant.BOMUsageCount > 0 {
			usedInBOMCount++
		}
	}
	pageItems, pagination := paginateAssetSlice(c, variants)
	paginationQuery := ""
	if selectedItemID > 0 {
		paginationQuery = fmt.Sprintf("&item_id=%d", selectedItemID)
	}
	Render(c, "catalog_item_variant.html", gin.H{
		"Title": "Varian Item", "Page": "catalog_item_variant", "Items": pageItems,
		"CatalogItems": items, "Brands": brands, "Pagination": pagination,
		"PaginationQuery": paginationQuery, "Error": message,
		"SelectedItemID": selectedItemID, "SelectedItemName": selectedItemName,
		"TotalVariants": len(variants), "ActiveCount": activeCount,
		"WithPriceCount": withPriceCount, "UsedInBOMCount": usedInBOMCount,
	})
}

func renderCatalogItemBOMPage(c *gin.Context, service *services.CatalogService, message string) {
	items, err := service.GetItemBOMs()
	if err != nil {
		c.String(http.StatusInternalServerError, err.Error())
		return
	}
	catalogItems, err := service.GetItems()
	if err != nil {
		c.String(http.StatusInternalServerError, err.Error())
		return
	}
	activeCount, completeCount, totalComponents := 0, 0, 0
	for _, item := range items {
		if item.IsActive {
			activeCount++
		}
		if item.ComponentCount > 0 && item.MissingPriceCount == 0 {
			completeCount++
		}
		totalComponents += item.ComponentCount
	}
	pageItems, pagination := paginateAssetSlice(c, items)
	Render(c, "catalog_item_bom.html", gin.H{
		"Title": "Rakitan Item", "Page": "catalog_item_bom", "Items": pageItems,
		"CatalogItems": catalogItems, "Pagination": pagination, "Error": message,
		"TotalBOMs": len(items), "ActiveCount": activeCount,
		"CompleteCount": completeCount, "TotalComponents": totalComponents,
	})
}

func renderCatalogItemBOMItemPage(c *gin.Context, service *services.CatalogService, message string) {
	items, err := service.GetItemBOMItems()
	if err != nil {
		c.String(http.StatusInternalServerError, err.Error())
		return
	}
	boms, err := service.GetItemBOMs()
	if err != nil {
		c.String(http.StatusInternalServerError, err.Error())
		return
	}
	catalogItems, err := service.GetItems()
	if err != nil {
		c.String(http.StatusInternalServerError, err.Error())
		return
	}
	variants, err := service.GetItemVariants()
	if err != nil {
		c.String(http.StatusInternalServerError, err.Error())
		return
	}
	vendors, err := service.GetVendors()
	if err != nil {
		c.String(http.StatusInternalServerError, err.Error())
		return
	}

	selectedBOMID, _ := strconv.ParseInt(c.Query("bom_id"), 10, 64)
	if selectedBOMID <= 0 {
		selectedBOMID = parseInt64Form(c, "return_bom_id")
	}
	selectedBOMName := ""
	if selectedBOMID > 0 {
		filtered := make([]models.CatalogItemBOMItem, 0)
		for _, item := range items {
			if item.BOMID == selectedBOMID {
				filtered = append(filtered, item)
			}
		}
		items = filtered
		for _, bom := range boms {
			if bom.ID == selectedBOMID {
				selectedBOMName = bom.BOMCode + " - " + bom.Name
				break
			}
		}
	}

	requiredCount, missingPriceCount := 0, 0
	total := 0.0
	for _, item := range items {
		if item.IsRequired {
			requiredCount++
		}
		if !item.HasCurrentPrice {
			missingPriceCount++
		}
		total += item.LineTotal
	}
	pageItems, pagination := paginateAssetSlice(c, items)
	paginationQuery := ""
	if selectedBOMID > 0 {
		paginationQuery = fmt.Sprintf("&bom_id=%d", selectedBOMID)
	}
	Render(c, "catalog_item_bom_item.html", gin.H{
		"Title": "Komponen Rakitan", "Page": "catalog_item_bom_item", "Items": pageItems,
		"BOMs": boms, "CatalogItems": catalogItems, "Variants": variants, "Vendors": vendors,
		"Pagination": pagination, "PaginationQuery": paginationQuery, "Error": message,
		"SelectedBOMID": selectedBOMID, "SelectedBOMName": selectedBOMName,
		"TotalComponents": len(items), "RequiredCount": requiredCount,
		"MissingPriceCount": missingPriceCount, "EstimatedTotalDisplay": formatCatalogAmount(total),
	})
}

func bindCatalogBrandInput(c *gin.Context) models.CatalogBrandInput {
	return models.CatalogBrandInput{
		Code: c.PostForm("code"), Name: c.PostForm("name"),
		Description: c.PostForm("description"), IsActive: c.PostForm("is_active") != "0",
	}
}

func bindCatalogItemVariantInput(c *gin.Context) models.CatalogItemVariantInput {
	return models.CatalogItemVariantInput{
		ItemID: parseInt64Form(c, "item_id"), BrandID: parseInt64Form(c, "brand_id"),
		VariantCode: c.PostForm("variant_code"), ModelName: c.PostForm("model_name"),
		ManufacturerSKU: c.PostForm("manufacturer_sku"),
		Specification:   c.PostForm("specification"), IsActive: c.PostForm("is_active") != "0",
	}
}

func bindCatalogItemBOMInput(c *gin.Context) models.CatalogItemBOMInput {
	return models.CatalogItemBOMInput{
		ParentItemID: parseInt64Form(c, "parent_item_id"), BOMCode: c.PostForm("bom_code"),
		Name: c.PostForm("name"), VersionNo: parseIntForm(c, "version_no"),
		Description: c.PostForm("description"), IsActive: c.PostForm("is_active") != "0",
	}
}

func bindCatalogItemBOMItemInput(c *gin.Context) models.CatalogItemBOMItemInput {
	return models.CatalogItemBOMItemInput{
		BOMID: parseInt64Form(c, "bom_id"), ComponentItemID: parseInt64Form(c, "component_item_id"),
		VariantID:         parseInt64Form(c, "variant_id"),
		PreferredVendorID: parseInt64Form(c, "preferred_vendor_id"),
		Qty:               parseFloatForm(c, "qty"), IsRequired: c.PostForm("is_required") != "0",
		SortOrder: parseIntForm(c, "sort_order"), Notes: c.PostForm("notes"),
	}
}

func redirectCatalogItemVariants(c *gin.Context) {
	itemID := parseInt64Form(c, "return_item_id")
	if itemID > 0 {
		c.Redirect(http.StatusSeeOther, fmt.Sprintf("/catalog/item-variants?item_id=%d", itemID))
		return
	}
	c.Redirect(http.StatusSeeOther, "/catalog/item-variants")
}

func redirectCatalogItemBOMItems(c *gin.Context) {
	bomID := parseInt64Form(c, "return_bom_id")
	if bomID > 0 {
		c.Redirect(http.StatusSeeOther, fmt.Sprintf("/catalog/item-bom-items?bom_id=%d", bomID))
		return
	}
	c.Redirect(http.StatusSeeOther, "/catalog/item-bom-items")
}

func formatCatalogAmount(value float64) string {
	raw := strconv.FormatInt(int64(value+0.5), 10)
	result := ""
	for index, char := range raw {
		if index > 0 && (len(raw)-index)%3 == 0 {
			result += "."
		}
		result += string(char)
	}
	return result
}
