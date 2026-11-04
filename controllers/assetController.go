package controllers

import (
	"gobase-app/models"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
)

func AssetTypeIndex(c *gin.Context) {
	renderAssetTypePage(c, assetService(), "")
}

func AssetTypeStore(c *gin.Context) {
	input := models.AssetTypeInput{
		Code:        c.PostForm("code"),
		AssetPrefix: c.PostForm("asset_prefix"),
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
		AssetPrefix: c.PostForm("asset_prefix"),
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
		renderAssetTypePage(c, assetService(), err.Error())
		return
	}
	c.Redirect(http.StatusSeeOther, "/asset-types")
}

func AssetCategoryIndex(c *gin.Context) {
	renderAssetCategoryPage(c, assetService(), "")
}

func AssetCategoryStore(c *gin.Context) {
	input := bindAssetCategoryInput(c)
	if err := assetService().SaveAssetCategory(input); err != nil {
		renderAssetCategoryPage(c, assetService(), err.Error())
		return
	}
	c.Redirect(http.StatusSeeOther, "/asset-categories")
}

func AssetCategoryUpdate(c *gin.Context) {
	input := bindAssetCategoryInput(c)
	input.ID = parseInt64Form(c, "id")
	if err := assetService().SaveAssetCategory(input); err != nil {
		renderAssetCategoryPage(c, assetService(), err.Error())
		return
	}
	c.Redirect(http.StatusSeeOther, "/asset-categories")
}

func AssetCategoryDelete(c *gin.Context) {
	id := parseInt64Param(c, "id")
	if err := assetService().DeleteAssetCategory(id); err != nil {
		renderAssetCategoryPage(c, assetService(), err.Error())
		return
	}
	c.Redirect(http.StatusSeeOther, "/asset-categories")
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
		renderComponentTypePage(c, assetService(), err.Error())
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
		renderAssetLocationPage(c, assetService(), err.Error())
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
	asset, err := service.GetAssetByID(id, currentAccessScope(c))
	if err != nil {
		c.String(http.StatusNotFound, err.Error())
		return
	}
	components, err := service.GetComponentsByAssetID(id)
	if err != nil {
		renderAssetInternalError(c, err)
		return
	}
	movements, err := service.GetAssetMovementsByAssetID(id, 5)
	if err != nil {
		renderAssetInternalError(c, err)
		return
	}
	depreciation, err := assetDepreciationService().GetAssetDepreciationDetail(id)
	if err != nil {
		renderAssetInternalError(c, err)
		return
	}
	depreciationPostings, err := assetDepreciationService().GetPostedDepreciationByAssetID(id, 12)
	if err != nil {
		renderAssetInternalError(c, err)
		return
	}
	disposal, err := assetDisposalService().GetPostedDisposalByAssetID(id)
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
	Render(c, "asset_detail.html", gin.H{
		"Title":                "Detail Aset",
		"Page":                 "asset",
		"Asset":                asset,
		"Components":           components,
		"Movements":            movements,
		"Depreciation":         depreciation,
		"DepreciationPostings": depreciationPostings,
		"Disposal":             disposal,
		"Types":                types,
		"Categories":           categories,
		"Locations":            locations,
		"Stores":               stores,
	})
}

func AssetStore(c *gin.Context) {
	input, err := bindAssetInput(c)
	if err != nil {
		renderAssetPage(c, assetService(), err.Error())
		return
	}
	input.AuditContext = depreciationAuditContext(c)
	if err := assetService().SaveAsset(input, currentAccessScope(c)); err != nil {
		renderAssetPage(c, assetService(), err.Error())
		return
	}
	c.Redirect(http.StatusSeeOther, "/asset-register")
}

func AssetUpdate(c *gin.Context) {
	input, err := bindAssetInput(c)
	if err != nil {
		renderAssetPage(c, assetService(), err.Error())
		return
	}
	input.ID = parseInt64Form(c, "id")
	input.AuditContext = depreciationAuditContext(c)
	if err := assetService().SaveAsset(input, currentAccessScope(c)); err != nil {
		renderAssetPage(c, assetService(), err.Error())
		return
	}
	if c.PostForm("redirect_to") == "detail" {
		c.Redirect(http.StatusSeeOther, "/asset-register/detail/"+strconv.FormatInt(input.ID, 10))
		return
	}
	c.Redirect(http.StatusSeeOther, "/asset-register")
}

func AssetDelete(c *gin.Context) {
	id := parseInt64Param(c, "id")
	if err := assetService().DeleteAsset(id, currentAccessScope(c)); err != nil {
		renderAssetPage(c, assetService(), err.Error())
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
	component, err := service.GetComponentByID(id, currentAccessScope(c))
	if err != nil {
		c.String(http.StatusNotFound, err.Error())
		return
	}
	movements, err := service.GetComponentMovementsByComponentID(id)
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
	input, err := bindComponentInput(c)
	if err != nil {
		renderAssetComponentPage(c, assetService(), err.Error())
		return
	}
	input.AuditContext = depreciationAuditContext(c)
	if err := assetService().SaveComponent(input, currentAccessScope(c)); err != nil {
		renderAssetComponentPage(c, assetService(), err.Error())
		return
	}
	c.Redirect(http.StatusSeeOther, "/asset-components")
}

func AssetComponentUpdate(c *gin.Context) {
	input, err := bindComponentInput(c)
	if err != nil {
		renderAssetComponentPage(c, assetService(), err.Error())
		return
	}
	input.ID = parseInt64Form(c, "id")
	input.AuditContext = depreciationAuditContext(c)
	if err := assetService().SaveComponent(input, currentAccessScope(c)); err != nil {
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
	if err := assetService().DeleteComponent(id, currentAccessScope(c)); err != nil {
		renderAssetComponentPage(c, assetService(), err.Error())
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
		AuditContext: depreciationAuditContext(c),
	}
	if err := assetService().CreateAssetMovement(input, currentAccessScope(c)); err != nil {
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
		AuditContext: depreciationAuditContext(c),
	}
	if err := assetService().CreateComponentMovement(input, currentAccessScope(c)); err != nil {
		renderComponentMovementPage(c, assetService(), err.Error())
		return
	}
	c.Redirect(http.StatusSeeOther, "/asset-component-movements")
}
