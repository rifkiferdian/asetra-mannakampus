package services

import (
	"errors"
	"fmt"
	"gobase-app/models"
	"gobase-app/repositories"
	"strings"
)

type GLAccountService struct {
	Repo *repositories.GLAccountRepository
}

func (s *GLAccountService) GetGLAccounts() ([]models.GLAccount, error) {
	return s.Repo.GetAll()
}

func (s *GLAccountService) CreateGLAccount(input models.GLAccountCreateInput) error {
	glCode := strings.TrimSpace(input.GLCode)
	glName := strings.TrimSpace(input.GLName)
	spendType := strings.ToUpper(strings.TrimSpace(input.SpendType))

	if glCode == "" {
		return errors.New("kode GL wajib diisi")
	}
	if glName == "" {
		return errors.New("nama GL wajib diisi")
	}
	if spendType != "OPEX" && spendType != "CAPEX" {
		return errors.New("spend type harus OPEX atau CAPEX")
	}

	exists, err := s.Repo.ExistsByCode(glCode)
	if err != nil {
		return err
	}
	if exists {
		return fmt.Errorf("kode GL %s sudah digunakan", glCode)
	}

	return s.Repo.Create(models.GLAccountCreateInput{
		GLCode:    glCode,
		GLName:    glName,
		SpendType: spendType,
		IsActive:  input.IsActive,
	})
}

func (s *GLAccountService) UpdateGLAccount(input models.GLAccountUpdateInput) error {
	glCode := strings.TrimSpace(input.GLCode)
	glName := strings.TrimSpace(input.GLName)
	spendType := strings.ToUpper(strings.TrimSpace(input.SpendType))

	if input.ID <= 0 {
		return errors.New("GL account tidak valid")
	}
	if glCode == "" {
		return errors.New("kode GL wajib diisi")
	}
	if glName == "" {
		return errors.New("nama GL wajib diisi")
	}
	if spendType != "OPEX" && spendType != "CAPEX" {
		return errors.New("spend type harus OPEX atau CAPEX")
	}

	exists, err := s.Repo.ExistsByID(input.ID)
	if err != nil {
		return err
	}
	if !exists {
		return errors.New("GL account tidak ditemukan")
	}

	exists, err = s.Repo.ExistsByCodeExceptID(glCode, input.ID)
	if err != nil {
		return err
	}
	if exists {
		return fmt.Errorf("kode GL %s sudah digunakan", glCode)
	}

	return s.Repo.Update(models.GLAccountUpdateInput{
		ID:        input.ID,
		GLCode:    glCode,
		GLName:    glName,
		SpendType: spendType,
		IsActive:  input.IsActive,
	})
}

func (s *GLAccountService) DeleteGLAccount(id int) error {
	if id <= 0 {
		return errors.New("GL account tidak valid")
	}
	return s.Repo.DeleteByID(id)
}
