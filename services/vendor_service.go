package services

import (
	"errors"
	"fmt"
	"gobase-app/models"
	"gobase-app/repositories"
	"net/mail"
	"strings"
)

type VendorService struct {
	Repo *repositories.VendorRepository
}

func (s *VendorService) GetVendors() ([]models.Vendor, error) {
	return s.Repo.GetAll()
}

func (s *VendorService) CreateVendor(input models.VendorCreateInput) error {
	name := strings.TrimSpace(input.Name)
	phone := strings.TrimSpace(input.Phone)
	email := strings.TrimSpace(input.Email)
	address := strings.TrimSpace(input.Address)

	if name == "" {
		return errors.New("nama vendor wajib diisi")
	}
	if email != "" {
		if _, err := mail.ParseAddress(email); err != nil {
			return errors.New("email vendor tidak valid")
		}
	}

	exists, err := s.Repo.ExistsByName(name)
	if err != nil {
		return err
	}
	if exists {
		return fmt.Errorf("vendor %s sudah terdaftar", name)
	}

	return s.Repo.Create(models.VendorCreateInput{
		Name:     name,
		Phone:    phone,
		Email:    email,
		Address:  address,
		IsActive: input.IsActive,
	})
}

func (s *VendorService) UpdateVendor(input models.VendorUpdateInput) error {
	name := strings.TrimSpace(input.Name)
	phone := strings.TrimSpace(input.Phone)
	email := strings.TrimSpace(input.Email)
	address := strings.TrimSpace(input.Address)

	if input.ID <= 0 {
		return errors.New("vendor tidak valid")
	}
	if name == "" {
		return errors.New("nama vendor wajib diisi")
	}
	if email != "" {
		if _, err := mail.ParseAddress(email); err != nil {
			return errors.New("email vendor tidak valid")
		}
	}

	exists, err := s.Repo.ExistsByID(input.ID)
	if err != nil {
		return err
	}
	if !exists {
		return errors.New("vendor tidak ditemukan")
	}

	exists, err = s.Repo.ExistsByNameExceptID(name, input.ID)
	if err != nil {
		return err
	}
	if exists {
		return fmt.Errorf("vendor %s sudah terdaftar", name)
	}

	return s.Repo.Update(models.VendorUpdateInput{
		ID:       input.ID,
		Name:     name,
		Phone:    phone,
		Email:    email,
		Address:  address,
		IsActive: input.IsActive,
	})
}

func (s *VendorService) DeleteVendor(id int64) error {
	if id <= 0 {
		return errors.New("vendor tidak valid")
	}
	return s.Repo.DeleteByID(id)
}
