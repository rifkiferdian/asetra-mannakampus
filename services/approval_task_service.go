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

func (s *ApprovalTaskService) GetInbox(userID int) ([]models.ApprovalTaskInboxItem, error) {
	if userID <= 0 {
		return nil, errors.New("user login tidak valid")
	}
	return s.Repo.GetInboxByUser(userID)
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
