package controllers

import (
	"errors"
	"gobase-app/models"
	"gobase-app/services"
	"log"
	"net/http"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
)

func bindAssetInput(c *gin.Context) (models.AssetInput, error) {
	value, err := parseNonNegativeMoneyForm(c, "acquisition_value", true)
	if err != nil {
		return models.AssetInput{}, err
	}
	return models.AssetInput{
		AssetCode:                c.PostForm("asset_code"),
		AssetName:                c.PostForm("asset_name"),
		AssetTypeID:              parseInt64Form(c, "asset_type_id"),
		AssetCategoryID:          parseInt64Form(c, "asset_category_id"),
		SerialNumber:             c.PostForm("serial_number"),
		StoreID:                  parseIntForm(c, "store_id"),
		LocationID:               parseInt64Form(c, "location_id"),
		AssignedPersonNIP:        c.PostForm("assigned_person_nip"),
		AssignedPersonName:       c.PostForm("assigned_person_name"),
		AssignedPersonDepartment: c.PostForm("assigned_person_department"),
		SourceGRItemID:           parseInt64Form(c, "source_gr_item_id"),
		AcquisitionDate:          c.PostForm("acquisition_date"),
		AcquisitionValue:         value,
		Status:                   c.PostForm("status"),
		Notes:                    c.PostForm("notes"),
	}, nil
}

func bindAssetCategoryInput(c *gin.Context) models.AssetCategoryInput {
	return models.AssetCategoryInput{
		AssetTypeID:    parseInt64Form(c, "asset_type_id"),
		Code:           c.PostForm("code"),
		CategoryPrefix: c.PostForm("category_prefix"),
		Name:           c.PostForm("name"),
		Description:    c.PostForm("description"),
		IsActive:       c.PostForm("is_active") != "0",
	}
}

func bindComponentInput(c *gin.Context) (models.AssetComponentInput, error) {
	value, err := parseNonNegativeMoneyForm(c, "acquisition_value", false)
	if err != nil {
		return models.AssetComponentInput{}, err
	}
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
		AcquisitionValue: value,
		Status:           c.PostForm("status"),
		Notes:            c.PostForm("notes"),
	}, nil
}

func assetService() *services.AssetService {
	if configuredAssetService == nil {
		panic("asset service is not configured")
	}
	return configuredAssetService
}

var configuredAssetService *services.AssetService

func ConfigureAssetService(service *services.AssetService) {
	if service == nil {
		panic("asset service cannot be nil")
	}
	configuredAssetService = service
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

func parseNonNegativeMoneyForm(c *gin.Context, name string, required bool) (models.Money, error) {
	raw := strings.TrimSpace(c.PostForm(name))
	if raw == "" {
		if required {
			return 0, errors.New("nilai perolehan wajib diisi")
		}
		return 0, nil
	}
	value, err := models.ParseMoney(raw)
	if err != nil || value.IsNegative() {
		return 0, errors.New("nilai perolehan harus berupa angka nol atau lebih")
	}
	return value, nil
}

func parseNonNegativeDecimalForm(c *gin.Context, name string, required bool) (float64, error) {
	raw := strings.TrimSpace(c.PostForm(name))
	if raw == "" {
		if required {
			return 0, errors.New(name + " wajib diisi")
		}
		return 0, nil
	}
	value, err := strconv.ParseFloat(raw, 64)
	if err != nil || value < 0 {
		return 0, errors.New(name + " harus berupa angka nol atau lebih")
	}
	return value, nil
}

func renderAssetInternalError(c *gin.Context, err error) {
	log.Printf("asset module error: %v", err)
	c.HTML(http.StatusInternalServerError, "error.html", gin.H{
		"code_error": http.StatusInternalServerError,
		"error":      "Terjadi kesalahan saat memproses data aset",
	})
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
