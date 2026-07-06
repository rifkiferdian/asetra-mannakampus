package controllers

import (
	"gobase-app/config"
	"gobase-app/models"
	"gobase-app/repositories"
	"gobase-app/services"
	"net/http"
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

func AssetTypeIndex(c *gin.Context) {
	renderAssetTypePage(c, assetService(), "")
}

func AssetTypeStore(c *gin.Context) {
	input := models.AssetTypeInput{
		Code:        c.PostForm("code"),
		Name:        c.PostForm("name"),
		Description: c.PostForm("description"),
		IsActive:    c.PostForm("is_active") != "0",
	}
	if err := assetService().SaveAssetType(input); err != nil {
		renderAssetTypePage(c, assetService(), err.Error())
		return
	}
	c.Redirect(http.StatusSeeOther, "/asset-types")
}

func AssetTypeUpdate(c *gin.Context) {
	input := models.AssetTypeInput{
		ID:          parseInt64Form(c, "id"),
		Code:        c.PostForm("code"),
		Name:        c.PostForm("name"),
		Description: c.PostForm("description"),
		IsActive:    c.PostForm("is_active") != "0",
	}
	if err := assetService().SaveAssetType(input); err != nil {
		renderAssetTypePage(c, assetService(), err.Error())
		return
	}
	c.Redirect(http.StatusSeeOther, "/asset-types")
}

func AssetTypeDelete(c *gin.Context) {
	id := parseInt64Param(c, "id")
	if err := assetService().DeleteAssetType(id); err != nil {
		c.String(http.StatusInternalServerError, err.Error())
		return
	}
	c.Redirect(http.StatusSeeOther, "/asset-types")
}

func ComponentTypeIndex(c *gin.Context) {
	renderComponentTypePage(c, assetService(), "")
}

func ComponentTypeStore(c *gin.Context) {
	input := models.ComponentTypeInput{
		Code:        c.PostForm("code"),
		Name:        c.PostForm("name"),
		Description: c.PostForm("description"),
		IsActive:    c.PostForm("is_active") != "0",
	}
	if err := assetService().SaveComponentType(input); err != nil {
		renderComponentTypePage(c, assetService(), err.Error())
		return
	}
	c.Redirect(http.StatusSeeOther, "/component-types")
}

func ComponentTypeUpdate(c *gin.Context) {
	input := models.ComponentTypeInput{
		ID:          parseInt64Form(c, "id"),
		Code:        c.PostForm("code"),
		Name:        c.PostForm("name"),
		Description: c.PostForm("description"),
		IsActive:    c.PostForm("is_active") != "0",
	}
	if err := assetService().SaveComponentType(input); err != nil {
		renderComponentTypePage(c, assetService(), err.Error())
		return
	}
	c.Redirect(http.StatusSeeOther, "/component-types")
}

func ComponentTypeDelete(c *gin.Context) {
	id := parseInt64Param(c, "id")
	if err := assetService().DeleteComponentType(id); err != nil {
		c.String(http.StatusInternalServerError, err.Error())
		return
	}
	c.Redirect(http.StatusSeeOther, "/component-types")
}

func AssetLocationIndex(c *gin.Context) {
	renderAssetLocationPage(c, assetService(), "")
}

func AssetLocationStore(c *gin.Context) {
	input := models.AssetLocationInput{
		LocationCode: c.PostForm("location_code"),
		LocationName: c.PostForm("location_name"),
		StoreID:      parseIntForm(c, "store_id"),
		ParentID:     parseInt64Form(c, "parent_id"),
		LocationType: c.PostForm("location_type"),
		IsActive:     c.PostForm("is_active") != "0",
	}
	if err := assetService().SaveLocation(input); err != nil {
		renderAssetLocationPage(c, assetService(), err.Error())
		return
	}
	c.Redirect(http.StatusSeeOther, "/asset-locations")
}

func AssetLocationUpdate(c *gin.Context) {
	input := models.AssetLocationInput{
		ID:           parseInt64Form(c, "id"),
		LocationCode: c.PostForm("location_code"),
		LocationName: c.PostForm("location_name"),
		StoreID:      parseIntForm(c, "store_id"),
		ParentID:     parseInt64Form(c, "parent_id"),
		LocationType: c.PostForm("location_type"),
		IsActive:     c.PostForm("is_active") != "0",
	}
	if err := assetService().SaveLocation(input); err != nil {
		renderAssetLocationPage(c, assetService(), err.Error())
		return
	}
	c.Redirect(http.StatusSeeOther, "/asset-locations")
}

func AssetLocationDelete(c *gin.Context) {
	id := parseInt64Param(c, "id")
	if err := assetService().DeleteLocation(id); err != nil {
		c.String(http.StatusInternalServerError, err.Error())
		return
	}
	c.Redirect(http.StatusSeeOther, "/asset-locations")
}

func AssetIndex(c *gin.Context) {
	renderAssetPage(c, assetService(), "")
}

func AssetDetailIndex(c *gin.Context) {
	id := parseInt64Param(c, "id")
	service := assetService()
	asset, err := service.GetAssetByID(id)
	if err != nil {
		c.String(http.StatusNotFound, err.Error())
		return
	}
	components, err := service.GetComponentsByAssetID(id)
	if err != nil {
		c.String(http.StatusInternalServerError, err.Error())
		return
	}
	movements, err := service.GetAssetMovementsByAssetID(id, 5)
	if err != nil {
		c.String(http.StatusInternalServerError, err.Error())
		return
	}
	types, _ := service.GetAssetTypes()
	locations, _ := service.GetLocations()
	stores, _ := service.GetStoreOptions()
	Render(c, "asset_detail.html", gin.H{
		"Title":      "Asset Detail",
		"Page":       "asset",
		"Asset":      asset,
		"Components": components,
		"Movements":  movements,
		"Types":      types,
		"Locations":  locations,
		"Stores":     stores,
	})
}

func AssetStore(c *gin.Context) {
	input := bindAssetInput(c)
	if err := assetService().SaveAsset(input); err != nil {
		renderAssetPage(c, assetService(), err.Error())
		return
	}
	if c.PostForm("redirect_to") == "detail" {
		c.Redirect(http.StatusSeeOther, "/asset-register/detail/"+strconv.FormatInt(input.ID, 10))
		return
	}
	c.Redirect(http.StatusSeeOther, "/asset-register")
}

func AssetUpdate(c *gin.Context) {
	input := bindAssetInput(c)
	input.ID = parseInt64Form(c, "id")
	if err := assetService().SaveAsset(input); err != nil {
		renderAssetPage(c, assetService(), err.Error())
		return
	}
	c.Redirect(http.StatusSeeOther, "/asset-register")
}

func AssetDelete(c *gin.Context) {
	id := parseInt64Param(c, "id")
	if err := assetService().DeleteAsset(id); err != nil {
		c.String(http.StatusInternalServerError, err.Error())
		return
	}
	c.Redirect(http.StatusSeeOther, "/asset-register")
}

func AssetComponentIndex(c *gin.Context) {
	renderAssetComponentPage(c, assetService(), "")
}

func AssetComponentDetailIndex(c *gin.Context) {
	id := parseInt64Param(c, "id")
	service := assetService()
	component, err := service.GetComponentByID(id)
	if err != nil {
		c.String(http.StatusNotFound, err.Error())
		return
	}
	movements, err := service.GetComponentMovementsByComponentID(id)
	if err != nil {
		c.String(http.StatusInternalServerError, err.Error())
		return
	}
	types, _ := service.GetComponentTypes()
	locations, _ := service.GetLocations()
	assets, _ := service.GetAssetOptions()
	Render(c, "asset_component_detail.html", gin.H{
		"Title":     "Component Detail",
		"Page":      "asset_component",
		"Component": component,
		"Movements": movements,
		"Types":     types,
		"Locations": locations,
		"Assets":    assets,
	})
}

func AssetComponentStore(c *gin.Context) {
	input := bindComponentInput(c)
	if err := assetService().SaveComponent(input); err != nil {
		renderAssetComponentPage(c, assetService(), err.Error())
		return
	}
	c.Redirect(http.StatusSeeOther, "/asset-components")
}

func AssetComponentUpdate(c *gin.Context) {
	input := bindComponentInput(c)
	input.ID = parseInt64Form(c, "id")
	if err := assetService().SaveComponent(input); err != nil {
		renderAssetComponentPage(c, assetService(), err.Error())
		return
	}
	if c.PostForm("redirect_to") == "detail" {
		c.Redirect(http.StatusSeeOther, "/asset-components/detail/"+strconv.FormatInt(input.ID, 10))
		return
	}
	c.Redirect(http.StatusSeeOther, "/asset-components")
}

func AssetComponentDelete(c *gin.Context) {
	id := parseInt64Param(c, "id")
	if err := assetService().DeleteComponent(id); err != nil {
		c.String(http.StatusInternalServerError, err.Error())
		return
	}
	c.Redirect(http.StatusSeeOther, "/asset-components")
}

func AssetMovementIndex(c *gin.Context) {
	renderAssetMovementPage(c, assetService(), "")
}

func AssetMovementStore(c *gin.Context) {
	input := models.AssetMovementInput{
		AssetID:      parseInt64Form(c, "asset_id"),
		MovementType: c.PostForm("movement_type"),
		ToStoreID:    parseIntForm(c, "to_store_id"),
		ToLocationID: parseInt64Form(c, "to_location_id"),
		ToUserID:     parseIntForm(c, "to_user_id"),
		ActedBy:      currentSessionUserID(c),
		Notes:        c.PostForm("notes"),
	}
	if err := assetService().CreateAssetMovement(input); err != nil {
		renderAssetMovementPage(c, assetService(), err.Error())
		return
	}
	c.Redirect(http.StatusSeeOther, "/asset-movements")
}

func AssetComponentMovementIndex(c *gin.Context) {
	renderComponentMovementPage(c, assetService(), "")
}

func AssetComponentMovementStore(c *gin.Context) {
	input := models.AssetComponentMovementInput{
		ComponentID:  parseInt64Form(c, "component_id"),
		MovementType: c.PostForm("movement_type"),
		ToAssetID:    parseInt64Form(c, "to_asset_id"),
		ToLocationID: parseInt64Form(c, "to_location_id"),
		ActedBy:      currentSessionUserID(c),
		Notes:        c.PostForm("notes"),
	}
	if err := assetService().CreateComponentMovement(input); err != nil {
		renderComponentMovementPage(c, assetService(), err.Error())
		return
	}
	c.Redirect(http.StatusSeeOther, "/asset-component-movements")
}

func renderAssetTypePage(c *gin.Context, service *services.AssetService, message string) {
	items, err := service.GetAssetTypes()
	if err != nil {
		c.String(http.StatusInternalServerError, err.Error())
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

func renderComponentTypePage(c *gin.Context, service *services.AssetService, message string) {
	items, err := service.GetComponentTypes()
	if err != nil {
		c.String(http.StatusInternalServerError, err.Error())
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
		c.String(http.StatusInternalServerError, err.Error())
		return
	}
	stores, err := service.GetStoreOptions()
	if err != nil {
		c.String(http.StatusInternalServerError, err.Error())
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
	items, err := service.GetAssets()
	if err != nil {
		c.String(http.StatusInternalServerError, err.Error())
		return
	}
	types, _ := service.GetAssetTypes()
	locations, _ := service.GetLocations()
	stores, _ := service.GetStoreOptions()
	statusCounts := map[string]int{}
	for _, item := range items {
		statusCounts[item.Status]++
	}
	Render(c, "asset.html", gin.H{
		"Title": "Assets", "Page": "asset", "Items": items, "Types": types,
		"Locations": locations, "Stores": stores, "Error": message,
		"TotalAssets":     len(items),
		"InUseAssets":     statusCounts["IN_USE"],
		"AvailableAssets": statusCounts["AVAILABLE"],
		"AttentionAssets": statusCounts["MAINTENANCE"] + statusCounts["BROKEN"],
	})
}

func renderAssetComponentPage(c *gin.Context, service *services.AssetService, message string) {
	items, err := service.GetComponents()
	if err != nil {
		c.String(http.StatusInternalServerError, err.Error())
		return
	}
	types, _ := service.GetComponentTypes()
	locations, _ := service.GetLocations()
	assets, _ := service.GetAssetOptions()
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
	items, err := service.GetAssetMovements()
	if err != nil {
		c.String(http.StatusInternalServerError, err.Error())
		return
	}
	assets, _ := service.GetAssetOptions()
	locations, _ := service.GetLocations()
	stores, _ := service.GetStoreOptions()
	users, _ := service.GetUserOptions()
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
	items, err := service.GetComponentMovements()
	if err != nil {
		c.String(http.StatusInternalServerError, err.Error())
		return
	}
	components, _ := service.GetComponentOptions()
	assets, _ := service.GetAssetOptions()
	locations, _ := service.GetLocations()
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

func bindAssetInput(c *gin.Context) models.AssetInput {
	return models.AssetInput{
		AssetCode:                c.PostForm("asset_code"),
		AssetName:                c.PostForm("asset_name"),
		AssetTypeID:              parseInt64Form(c, "asset_type_id"),
		SerialNumber:             c.PostForm("serial_number"),
		StoreID:                  parseIntForm(c, "store_id"),
		LocationID:               parseInt64Form(c, "location_id"),
		AssignedPersonNIP:        c.PostForm("assigned_person_nip"),
		AssignedPersonName:       c.PostForm("assigned_person_name"),
		AssignedPersonDepartment: c.PostForm("assigned_person_department"),
		SourceGRItemID:           parseInt64Form(c, "source_gr_item_id"),
		AcquisitionDate:          c.PostForm("acquisition_date"),
		AcquisitionValue:         parseFloatForm(c, "acquisition_value"),
		Status:                   c.PostForm("status"),
		Notes:                    c.PostForm("notes"),
	}
}

func bindComponentInput(c *gin.Context) models.AssetComponentInput {
	return models.AssetComponentInput{
		ComponentCode:    c.PostForm("component_code"),
		ComponentName:    c.PostForm("component_name"),
		ComponentTypeID:  parseInt64Form(c, "component_type_id"),
		Brand:            c.PostForm("brand"),
		Model:            c.PostForm("model"),
		Specification:    c.PostForm("specification"),
		SerialNumber:     c.PostForm("serial_number"),
		ParentAssetID:    parseInt64Form(c, "parent_asset_id"),
		LocationID:       parseInt64Form(c, "location_id"),
		SourceGRItemID:   parseInt64Form(c, "source_gr_item_id"),
		AcquisitionDate:  c.PostForm("acquisition_date"),
		AcquisitionValue: parseFloatForm(c, "acquisition_value"),
		Status:           c.PostForm("status"),
		Notes:            c.PostForm("notes"),
	}
}

func assetService() *services.AssetService {
	return &services.AssetService{Repo: &repositories.AssetRepository{DB: config.DB}}
}

func parseInt64Param(c *gin.Context, name string) int64 {
	value, _ := strconv.ParseInt(c.Param(name), 10, 64)
	return value
}

func parseInt64Form(c *gin.Context, name string) int64 {
	value, _ := strconv.ParseInt(c.PostForm(name), 10, 64)
	return value
}

func parseIntForm(c *gin.Context, name string) int {
	value, _ := strconv.Atoi(c.PostForm(name))
	return value
}

func parseFloatForm(c *gin.Context, name string) float64 {
	value, _ := strconv.ParseFloat(c.PostForm(name), 64)
	return value
}

func paginateAssetSlice[T any](c *gin.Context, items []T) ([]T, assetPaginationMeta) {
	totalRows := len(items)
	totalPages := 1
	if totalRows > 0 {
		totalPages = (totalRows + assetListPageSize - 1) / assetListPageSize
	}

	page, _ := strconv.Atoi(c.Query("page"))
	if page < 1 {
		page = 1
	}
	if page > totalPages {
		page = totalPages
	}

	startIndex := (page - 1) * assetListPageSize
	if startIndex > totalRows {
		startIndex = totalRows
	}
	endIndex := startIndex + assetListPageSize
	if endIndex > totalRows {
		endIndex = totalRows
	}

	pageStart := 0
	pageEnd := 0
	if totalRows > 0 {
		pageStart = startIndex + 1
		pageEnd = endIndex
	}

	return items[startIndex:endIndex], assetPaginationMeta{
		CurrentPage: page,
		PrevPage:    page - 1,
		NextPage:    page + 1,
		TotalPages:  totalPages,
		PageSize:    assetListPageSize,
		PageStart:   pageStart,
		PageEnd:     pageEnd,
		TotalRows:   totalRows,
		HasPrev:     page > 1,
		HasNext:     page < totalPages,
	}
}
