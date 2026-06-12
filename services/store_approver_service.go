package services

import (
	"errors"
	"gobase-app/models"
	"gobase-app/repositories"
)

type StoreApproverService struct {
	Repo *repositories.StoreApproverRepository
}

func (s *StoreApproverService) GetStoreApprovers() ([]models.StoreApprover, error) {
	return s.Repo.GetAll()
}

func (s *StoreApproverService) CreateStoreApprover(input models.StoreApproverCreateInput) error {
	if input.StoreID <= 0 {
		return errors.New("store wajib dipilih")
	}
	if input.RoleID <= 0 {
		return errors.New("role approver wajib dipilih")
	}
	if input.UserID <= 0 {
		return errors.New("user approver wajib dipilih")
	}

	if err := s.validateReferences(input.StoreID, input.RoleID, input.UserID); err != nil {
		return err
	}

	exists, err := s.Repo.ExistsDuplicate(input.StoreID, input.RoleID, input.UserID, 0)
	if err != nil {
		return err
	}
	if exists {
		return errors.New("mapping store approver yang sama sudah ada")
	}

	return s.Repo.Create(input)
}

func (s *StoreApproverService) UpdateStoreApprover(input models.StoreApproverUpdateInput) error {
	if input.ID <= 0 {
		return errors.New("store approver tidak valid")
	}
	if input.StoreID <= 0 {
		return errors.New("store wajib dipilih")
	}
	if input.RoleID <= 0 {
		return errors.New("role approver wajib dipilih")
	}
	if input.UserID <= 0 {
		return errors.New("user approver wajib dipilih")
	}

	exists, err := s.Repo.ExistsByID(input.ID)
	if err != nil {
		return err
	}
	if !exists {
		return errors.New("store approver tidak ditemukan")
	}

	if err := s.validateReferences(input.StoreID, input.RoleID, input.UserID); err != nil {
		return err
	}

	exists, err = s.Repo.ExistsDuplicate(input.StoreID, input.RoleID, input.UserID, input.ID)
	if err != nil {
		return err
	}
	if exists {
		return errors.New("mapping store approver yang sama sudah ada")
	}

	return s.Repo.Update(input)
}

func (s *StoreApproverService) DeleteStoreApprover(id int64) error {
	if id <= 0 {
		return errors.New("store approver tidak valid")
	}
	return s.Repo.DeleteByID(id)
}

func (s *StoreApproverService) validateReferences(storeID int, roleID int64, userID int) error {
	exists, err := s.Repo.ExistsStore(storeID)
	if err != nil {
		return err
	}
	if !exists {
		return errors.New("store tidak ditemukan")
	}

	exists, err = s.Repo.ExistsRole(roleID)
	if err != nil {
		return err
	}
	if !exists {
		return errors.New("role approver tidak ditemukan")
	}

	exists, err = s.Repo.ExistsUser(userID)
	if err != nil {
		return err
	}
	if !exists {
		return errors.New("user approver tidak ditemukan")
	}

	return nil
}
