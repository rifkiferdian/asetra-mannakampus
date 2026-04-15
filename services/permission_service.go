package services

import (
	"gobase-app/models"
	"gobase-app/repositories"
)

type PermissionService struct {
	Repo *repositories.PermissionRepository
}

func (s *PermissionService) GetGroupedPermissions() ([]models.PermissionGroup, error) {
	return s.Repo.GetGrouped()
}

