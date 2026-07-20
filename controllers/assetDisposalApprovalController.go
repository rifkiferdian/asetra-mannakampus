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

func AssetDisposalApprovalRuleIndex(c *gin.Context) {
	service := assetDisposalApprovalService()
	items, err := service.GetRules()
	if err != nil {
		c.String(http.StatusInternalServerError, err.Error())
		return
	}
	types, err := assetDisposalService().GetDisposalTypes()
	if err != nil {
		c.String(http.StatusInternalServerError, err.Error())
		return
	}
	assetTypes, err := assetService().GetAssetTypes()
	if err != nil {
		c.String(http.StatusInternalServerError, err.Error())
		return
	}
	roles, err := (&repositories.RoleRepository{DB: config.DB}).GetAll()
	if err != nil {
		c.String(http.StatusInternalServerError, err.Error())
		return
	}
	active, used := 0, 0
	for _, item := range items {
		if item.IsActive {
			active++
		}
		used += item.ApprovalCount
	}
	Render(c, "asset_disposal_approval_rule.html", gin.H{"Title": "Aturan Approval Disposal", "Page": "asset_disposal_approval_rule", "Items": items, "Types": types, "AssetTypes": assetTypes, "Roles": roles, "Total": len(items), "Active": active, "Used": used, "Success": strings.TrimSpace(c.Query("success")), "Error": strings.TrimSpace(c.Query("error"))})
}

func AssetDisposalApprovalRuleStore(c *gin.Context) {
	input, err := assetDisposalApprovalRuleInput(c)
	if err != nil {
		redirectDisposalApproval(c, "/asset-disposal-approval/rules", "", err.Error())
		return
	}
	if err := assetDisposalApprovalService().SaveRule(input); err != nil {
		redirectDisposalApproval(c, "/asset-disposal-approval/rules", "", err.Error())
		return
	}
	redirectDisposalApproval(c, "/asset-disposal-approval/rules", "Aturan approval berhasil disimpan", "")
}
func AssetDisposalApprovalRuleUpdate(c *gin.Context) {
	input, err := assetDisposalApprovalRuleInput(c)
	if err != nil {
		redirectDisposalApproval(c, "/asset-disposal-approval/rules", "", err.Error())
		return
	}
	if input.ID <= 0 {
		redirectDisposalApproval(c, "/asset-disposal-approval/rules", "", "Aturan approval tidak valid")
		return
	}
	if err := assetDisposalApprovalService().SaveRule(input); err != nil {
		redirectDisposalApproval(c, "/asset-disposal-approval/rules", "", err.Error())
		return
	}
	redirectDisposalApproval(c, "/asset-disposal-approval/rules", "Aturan approval berhasil diperbarui", "")
}
func AssetDisposalApprovalRuleDelete(c *gin.Context) {
	if err := assetDisposalApprovalService().DeleteRule(parseInt64Param(c, "id")); err != nil {
		redirectDisposalApproval(c, "/asset-disposal-approval/rules", "", err.Error())
		return
	}
	redirectDisposalApproval(c, "/asset-disposal-approval/rules", "Aturan approval berhasil dihapus", "")
}

func AssetDisposalApproverIndex(c *gin.Context) {
	items, err := assetDisposalApprovalService().GetApprovers()
	if err != nil {
		c.String(http.StatusInternalServerError, err.Error())
		return
	}
	roles, err := (&repositories.RoleRepository{DB: config.DB}).GetAll()
	if err != nil {
		c.String(http.StatusInternalServerError, err.Error())
		return
	}
	stores, err := (&repositories.StoreRepository{DB: config.DB}).GetAll()
	if err != nil {
		c.String(http.StatusInternalServerError, err.Error())
		return
	}
	users, err := assetService().GetUserOptions()
	if err != nil {
		c.String(http.StatusInternalServerError, err.Error())
		return
	}
	active, storeCount, hoCount := 0, 0, 0
	for _, item := range items {
		if item.IsActive {
			active++
		}
		if item.Scope == "STORE" {
			storeCount++
		} else {
			hoCount++
		}
	}
	Render(c, "asset_disposal_approver.html", gin.H{"Title": "Pemetaan Approver Disposal", "Page": "asset_disposal_approver", "Items": items, "Roles": roles, "Stores": stores, "Users": users, "Total": len(items), "Active": active, "StoreCount": storeCount, "HOCount": hoCount, "Success": strings.TrimSpace(c.Query("success")), "Error": strings.TrimSpace(c.Query("error"))})
}
func AssetDisposalApproverStore(c *gin.Context) {
	input := models.AssetDisposalApproverInput{ID: parseInt64Form(c, "id"), Scope: c.PostForm("scope"), StoreID: int(parseInt64Form(c, "store_id")), RoleID: parseInt64Form(c, "role_id"), UserID: int(parseInt64Form(c, "user_id")), IsActive: c.PostForm("is_active") != "0"}
	if err := assetDisposalApprovalService().SaveApprover(input); err != nil {
		redirectDisposalApproval(c, "/asset-disposal-approval/approvers", "", err.Error())
		return
	}
	redirectDisposalApproval(c, "/asset-disposal-approval/approvers", "Pemetaan approver berhasil disimpan", "")
}
func AssetDisposalApproverUpdate(c *gin.Context) {
	input := models.AssetDisposalApproverInput{ID: parseInt64Form(c, "id"), Scope: c.PostForm("scope"), StoreID: int(parseInt64Form(c, "store_id")), RoleID: parseInt64Form(c, "role_id"), UserID: int(parseInt64Form(c, "user_id")), IsActive: c.PostForm("is_active") != "0"}
	if input.ID <= 0 {
		redirectDisposalApproval(c, "/asset-disposal-approval/approvers", "", "Pemetaan approver tidak valid")
		return
	}
	if err := assetDisposalApprovalService().SaveApprover(input); err != nil {
		redirectDisposalApproval(c, "/asset-disposal-approval/approvers", "", err.Error())
		return
	}
	redirectDisposalApproval(c, "/asset-disposal-approval/approvers", "Pemetaan approver berhasil diperbarui", "")
}
func AssetDisposalApproverDelete(c *gin.Context) {
	if err := assetDisposalApprovalService().DeleteApprover(parseInt64Param(c, "id")); err != nil {
		redirectDisposalApproval(c, "/asset-disposal-approval/approvers", "", err.Error())
		return
	}
	redirectDisposalApproval(c, "/asset-disposal-approval/approvers", "Pemetaan approver berhasil dihapus", "")
}

func AssetDisposalApprovalInboxIndex(c *gin.Context) {
	filter := assetDisposalApprovalInboxFilter(c)
	result, err := assetDisposalApprovalService().GetInbox(currentSessionUserID(c), filter)
	if err != nil {
		c.String(http.StatusInternalServerError, err.Error())
		return
	}
	Render(c, "asset_disposal_approval_inbox.html", gin.H{"Title": "Approval Disposal", "Page": "asset_disposal_approval_inbox", "Items": result.Items, "Stats": result.Stats, "Filter": filter, "Pagination": assetDisposalApprovalPagination(filter, result), "Success": strings.TrimSpace(c.Query("success")), "Error": strings.TrimSpace(c.Query("error"))})
}
func AssetDisposalApprovalTaskApprove(c *gin.Context) {
	if err := assetDisposalApprovalService().ApproveTask(parseInt64Param(c, "id"), c.PostForm("comment"), depreciationAuditContext(c)); err != nil {
		redirectDisposalApproval(c, "/asset-disposal-approval/inbox", "", err.Error())
		return
	}
	redirectDisposalApproval(c, "/asset-disposal-approval/inbox", "Disposal berhasil disetujui", "")
}
func AssetDisposalApprovalTaskReject(c *gin.Context) {
	if err := assetDisposalApprovalService().RejectTask(parseInt64Param(c, "id"), c.PostForm("reason"), depreciationAuditContext(c)); err != nil {
		redirectDisposalApproval(c, "/asset-disposal-approval/inbox", "", err.Error())
		return
	}
	redirectDisposalApproval(c, "/asset-disposal-approval/inbox", "Disposal berhasil ditolak", "")
}

func AssetDisposalSubmit(c *gin.Context) {
	if err := assetDisposalApprovalService().Submit(parseInt64Param(c, "id"), depreciationAuditContext(c)); err != nil {
		redirectAssetDisposal(c, "", err.Error())
		return
	}
	redirectAssetDisposal(c, "Disposal berhasil diajukan untuk approval", "")
}

func assetDisposalApprovalRuleInput(c *gin.Context) (models.AssetDisposalApprovalRuleInput, error) {
	input := models.AssetDisposalApprovalRuleInput{ID: parseInt64Form(c, "id"), Name: c.PostForm("name"), DisposalTypeID: parseInt64Form(c, "disposal_type_id"), AssetTypeID: parseInt64Form(c, "asset_type_id"), MinBookValue: parseFloatForm(c, "min_book_value"), Priority: int(parseInt64Form(c, "priority")), IsActive: c.PostForm("is_active") != "0", EffectiveFrom: strings.TrimSpace(c.PostForm("effective_from")), EffectiveUntil: strings.TrimSpace(c.PostForm("effective_until"))}
	if raw := strings.TrimSpace(c.PostForm("max_book_value")); raw != "" {
		value, err := strconv.ParseFloat(raw, 64)
		if err != nil {
			return input, err
		}
		input.MaxBookValue = &value
	}
	orders, roles, scopes := c.PostFormArray("step_order[]"), c.PostFormArray("step_role_id[]"), c.PostFormArray("step_scope[]")
	parallel, required := c.PostFormArray("step_parallel[]"), c.PostFormArray("step_required[]")
	if len(orders) == 0 || len(orders) != len(roles) || len(orders) != len(scopes) || len(orders) != len(parallel) || len(orders) != len(required) {
		return input, &formInputError{"data tahap approval tidak lengkap"}
	}
	for i := range orders {
		order, _ := strconv.Atoi(orders[i])
		roleID, _ := strconv.ParseInt(roles[i], 10, 64)
		input.Steps = append(input.Steps, models.AssetDisposalApprovalRuleStep{StepOrder: order, RoleID: roleID, Scope: scopes[i], IsParallel: parallel[i] == "1", IsRequired: required[i] != "0"})
	}
	return input, nil
}

type formInputError struct{ message string }

func (e *formInputError) Error() string { return e.message }
func assetDisposalApprovalInboxFilter(c *gin.Context) models.AssetDisposalApprovalInboxFilter {
	page, _ := strconv.Atoi(c.Query("page"))
	if page < 1 {
		page = 1
	}
	status := strings.ToUpper(strings.TrimSpace(c.Query("status")))
	if status == "" {
		status = "ALL"
	}
	return models.AssetDisposalApprovalInboxFilter{Status: status, Search: strings.TrimSpace(c.Query("search")), Page: page, PerPage: 50}
}
func assetDisposalApprovalPagination(filter models.AssetDisposalApprovalInboxFilter, result models.AssetDisposalApprovalInboxResult) assetPaginationMeta {
	page := filter.Page
	if page > result.TotalPages {
		page = result.TotalPages
	}
	start, end := 0, 0
	if result.TotalRows > 0 {
		start = (page-1)*filter.PerPage + 1
		end = start + len(result.Items) - 1
	}
	return assetPaginationMeta{CurrentPage: page, PrevPage: page - 1, NextPage: page + 1, TotalPages: result.TotalPages, PageSize: filter.PerPage, PageStart: start, PageEnd: end, TotalRows: result.TotalRows, HasPrev: page > 1, HasNext: page < result.TotalPages}
}
func assetDisposalApprovalService() *services.AssetDisposalApprovalService {
	return &services.AssetDisposalApprovalService{Repo: &repositories.AssetDisposalApprovalRepository{DB: config.DB}}
}
func redirectDisposalApproval(c *gin.Context, path, success, message string) {
	values := url.Values{}
	if success != "" {
		values.Set("success", success)
	}
	if message != "" {
		values.Set("error", message)
	}
	if query := values.Encode(); query != "" {
		path += "?" + query
	}
	c.Redirect(http.StatusSeeOther, path)
}
