package services

import (
	"database/sql"
	"errors"
	"fmt"
	"gobase-app/models"
	"gobase-app/repositories"
	"strconv"
	"strings"
)

type RoleService struct {
	Repo *repositories.RoleRepository
}

func (s *RoleService) GetRoles() ([]models.Role, error) {
	return s.Repo.GetAll()
}

// GetRoleDetail mengambil detail role beserta permission yang dimilikinya.
func (s *RoleService) GetRoleDetail(id int) (*models.RoleDetail, error) {
	if id <= 0 {
		return nil, errors.New("role id tidak valid")
	}

	role, err := s.Repo.GetByID(id)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, fmt.Errorf("role dengan id %d tidak ditemukan", id)
		}
		return nil, err
	}

	role.PermissionIDs = uniqueInt64(role.PermissionIDs)
	return role, nil
}

// CreateRole memvalidasi input lalu menyimpan role baru beserta permission yang dipilih.
func (s *RoleService) CreateRole(input models.RoleCreateInput) error {
	name := strings.TrimSpace(input.Name)
	guard := strings.TrimSpace(input.GuardName)
	if guard == "" {
		guard = "web"
	}

	if name == "" {
		return errors.New("nama role wajib diisi")
	}

	exists, err := s.Repo.ExistsByNameAndGuard(name, guard)
	if err != nil {
		return err
	}
	if exists {
		return fmt.Errorf("role '%s' sudah ada pada guard %s", name, guard)
	}

	permIDs := uniqueInt64(input.PermissionIDs)
	if len(permIDs) > 0 {
		found, err := s.Repo.FindExistingPermissionIDs(permIDs)
		if err != nil {
			return err
		}

		var missing []int64
		for _, id := range permIDs {
			if !found[id] {
				missing = append(missing, id)
			}
		}

		if len(missing) > 0 {
			return fmt.Errorf("permission tidak ditemukan: %s", formatInt64Slice(missing))
		}
	}

	_, err = s.Repo.CreateRoleWithPermissions(repositories.RoleCreateParams{
		Name:          name,
		GuardName:     guard,
		IsAdmin:       false,
		PermissionIDs: permIDs,
	})

	return err
}

// UpdateRole memvalidasi input lalu memperbarui role beserta permission yang dipilih.
func (s *RoleService) UpdateRole(input models.RoleUpdateInput) error {
	name := strings.TrimSpace(input.Name)
	guard := strings.TrimSpace(input.GuardName)
	if guard == "" {
		guard = "web"
	}

	if input.ID <= 0 {
		return errors.New("role tidak valid")
	}
	if name == "" {
		return errors.New("nama role wajib diisi")
	}

	role, err := s.Repo.GetByID(input.ID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return fmt.Errorf("role dengan id %d tidak ditemukan", input.ID)
		}
		return err
	}

	exists, err := s.Repo.ExistsByNameAndGuardExceptID(name, guard, input.ID)
	if err != nil {
		return err
	}
	if exists {
		return fmt.Errorf("role '%s' sudah ada pada guard %s", name, guard)
	}

	permIDs := uniqueInt64(input.PermissionIDs)
	if len(permIDs) > 0 {
		found, err := s.Repo.FindExistingPermissionIDs(permIDs)
		if err != nil {
			return err
		}

		var missing []int64
		for _, id := range permIDs {
			if !found[id] {
				missing = append(missing, id)
			}
		}

		if len(missing) > 0 {
			return fmt.Errorf("permission tidak ditemukan: %s", formatInt64Slice(missing))
		}
	}

	return s.Repo.UpdateRoleWithPermissions(repositories.RoleUpdateParams{
		ID:            input.ID,
		Name:          name,
		GuardName:     guard,
		IsAdmin:       role.IsAdmin,
		PermissionIDs: permIDs,
	})
}

// DeleteRole validates input and removes the role by ID.
func (s *RoleService) DeleteRole(id int) error {
	if id <= 0 {
		return errors.New("role id tidak valid")
	}
	return s.Repo.DeleteByID(id)
}

func uniqueInt64(values []int64) []int64 {
	seen := make(map[int64]bool)
	var result []int64
	for _, v := range values {
		if v <= 0 || seen[v] {
			continue
		}
		seen[v] = true
		result = append(result, v)
	}
	return result
}

func formatInt64Slice(values []int64) string {
	if len(values) == 0 {
		return ""
	}

	parts := make([]string, len(values))
	for i, v := range values {
		parts[i] = strconv.FormatInt(v, 10)
	}

	return strings.Join(parts, ", ")
}

