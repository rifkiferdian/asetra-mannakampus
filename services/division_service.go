package services

import (
	"errors"
	"fmt"
	"gobase-app/models"
	"gobase-app/repositories"
	"strings"
)

type DivisionService struct {
	Repo *repositories.DivisionRepository
}

func (s *DivisionService) GetDivisions() ([]models.Division, error) {
	return s.Repo.GetAll()
}

func (s *DivisionService) CreateDivision(input models.DivisionCreateInput) error {
	code := strings.ToUpper(strings.TrimSpace(input.DivisionCode))
	name := strings.TrimSpace(input.DivisionName)
	if code == "" {
		return errors.New("kode division wajib diisi")
	}
	if name == "" {
		return errors.New("nama division wajib diisi")
	}

	exists, err := s.Repo.ExistsByCode(code)
	if err != nil {
		return err
	}
	if exists {
		return fmt.Errorf("kode division %s sudah digunakan", code)
	}

	return s.Repo.Create(models.DivisionCreateInput{DivisionCode: code, DivisionName: name})
}

func (s *DivisionService) UpdateDivision(input models.DivisionUpdateInput) error {
	code := strings.ToUpper(strings.TrimSpace(input.DivisionCode))
	name := strings.TrimSpace(input.DivisionName)
	if input.ID <= 0 {
		return errors.New("division tidak valid")
	}
	if code == "" {
		return errors.New("kode division wajib diisi")
	}
	if name == "" {
		return errors.New("nama division wajib diisi")
	}

	exists, err := s.Repo.ExistsByID(input.ID)
	if err != nil {
		return err
	}
	if !exists {
		return errors.New("division tidak ditemukan")
	}

	exists, err = s.Repo.ExistsByCodeExceptID(code, input.ID)
	if err != nil {
		return err
	}
	if exists {
		return fmt.Errorf("kode division %s sudah digunakan", code)
	}

	return s.Repo.Update(models.DivisionUpdateInput{ID: input.ID, DivisionCode: code, DivisionName: name})
}

func (s *DivisionService) DeleteDivision(id int) error {
	if id <= 0 {
		return errors.New("division tidak valid")
	}
	return s.Repo.DeleteByID(id)
}
