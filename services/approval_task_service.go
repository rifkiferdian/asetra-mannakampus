package services

import (
	"errors"
	"gobase-app/models"
	"gobase-app/repositories"
	"strings"
)

type ApprovalTaskService struct {
	Repo *repositories.ApprovalTaskRepository
}

func (s *ApprovalTaskService) GetInbox(filter models.ApprovalTaskInboxFilter) (*models.ApprovalTaskInboxResult, error) {
	if filter.UserID <= 0 {
		return nil, errors.New("user login tidak valid")
	}
	filter.Urgency = strings.ToUpper(strings.TrimSpace(filter.Urgency))
	if filter.Urgency != "" && filter.Urgency != "NORMAL" && filter.Urgency != "URGENT" && filter.Urgency != "EMERGENCY" {
		filter.Urgency = ""
	}
	filter.SpendType = strings.ToUpper(strings.TrimSpace(filter.SpendType))
	if filter.SpendType != "" && filter.SpendType != "OPEX" && filter.SpendType != "CAPEX" {
		filter.SpendType = ""
	}
	filter.NeededDateSort = strings.ToLower(strings.TrimSpace(filter.NeededDateSort))
	if filter.NeededDateSort != "asc" && filter.NeededDateSort != "desc" {
		filter.NeededDateSort = ""
	}
	if filter.Page < 1 {
		filter.Page = 1
	}
	if filter.PerPage <= 0 {
		filter.PerPage = 50
	}
	if filter.PerPage > 50 {
		filter.PerPage = 50
	}
	return s.Repo.GetInboxByUser(filter)
}

func (s *ApprovalTaskService) GetDetail(taskID int64, userID int) (*models.ApprovalTaskDetail, error) {
	if taskID <= 0 {
		return nil, errors.New("task approval tidak valid")
	}
	if userID <= 0 {
		return nil, errors.New("user login tidak valid")
	}
	return s.Repo.GetDetailByID(taskID, userID)
}

func (s *ApprovalTaskService) Approve(input models.ApprovalActionInput) error {
	input.Comment = strings.TrimSpace(input.Comment)
	return s.Repo.Approve(input)
}

func (s *ApprovalTaskService) Reject(input models.ApprovalActionInput) error {
	input.Comment = strings.TrimSpace(input.Comment)
	if input.Comment == "" {
		return errors.New("catatan reject wajib diisi")
	}
	return s.Repo.Reject(input)
}
