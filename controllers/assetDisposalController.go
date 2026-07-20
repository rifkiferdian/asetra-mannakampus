package controllers

import (
	"gobase-app/config"
	"gobase-app/models"
	"gobase-app/repositories"
	"gobase-app/services"
	"net/http"
	"net/url"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
)

func AssetDisposalTypeIndex(c *gin.Context) {
	items, err := assetDisposalService().GetDisposalTypes()
	if err != nil {
		c.String(http.StatusInternalServerError, err.Error())
		return
	}
	active, used := 0, 0
	for _, item := range items {
		if item.IsActive {
			active++
		}
		used += item.DisposalCount
	}
	Render(c, "asset_disposal_type.html", gin.H{
		"Title": "Jenis Disposal Aset", "Page": "asset_disposal_type",
		"Items": items, "Total": len(items), "Active": active, "Used": used,
		"Success": strings.TrimSpace(c.Query("success")), "Error": strings.TrimSpace(c.Query("error")),
	})
}

func AssetDisposalTypeStore(c *gin.Context) {
	input := models.AssetDisposalTypeInput{
		Code: c.PostForm("code"), Name: c.PostForm("name"), Description: c.PostForm("description"),
		IsActive: c.PostForm("is_active") != "0",
	}
	if err := assetDisposalService().SaveDisposalType(input); err != nil {
		redirectAssetDisposalType(c, "", err.Error())
		return
	}
	redirectAssetDisposalType(c, "Jenis disposal berhasil ditambahkan", "")
}

func AssetDisposalTypeUpdate(c *gin.Context) {
	input := models.AssetDisposalTypeInput{
		ID: parseInt64Form(c, "id"), Code: c.PostForm("code"), Name: c.PostForm("name"),
		Description: c.PostForm("description"), IsActive: c.PostForm("is_active") != "0",
	}
	if err := assetDisposalService().SaveDisposalType(input); err != nil {
		redirectAssetDisposalType(c, "", err.Error())
		return
	}
	redirectAssetDisposalType(c, "Jenis disposal berhasil diperbarui", "")
}

func AssetDisposalTypeDelete(c *gin.Context) {
	if err := assetDisposalService().DeleteDisposalType(parseInt64Param(c, "id")); err != nil {
		redirectAssetDisposalType(c, "", err.Error())
		return
	}
	redirectAssetDisposalType(c, "Jenis disposal berhasil dihapus", "")
}

func AssetDisposalIndex(c *gin.Context) {
	filter := assetDisposalFilter(c)
	service := assetDisposalService()
	result, err := service.GetDisposals(filter)
	if err != nil {
		c.String(http.StatusInternalServerError, err.Error())
		return
	}
	types, err := service.GetDisposalTypes()
	if err != nil {
		c.String(http.StatusInternalServerError, err.Error())
		return
	}
	assets, err := service.GetDisposalAssetOptions()
	if err != nil {
		c.String(http.StatusInternalServerError, err.Error())
		return
	}
	Render(c, "asset_disposal.html", gin.H{
		"Title": "Disposal Aset", "Page": "asset_disposal",
		"Items": result.Items, "Stats": result.Stats, "Types": types, "Assets": assets,
		"Filter": filter, "Pagination": assetDisposalPagination(filter, result),
		"Success": strings.TrimSpace(c.Query("success")), "Error": strings.TrimSpace(c.Query("error")),
	})
}

func AssetDisposalStore(c *gin.Context) {
	input := assetDisposalInput(c)
	if err := assetDisposalService().SaveDisposal(input); err != nil {
		redirectAssetDisposal(c, "", err.Error())
		return
	}
	redirectAssetDisposal(c, "Draft disposal berhasil dibuat", "")
}

func AssetDisposalUpdate(c *gin.Context) {
	input := assetDisposalInput(c)
	input.ID = parseInt64Form(c, "id")
	if err := assetDisposalService().SaveDisposal(input); err != nil {
		redirectAssetDisposal(c, "", err.Error())
		return
	}
	redirectAssetDisposal(c, "Draft disposal berhasil diperbarui", "")
}

func AssetDisposalPost(c *gin.Context) {
	if err := assetDisposalService().PostDisposal(parseInt64Param(c, "id"), depreciationAuditContext(c)); err != nil {
		redirectAssetDisposal(c, "", err.Error())
		return
	}
	redirectAssetDisposal(c, "Disposal berhasil diposting dan aset telah dihentikan", "")
}

func AssetDisposalCancel(c *gin.Context) {
	if err := assetDisposalService().CancelDisposal(
		parseInt64Param(c, "id"), c.PostForm("cancellation_reason"), depreciationAuditContext(c),
	); err != nil {
		redirectAssetDisposal(c, "", err.Error())
		return
	}
	redirectAssetDisposal(c, "Disposal berhasil dibatalkan", "")
}

func AssetDisposalReverse(c *gin.Context) {
	if err := assetDisposalService().ReverseDisposal(parseInt64Param(c, "id"), c.PostForm("reversal_reason"), depreciationAuditContext(c)); err != nil {
		redirectAssetDisposal(c, "", err.Error())
		return
	}
	redirectAssetDisposal(c, "Posting disposal berhasil direversal dan status aset dipulihkan", "")
}

func assetDisposalInput(c *gin.Context) models.AssetDisposalInput {
	return models.AssetDisposalInput{
		AssetID: parseInt64Form(c, "asset_id"), DisposalTypeID: parseInt64Form(c, "disposal_type_id"),
		DisposalDate: c.PostForm("disposal_date"), DisposalValue: parseFloatForm(c, "disposal_value"),
		BuyerName: c.PostForm("buyer_name"), DocumentReference: c.PostForm("document_reference"),
		Reason: c.PostForm("reason"), Notes: c.PostForm("notes"), AuditContext: depreciationAuditContext(c),
	}
}

func assetDisposalFilter(c *gin.Context) models.AssetDisposalFilter {
	page, _ := strconv.Atoi(c.Query("page"))
	if page < 1 {
		page = 1
	}
	status := strings.ToUpper(strings.TrimSpace(c.Query("status")))
	if status == "" {
		status = "ALL"
	}
	return models.AssetDisposalFilter{Status: status, Search: strings.TrimSpace(c.Query("search")), Page: page, PerPage: 50}
}

func assetDisposalPagination(filter models.AssetDisposalFilter, result models.AssetDisposalResult) assetPaginationMeta {
	page := filter.Page
	if page > result.TotalPages {
		page = result.TotalPages
	}
	start, end := 0, 0
	if result.TotalRows > 0 {
		start = (page-1)*filter.PerPage + 1
		end = start + len(result.Items) - 1
	}
	return assetPaginationMeta{
		CurrentPage: page, PrevPage: page - 1, NextPage: page + 1, TotalPages: result.TotalPages,
		PageSize: filter.PerPage, PageStart: start, PageEnd: end, TotalRows: result.TotalRows,
		HasPrev: page > 1, HasNext: page < result.TotalPages,
	}
}

func assetDisposalService() *services.AssetDisposalService {
	return &services.AssetDisposalService{Repo: &repositories.AssetDisposalRepository{DB: config.DB}}
}

func redirectAssetDisposalType(c *gin.Context, success, message string) {
	values := url.Values{}
	if success != "" {
		values.Set("success", success)
	}
	if message != "" {
		values.Set("error", message)
	}
	target := "/asset-disposal-types"
	if query := values.Encode(); query != "" {
		target += "?" + query
	}
	c.Redirect(http.StatusSeeOther, target)
}

func redirectAssetDisposal(c *gin.Context, success, message string) {
	values := url.Values{}
	if success != "" {
		values.Set("success", success)
	}
	if message != "" {
		values.Set("error", message)
	}
	target := "/asset-disposals"
	if query := values.Encode(); query != "" {
		target += "?" + query
	}
	c.Redirect(http.StatusSeeOther, target)
}
