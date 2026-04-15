package services

import (
	"errors"
	"fmt"
	"gobase-app/models"
	"gobase-app/repositories"
	"strings"
)

type StoreService struct {
	Repo *repositories.StoreRepository
}

func (s *StoreService) GetStores() ([]models.Store, error) {
	return s.Repo.GetAll()
}

func (s *StoreService) CreateStore(input models.StoreCreateInput) error {
	storeCode := strings.TrimSpace(input.StoreCode)
	storeName := strings.TrimSpace(input.StoreName)
	storeAddress := strings.TrimSpace(input.StoreAddress)

	if input.StoreID <= 0 {
		return errors.New("store id wajib diisi")
	}
	if storeCode == "" {
		return errors.New("kode store wajib diisi")
	}
	if storeName == "" {
		return errors.New("nama store wajib diisi")
	}
	if storeAddress == "" {
		return errors.New("alamat store wajib diisi")
	}

	exists, err := s.Repo.ExistsByID(input.StoreID)
	if err != nil {
		return err
	}
	if exists {
		return fmt.Errorf("store id %d sudah digunakan", input.StoreID)
	}

	return s.Repo.Create(models.StoreCreateInput{
		StoreID:      input.StoreID,
		StoreCode:    storeCode,
		StoreName:    storeName,
		StoreAddress: storeAddress,
		IsActive:     input.IsActive,
	})
}

func (s *StoreService) UpdateStore(input models.StoreUpdateInput) error {
	storeCode := strings.TrimSpace(input.StoreCode)
	storeName := strings.TrimSpace(input.StoreName)
	storeAddress := strings.TrimSpace(input.StoreAddress)

	if input.StoreID <= 0 {
		return errors.New("store id tidak valid")
	}
	if storeCode == "" {
		return errors.New("kode store wajib diisi")
	}
	if storeName == "" {
		return errors.New("nama store wajib diisi")
	}
	if storeAddress == "" {
		return errors.New("alamat store wajib diisi")
	}

	exists, err := s.Repo.ExistsByID(input.StoreID)
	if err != nil {
		return err
	}
	if !exists {
		return fmt.Errorf("store id %d tidak ditemukan", input.StoreID)
	}

	return s.Repo.Update(models.StoreUpdateInput{
		StoreID:      input.StoreID,
		StoreCode:    storeCode,
		StoreName:    storeName,
		StoreAddress: storeAddress,
		IsActive:     input.IsActive,
	})
}

func (s *StoreService) DeleteStore(id int) error {
	if id <= 0 {
		return errors.New("store id tidak valid")
	}
	return s.Repo.DeleteByID(id)
}
