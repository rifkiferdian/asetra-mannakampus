package services

import (
	"errors"
	"fmt"
	"gobase-app/models"
	"gobase-app/repositories"
	"strings"
)

type BudgetService struct {
	Repo         *repositories.BudgetRepository
	StoreRepo    *repositories.StoreRepository
	DivisionRepo *repositories.DivisionRepository
	GLRepo       *repositories.GLAccountRepository
}

func (s *BudgetService) GetBudgets() ([]models.Budget, error) {
	return s.Repo.GetAll()
}

func (s *BudgetService) CreateBudget(input models.BudgetCreateInput) error {
	normalized, err := s.normalizeCreateInput(input)
	if err != nil {
		return err
	}

	exists, err := s.Repo.ExistsDuplicate(normalized)
	if err != nil {
		return err
	}
	if exists {
		return errors.New("budget untuk kombinasi periode/store/division/GL tersebut sudah ada")
	}

	return s.Repo.Create(normalized)
}

func (s *BudgetService) UpdateBudget(input models.BudgetUpdateInput) error {
	if input.ID <= 0 {
		return errors.New("budget tidak valid")
	}

	exists, err := s.Repo.ExistsByID(input.ID)
	if err != nil {
		return err
	}
	if !exists {
		return errors.New("budget tidak ditemukan")
	}

	createInput, err := s.normalizeCreateInput(models.BudgetCreateInput{
		FiscalYear:  input.FiscalYear,
		PeriodType:  input.PeriodType,
		PeriodKey:   input.PeriodKey,
		StoreID:     input.StoreID,
		DivisionID:  input.DivisionID,
		GLAccountID: input.GLAccountID,
		Amount:      input.Amount,
	})
	if err != nil {
		return err
	}

	normalized := models.BudgetUpdateInput{
		ID:          input.ID,
		FiscalYear:  createInput.FiscalYear,
		PeriodType:  createInput.PeriodType,
		PeriodKey:   createInput.PeriodKey,
		StoreID:     createInput.StoreID,
		DivisionID:  createInput.DivisionID,
		GLAccountID: createInput.GLAccountID,
		Amount:      createInput.Amount,
	}

	exists, err = s.Repo.ExistsDuplicateExceptID(normalized)
	if err != nil {
		return err
	}
	if exists {
		return errors.New("budget untuk kombinasi periode/store/division/GL tersebut sudah ada")
	}

	return s.Repo.Update(normalized)
}

func (s *BudgetService) DeleteBudget(id int64) error {
	if id <= 0 {
		return errors.New("budget tidak valid")
	}
	return s.Repo.DeleteByID(id)
}

func (s *BudgetService) normalizeCreateInput(input models.BudgetCreateInput) (models.BudgetCreateInput, error) {
	periodType := strings.ToUpper(strings.TrimSpace(input.PeriodType))
	periodKey := strings.TrimSpace(input.PeriodKey)

	if input.FiscalYear < 2000 || input.FiscalYear > 2100 {
		return models.BudgetCreateInput{}, errors.New("fiscal year tidak valid")
	}
	if periodType != "MONTHLY" && periodType != "QUARTERLY" && periodType != "YEARLY" {
		return models.BudgetCreateInput{}, errors.New("period type harus MONTHLY, QUARTERLY, atau YEARLY")
	}
	if periodKey == "" {
		return models.BudgetCreateInput{}, errors.New("period key wajib diisi")
	}
	if input.GLAccountID <= 0 {
		return models.BudgetCreateInput{}, errors.New("GL account wajib dipilih")
	}
	if input.Amount < 0 {
		return models.BudgetCreateInput{}, errors.New("amount tidak boleh negatif")
	}

	if input.StoreID > 0 {
		exists, err := s.StoreRepo.ExistsByID(input.StoreID)
		if err != nil {
			return models.BudgetCreateInput{}, err
		}
		if !exists {
			return models.BudgetCreateInput{}, fmt.Errorf("store id %d tidak ditemukan", input.StoreID)
		}
	}

	if input.DivisionID > 0 {
		exists, err := s.DivisionRepo.ExistsByID(input.DivisionID)
		if err != nil {
			return models.BudgetCreateInput{}, err
		}
		if !exists {
			return models.BudgetCreateInput{}, fmt.Errorf("division id %d tidak ditemukan", input.DivisionID)
		}
	}

	exists, err := s.GLRepo.ExistsByID(input.GLAccountID)
	if err != nil {
		return models.BudgetCreateInput{}, err
	}
	if !exists {
		return models.BudgetCreateInput{}, errors.New("GL account tidak ditemukan")
	}

	return models.BudgetCreateInput{
		FiscalYear:  input.FiscalYear,
		PeriodType:  periodType,
		PeriodKey:   periodKey,
		StoreID:     input.StoreID,
		DivisionID:  input.DivisionID,
		GLAccountID: input.GLAccountID,
		Amount:      input.Amount,
	}, nil
}
