package controllers

import (
	"gobase-app/services"
	"strconv"

	"github.com/gin-gonic/gin"
)

const assetListPageSize = 50

type assetPaginationMeta struct {
	CurrentPage int
	PrevPage    int
	NextPage    int
	TotalPages  int
	PageSize    int
	PageStart   int
	PageEnd     int
	TotalRows   int
	HasPrev     bool
	HasNext     bool
}

func renderAssetTypePage(c *gin.Context, service *services.AssetService, message string) {
	items, err := service.GetAssetTypes()
	if err != nil {
		renderAssetInternalError(c, err)
		return
	}
	totalAssets := 0
	activeTypes := 0
	for _, item := range items {
		totalAssets += item.AssetCount
		if item.IsActive {
			activeTypes++
		}
	}
	pageItems, pagination := paginateAssetSlice(c, items)
	Render(c, "asset_type.html", gin.H{
		"Title":           "Asset Types",
		"Page":            "asset_type",
		"Items":           pageItems,
		"Pagination":      pagination,
		"Error":           message,
		"TotalCategories": len(items),
		"ActiveTypes":     activeTypes,
		"InactiveTypes":   len(items) - activeTypes,
		"TotalAssets":     totalAssets,
	})
}

func renderAssetCategoryPage(c *gin.Context, service *services.AssetService, message string) {
	items, err := service.GetAssetCategories()
	if err != nil {
		renderAssetInternalError(c, err)
		return
	}
	types, err := service.GetAssetTypes()
	if err != nil {
		renderAssetInternalError(c, err)
		return
	}
	activeCategories := 0
	totalAssets := 0
	for _, item := range items {
		totalAssets += item.AssetCount
		if item.IsActive {
			activeCategories++
		}
	}
	pageItems, pagination := paginateAssetSlice(c, items)
	Render(c, "asset_category.html", gin.H{
		"Title": "Asset Categories", "Page": "asset_category",
		"Items": pageItems, "Types": types, "Pagination": pagination, "Error": message,
		"TotalCategories": len(items), "ActiveCategories": activeCategories,
		"InactiveCategories": len(items) - activeCategories, "TotalAssets": totalAssets,
	})
}

func renderComponentTypePage(c *gin.Context, service *services.AssetService, message string) {
	items, err := service.GetComponentTypes()
	if err != nil {
		renderAssetInternalError(c, err)
		return
	}
	activeTypes := 0
	totalComponents := 0
	for _, item := range items {
		if item.IsActive {
			activeTypes++
		}
		totalComponents += item.ComponentCount
	}
	Render(c, "component_type.html", gin.H{
		"Title":           "Component Types",
		"Page":            "component_type",
		"Items":           items,
		"Error":           message,
		"TotalTypes":      len(items),
		"ActiveTypes":     activeTypes,
		"InactiveTypes":   len(items) - activeTypes,
		"TotalComponents": totalComponents,
	})
}

func renderAssetLocationPage(c *gin.Context, service *services.AssetService, message string) {
	items, err := service.GetLocations()
	if err != nil {
		renderAssetInternalError(c, err)
		return
	}
	stores, err := service.GetStoreOptions(currentAccessScope(c))
	if err != nil {
		renderAssetInternalError(c, err)
		return
	}
	typeCounts := map[string]int{}
	activeCount := 0
	for _, item := range items {
		typeCounts[item.LocationType]++
		if item.IsActive {
			activeCount++
		}
	}
	pageItems, pagination := paginateAssetSlice(c, items)
	Render(c, "asset_location.html", gin.H{
		"Title":          "Asset Locations",
		"Page":           "asset_location",
		"Items":          pageItems,
		"AllLocations":   items,
		"Pagination":     pagination,
		"Stores":         stores,
		"Error":          message,
		"TotalLocations": len(items),
		"ActiveCount":    activeCount,
		"InactiveCount":  len(items) - activeCount,
		"WarehouseCount": typeCounts["WAREHOUSE"],
		"RoomCount":      typeCounts["ROOM"],
	})
}

func renderAssetPage(c *gin.Context, service *services.AssetService, message string) {
	page, _ := strconv.Atoi(c.Query("page"))
	if page < 1 {
		page = 1
	}
	items, totalAssets, statusCounts, err := service.GetAssetsPage(page, assetListPageSize, currentAccessScope(c))
	if err != nil {
		renderAssetInternalError(c, err)
		return
	}
	types, err := service.GetAssetTypes()
	if err != nil {
		renderAssetInternalError(c, err)
		return
	}
	categories, err := service.GetAssetCategories()
	if err != nil {
		renderAssetInternalError(c, err)
		return
	}
	locations, err := service.GetLocations()
	if err != nil {
		renderAssetInternalError(c, err)
		return
	}
	stores, err := service.GetStoreOptions(currentAccessScope(c))
	if err != nil {
		renderAssetInternalError(c, err)
		return
	}
	pagination := assetPaginationForTotal(page, assetListPageSize, totalAssets)
	Render(c, "asset.html", gin.H{
		"Title": "Assets", "Page": "asset", "Items": items, "Types": types,
		"Categories": categories, "Locations": locations, "Stores": stores, "Error": message, "Pagination": pagination,
		"TotalAssets":     totalAssets,
		"InUseAssets":     statusCounts["IN_USE"],
		"AvailableAssets": statusCounts["AVAILABLE"],
		"AttentionAssets": statusCounts["MAINTENANCE"] + statusCounts["BROKEN"],
	})
}

func assetPaginationForTotal(page, pageSize, totalRows int) assetPaginationMeta {
	totalPages := 1
	if totalRows > 0 {
		totalPages = (totalRows + pageSize - 1) / pageSize
	}
	if page > totalPages {
		page = totalPages
	}
	start := 0
	end := 0
	if totalRows > 0 {
		start = (page-1)*pageSize + 1
		end = page * pageSize
		if end > totalRows {
			end = totalRows
		}
	}
	return assetPaginationMeta{CurrentPage: page, PrevPage: page - 1, NextPage: page + 1, TotalPages: totalPages, PageSize: pageSize, PageStart: start, PageEnd: end, TotalRows: totalRows, HasPrev: page > 1, HasNext: page < totalPages}
}

func renderAssetComponentPage(c *gin.Context, service *services.AssetService, message string) {
	items, err := service.GetComponents(currentAccessScope(c))
	if err != nil {
		renderAssetInternalError(c, err)
		return
	}
	types, err := service.GetComponentTypes()
	if err != nil {
		renderAssetInternalError(c, err)
		return
	}
	locations, err := service.GetLocations()
	if err != nil {
		renderAssetInternalError(c, err)
		return
	}
	assets, err := service.GetAssetOptions(currentAccessScope(c))
	if err != nil {
		renderAssetInternalError(c, err)
		return
	}
	statusCounts := map[string]int{}
	for _, item := range items {
		statusCounts[item.Status]++
	}
	Render(c, "asset_component.html", gin.H{
		"Title": "Asset Components", "Page": "asset_component", "Items": items,
		"Types": types, "Locations": locations, "Assets": assets, "Error": message,
		"TotalComponents":     len(items),
		"InstalledComponents": statusCounts["INSTALLED"],
		"StorageComponents":   statusCounts["IN_STORAGE"],
		"AttentionComponents": statusCounts["MAINTENANCE"] + statusCounts["BROKEN"],
	})
}

func renderAssetMovementPage(c *gin.Context, service *services.AssetService, message string) {
	items, err := service.GetAssetMovements(currentAccessScope(c))
	if err != nil {
		renderAssetInternalError(c, err)
		return
	}
	assets, err := service.GetAssetOptions(currentAccessScope(c))
	if err != nil {
		renderAssetInternalError(c, err)
		return
	}
	locations, err := service.GetLocations()
	if err != nil {
		renderAssetInternalError(c, err)
		return
	}
	stores, err := service.GetStoreOptions(currentAccessScope(c))
	if err != nil {
		renderAssetInternalError(c, err)
		return
	}
	users, err := service.GetUserOptions()
	if err != nil {
		renderAssetInternalError(c, err)
		return
	}
	typeCounts := map[string]int{}
	for _, item := range items {
		typeCounts[item.MovementType]++
	}
	pageItems, pagination := paginateAssetSlice(c, items)
	Render(c, "asset_movement.html", gin.H{
		"Title": "Asset Movements", "Page": "asset_movement", "Items": pageItems,
		"Assets": assets, "Locations": locations, "Stores": stores, "Users": users, "Error": message,
		"Pagination":         pagination,
		"TotalMovements":     len(items),
		"TransferMovements":  typeCounts["TRANSFER"],
		"AssignMovements":    typeCounts["ASSIGN"],
		"AttentionMovements": typeCounts["MAINTENANCE"] + typeCounts["BROKEN"] + typeCounts["DISPOSE"],
	})
}

func renderComponentMovementPage(c *gin.Context, service *services.AssetService, message string) {
	items, err := service.GetComponentMovements(currentAccessScope(c))
	if err != nil {
		renderAssetInternalError(c, err)
		return
	}
	components, err := service.GetComponentOptions(currentAccessScope(c))
	if err != nil {
		renderAssetInternalError(c, err)
		return
	}
	assets, err := service.GetAssetOptions(currentAccessScope(c))
	if err != nil {
		renderAssetInternalError(c, err)
		return
	}
	locations, err := service.GetLocations()
	if err != nil {
		renderAssetInternalError(c, err)
		return
	}
	typeCounts := map[string]int{}
	for _, item := range items {
		typeCounts[item.MovementType]++
	}
	pageItems, pagination := paginateAssetSlice(c, items)
	Render(c, "asset_component_movement.html", gin.H{
		"Title": "Component Movements", "Page": "asset_component_movement", "Items": pageItems,
		"Components": components, "Assets": assets, "Locations": locations, "Error": message,
		"Pagination":         pagination,
		"TotalMovements":     len(items),
		"InstallMovements":   typeCounts["INSTALL"],
		"StorageMovements":   typeCounts["UNINSTALL"] + typeCounts["RETURN_TO_STORAGE"] + typeCounts["TRANSFER"],
		"AttentionMovements": typeCounts["MAINTENANCE"] + typeCounts["BROKEN"] + typeCounts["DISPOSE"],
	})
}
