package controllers

import (
	"fmt"
	"gobase-app/config"
	"gobase-app/models"
	"gobase-app/repositories"
	"gobase-app/services"
	"net/http"
	"net/url"
	"strconv"
	"strings"

	"github.com/gin-contrib/sessions"
	"github.com/gin-gonic/gin"
)

func ApprovalTaskInboxIndex(c *gin.Context) {
	session := sessions.Default(c)
	userID := sessionUserID(session)
	page := parsePositiveInt(c.Query("page"), 1)
	filter := models.ApprovalTaskInboxFilter{
		UserID:         userID,
		Urgency:        strings.TrimSpace(c.Query("urgency")),
		SpendType:      strings.TrimSpace(c.Query("spend_type")),
		NeededDateSort: strings.TrimSpace(c.Query("needed_date_sort")),
		Page:           page,
		PerPage:        50,
	}

	service := buildApprovalTaskService()
	result, err := service.GetInbox(filter)
	if err != nil {
		c.String(http.StatusInternalServerError, err.Error())
		return
	}

	totalPages := 1
	if result.TotalRows > 0 {
		totalPages = (result.TotalRows + filter.PerPage - 1) / filter.PerPage
	}
	if page > totalPages {
		values := c.Request.URL.Query()
		values.Set("page", strconv.Itoa(totalPages))
		c.Redirect(http.StatusSeeOther, "/approval-tasks?"+values.Encode())
		return
	}

	Render(c, "approval_task_inbox.html", gin.H{
		"Title":             "Approval Inbox",
		"Page":              "approval_task_inbox",
		"Tasks":             result.Items,
		"PendingCount":      result.TotalRows,
		"QueueValueDisplay": formatQueueValue(result.QueueValue),
		"SLAStatusLabel":    "Optimal",
		"Filters": gin.H{
			"Urgency":        strings.ToUpper(strings.TrimSpace(filter.Urgency)),
			"SpendType":      strings.ToUpper(strings.TrimSpace(filter.SpendType)),
			"NeededDateSort": strings.ToLower(strings.TrimSpace(filter.NeededDateSort)),
		},
		"Pagination": buildPaginationView(c, page, filter.PerPage, result.TotalRows),
		"Success":    strings.TrimSpace(c.Query("success")),
		"Error":      strings.TrimSpace(c.Query("error")),
	})
}

func ApprovalTaskDetail(c *gin.Context) {
	taskID, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil || taskID <= 0 {
		c.Redirect(http.StatusSeeOther, "/approval-tasks?error="+url.QueryEscape("task approval tidak valid"))
		return
	}

	session := sessions.Default(c)
	userID := sessionUserID(session)

	service := buildApprovalTaskService()
	task, err := service.GetDetail(taskID, userID)
	if err != nil {
		c.Redirect(http.StatusSeeOther, "/approval-tasks?error="+url.QueryEscape("task approval tidak ditemukan atau belum siap ditindak"))
		return
	}

	data := gin.H{
		"Title":   "Approval Task Detail",
		"Page":    "approval_task_inbox",
		"Task":    task,
		"Success": strings.TrimSpace(c.Query("success")),
		"Error":   strings.TrimSpace(c.Query("error")),
	}

	if task.RefType == "PR" {
		prService := buildPurchaseRequestService()
		prDetail, err := prService.GetPurchaseRequestDetail(task.RefID, userID)
		if err != nil {
			c.Redirect(http.StatusSeeOther, "/approval-tasks?error="+url.QueryEscape(err.Error()))
			return
		}
		data["PR"] = prDetail
	}

	Render(c, "approval_task_detail.html", data)
}

func ApprovalTaskApprove(c *gin.Context) {
	handleApprovalTaskAction(c, "approve")
}

func ApprovalTaskReject(c *gin.Context) {
	handleApprovalTaskAction(c, "reject")
}

func handleApprovalTaskAction(c *gin.Context, action string) {
	taskID, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil || taskID <= 0 {
		c.Redirect(http.StatusSeeOther, "/approval-tasks?error=task approval tidak valid")
		return
	}
	redirectToDetail := strings.TrimSpace(c.PostForm("redirect_to")) == "detail"
	errorRedirect := "/approval-tasks"
	if redirectToDetail {
		errorRedirect = "/approval-tasks/" + strconv.FormatInt(taskID, 10)
	}

	session := sessions.Default(c)
	userID := sessionUserID(session)
	input := models.ApprovalActionInput{
		TaskID:      taskID,
		ActorUserID: userID,
		Action:      action,
		Comment:     strings.TrimSpace(c.PostForm("comment")),
		AuditContext: models.AuditContext{
			ActorUserID: userID,
			IPAddress:   c.ClientIP(),
			UserAgent:   c.Request.UserAgent(),
		},
	}

	service := buildApprovalTaskService()
	if action == "approve" {
		err = service.Approve(input)
	} else {
		err = service.Reject(input)
	}
	if err != nil {
		c.Redirect(http.StatusSeeOther, errorRedirect+"?error="+url.QueryEscape(err.Error()))
		return
	}

	success := "Task approval berhasil di-approve"
	if action == "reject" {
		success = "Task approval berhasil di-reject"
	}
	c.Redirect(http.StatusSeeOther, "/approval-tasks?success="+url.QueryEscape(success))
}

func buildApprovalTaskService() *services.ApprovalTaskService {
	return &services.ApprovalTaskService{
		Repo: &repositories.ApprovalTaskRepository{DB: config.DB},
	}
}

func formatQueueValue(value float64) string {
	if value >= 1000000000 {
		return fmt.Sprintf("Rp %.1f M", value/1000000000)
	}
	if value >= 1000000 {
		return fmt.Sprintf("Rp %.1f jt", value/1000000)
	}
	return fmt.Sprintf("Rp %.0f", value)
}

func parsePositiveInt(value string, fallback int) int {
	parsed, err := strconv.Atoi(strings.TrimSpace(value))
	if err != nil || parsed < 1 {
		return fallback
	}
	return parsed
}

func buildPaginationView(c *gin.Context, page, perPage, totalRows int) models.PaginationView {
	totalPages := 1
	if totalRows > 0 {
		totalPages = (totalRows + perPage - 1) / perPage
	}
	if page < 1 {
		page = 1
	}
	if page > totalPages {
		page = totalPages
	}

	from := 0
	to := 0
	if totalRows > 0 {
		from = (page-1)*perPage + 1
		to = page * perPage
		if to > totalRows {
			to = totalRows
		}
	}

	pages := make([]models.PaginationPage, 0, totalPages)
	for i := 1; i <= totalPages; i++ {
		pages = append(pages, models.PaginationPage{
			Page:   i,
			URL:    approvalTaskPageURL(c, i),
			Active: i == page,
		})
	}

	return models.PaginationView{
		Page:       page,
		PerPage:    perPage,
		TotalRows:  totalRows,
		TotalPages: totalPages,
		From:       from,
		To:         to,
		HasPrev:    page > 1,
		HasNext:    page < totalPages,
		PrevURL:    approvalTaskPageURL(c, page-1),
		NextURL:    approvalTaskPageURL(c, page+1),
		Pages:      pages,
	}
}

func approvalTaskPageURL(c *gin.Context, page int) string {
	if page < 1 {
		page = 1
	}
	values := c.Request.URL.Query()
	values.Set("page", strconv.Itoa(page))
	values.Del("success")
	values.Del("error")
	return "/approval-tasks?" + values.Encode()
}
