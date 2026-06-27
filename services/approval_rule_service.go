package services

import (
	"errors"
	"fmt"
	"gobase-app/models"
	"gobase-app/repositories"
	"sort"
	"strings"
)

type ApprovalRuleService struct {
	Repo *repositories.ApprovalRuleRepository
}

func (s *ApprovalRuleService) GetApprovalRules() ([]models.ApprovalRule, error) {
	return s.Repo.GetAll()
}

func (s *ApprovalRuleService) GetApprovalRuleDetail(id int64) (*models.ApprovalRuleDetail, error) {
	if id <= 0 {
		return nil, errors.New("approval rule tidak valid")
	}
	return s.Repo.GetByID(id)
}

func (s *ApprovalRuleService) CreateApprovalRule(input models.ApprovalRuleCreateInput) error {
	if err := s.validateRuleInput(input.Name, input.MinAmount, input.MaxAmount, input.LocationScope, input.SpendType, input.UrgentLevel, input.Steps, 0); err != nil {
		return err
	}

	exists, err := s.Repo.ExistsByName(strings.TrimSpace(input.Name))
	if err != nil {
		return err
	}
	if exists {
		return fmt.Errorf("nama rule %s sudah digunakan", strings.TrimSpace(input.Name))
	}

	normalized, err := s.normalizeInput(input.Name, input.IsActive, input.MinAmount, input.MaxAmount, input.LocationScope, input.SpendType, input.UrgentLevel, input.Steps)
	if err != nil {
		return err
	}

	_, err = s.Repo.Create(normalized)
	return err
}

func (s *ApprovalRuleService) UpdateApprovalRule(input models.ApprovalRuleUpdateInput) error {
	if input.ID <= 0 {
		return errors.New("approval rule tidak valid")
	}

	exists, err := s.Repo.ExistsByID(input.ID)
	if err != nil {
		return err
	}
	if !exists {
		return errors.New("approval rule tidak ditemukan")
	}

	if err := s.validateRuleInput(input.Name, input.MinAmount, input.MaxAmount, input.LocationScope, input.SpendType, input.UrgentLevel, input.Steps, input.ID); err != nil {
		return err
	}

	exists, err = s.Repo.ExistsByNameExceptID(strings.TrimSpace(input.Name), input.ID)
	if err != nil {
		return err
	}
	if exists {
		return fmt.Errorf("nama rule %s sudah digunakan", strings.TrimSpace(input.Name))
	}

	normalized, err := s.normalizeInput(input.Name, input.IsActive, input.MinAmount, input.MaxAmount, input.LocationScope, input.SpendType, input.UrgentLevel, input.Steps)
	if err != nil {
		return err
	}

	return s.Repo.Update(models.ApprovalRuleUpdateInput{
		ID:            input.ID,
		Name:          normalized.Name,
		IsActive:      normalized.IsActive,
		MinAmount:     normalized.MinAmount,
		MaxAmount:     normalized.MaxAmount,
		LocationScope: normalized.LocationScope,
		SpendType:     normalized.SpendType,
		UrgentLevel:   normalized.UrgentLevel,
		Steps:         normalized.Steps,
	})
}

func (s *ApprovalRuleService) DeleteApprovalRule(id int64) error {
	if id <= 0 {
		return errors.New("approval rule tidak valid")
	}

	exists, err := s.Repo.ExistsByID(id)
	if err != nil {
		return err
	}
	if !exists {
		return errors.New("approval rule tidak ditemukan")
	}

	usedCount, err := s.Repo.CountApprovalsByRuleID(id)
	if err != nil {
		return err
	}
	if usedCount > 0 {
		return fmt.Errorf("approval rule tidak bisa dihapus karena sudah dipakai oleh %d approval workflow. Nonaktifkan rule jika tidak ingin dipakai untuk PR baru", usedCount)
	}

	return s.Repo.DeleteByID(id)
}

func (s *ApprovalRuleService) validateRuleInput(name string, minAmount float64, maxAmount *float64, locationScope, spendType, urgentLevel string, steps []models.ApprovalRuleStepInput, currentID int64) error {
	name = strings.TrimSpace(name)
	locationScope = strings.ToUpper(strings.TrimSpace(locationScope))
	spendType = strings.ToUpper(strings.TrimSpace(spendType))
	urgentLevel = strings.ToUpper(strings.TrimSpace(urgentLevel))

	if name == "" {
		return errors.New("nama rule wajib diisi")
	}
	if minAmount < 0 {
		return errors.New("minimum amount tidak boleh negatif")
	}
	if maxAmount != nil && *maxAmount < minAmount {
		return errors.New("maximum amount tidak boleh lebih kecil dari minimum amount")
	}
	if locationScope != "STORE" && locationScope != "HO" && locationScope != "ANY" {
		return errors.New("location scope tidak valid")
	}
	if spendType != "OPEX" && spendType != "CAPEX" && spendType != "ANY" {
		return errors.New("spend type tidak valid")
	}
	if urgentLevel != "NORMAL" && urgentLevel != "URGENT" && urgentLevel != "EMERGENCY" && urgentLevel != "ANY" {
		return errors.New("urgent level tidak valid")
	}
	if len(steps) == 0 {
		return errors.New("minimal harus ada 1 approval step")
	}

	for i, step := range steps {
		if step.StepOrder <= 0 {
			return fmt.Errorf("step order baris %d tidak valid", i+1)
		}
		if step.RoleID <= 0 {
			return fmt.Errorf("role approver baris %d wajib dipilih", i+1)
		}
		scope := strings.ToUpper(strings.TrimSpace(step.Scope))
		if scope != "STORE" && scope != "HO" {
			return fmt.Errorf("scope step baris %d tidak valid", i+1)
		}

		exists, err := s.Repo.RoleExists(step.RoleID)
		if err != nil {
			return err
		}
		if !exists {
			return fmt.Errorf("role approver baris %d tidak ditemukan", i+1)
		}
	}

	return nil
}

func (s *ApprovalRuleService) normalizeInput(name string, isActive bool, minAmount float64, maxAmount *float64, locationScope, spendType, urgentLevel string, steps []models.ApprovalRuleStepInput) (models.ApprovalRuleCreateInput, error) {
	normalizedSteps := make([]models.ApprovalRuleStepInput, 0, len(steps))
	for _, step := range steps {
		normalizedSteps = append(normalizedSteps, models.ApprovalRuleStepInput{
			StepOrder:  step.StepOrder,
			RoleID:     step.RoleID,
			Scope:      strings.ToUpper(strings.TrimSpace(step.Scope)),
			IsParallel: step.IsParallel,
			IsRequired: step.IsRequired,
		})
	}

	sort.SliceStable(normalizedSteps, func(i, j int) bool {
		if normalizedSteps[i].StepOrder == normalizedSteps[j].StepOrder {
			return normalizedSteps[i].RoleID < normalizedSteps[j].RoleID
		}
		return normalizedSteps[i].StepOrder < normalizedSteps[j].StepOrder
	})

	return models.ApprovalRuleCreateInput{
		Name:          strings.TrimSpace(name),
		IsActive:      isActive,
		MinAmount:     minAmount,
		MaxAmount:     maxAmount,
		LocationScope: strings.ToUpper(strings.TrimSpace(locationScope)),
		SpendType:     strings.ToUpper(strings.TrimSpace(spendType)),
		UrgentLevel:   strings.ToUpper(strings.TrimSpace(urgentLevel)),
		Steps:         normalizedSteps,
	}, nil
}
