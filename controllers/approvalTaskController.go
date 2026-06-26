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

	"github.com/gin-contrib/sessions"
	"github.com/gin-gonic/gin"
)

func ApprovalTaskInboxIndex(c *gin.Context) {
	session := sessions.Default(c)
	userID := sessionUserID(session)

	service := buildApprovalTaskService()
	items, err := service.GetInbox(userID)
	if err != nil {
		c.String(http.StatusInternalServerError, err.Error())
		return
	}

	Render(c, "approval_task_inbox.html", gin.H{
		"Title":   "Approval Inbox",
		"Page":    "approval_task_inbox",
		"Tasks":   items,
		"Success": strings.TrimSpace(c.Query("success")),
		"Error":   strings.TrimSpace(c.Query("error")),
	})
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
		c.Redirect(http.StatusSeeOther, "/approval-tasks?error="+url.QueryEscape(err.Error()))
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
