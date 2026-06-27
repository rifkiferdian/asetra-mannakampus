package services

import (
	"errors"
	"fmt"
	"gobase-app/models"
	"gobase-app/repositories"
	"strings"
	"time"
)

type PurchaseRequestService struct {
	Repo         *repositories.PurchaseRequestRepository
	DivisionRepo *repositories.DivisionRepository
	StoreRepo    *repositories.StoreRepository
	GlRepo       *repositories.GLAccountRepository
}

func (s *PurchaseRequestService) GetPurchaseRequests() ([]models.PurchaseRequest, error) {
	return s.Repo.GetAll()
}

func (s *PurchaseRequestService) GetPurchaseRequestDetail(id int64, userID int) (*models.PurchaseRequestDetail, error) {
	if id <= 0 {
		return nil, errors.New("purchase request tidak valid")
	}
	return s.Repo.GetDetailByID(id, userID)
}

func (s *PurchaseRequestService) CreatePurchaseRequest(input models.PurchaseRequestCreateInput) error {
	if err := s.validateInput(input); err != nil {
		return err
	}

	totalAmount := 0.0
	normalizedItems := make([]models.PurchaseRequestItemInput, 0, len(input.Items))
	for _, item := range input.Items {
		normalized := models.PurchaseRequestItemInput{
			ItemName:     strings.TrimSpace(item.ItemName),
			Qty:          item.Qty,
			UOM:          strings.TrimSpace(item.UOM),
			EstUnitPrice: item.EstUnitPrice,
			Notes:        strings.TrimSpace(item.Notes),
		}
		totalAmount += normalized.Qty * normalized.EstUnitPrice
		normalizedItems = append(normalizedItems, normalized)
	}

	input.Items = normalizedItems
	input.Justification = strings.TrimSpace(input.Justification)
	input.SpendType = strings.ToUpper(strings.TrimSpace(input.SpendType))
	input.UrgentLevel = strings.ToUpper(strings.TrimSpace(input.UrgentLevel))
	input.Action = strings.ToLower(strings.TrimSpace(input.Action))

	_, err := s.Repo.Create(input, totalAmount)
	return err
}

func (s *PurchaseRequestService) RegenerateApprovalFlow(prID int64, auditCtx models.AuditContext) error {
	if prID <= 0 {
		return errors.New("purchase request tidak valid")
	}
	if auditCtx.ActorUserID <= 0 {
		return errors.New("user login tidak valid")
	}
	return s.Repo.RegenerateApprovalFlow(prID, auditCtx)
}

func (s *PurchaseRequestService) UpdatePurchaseRequest(input models.PurchaseRequestUpdateInput) error {
	if input.ID <= 0 {
		return errors.New("purchase request tidak valid")
	}

	createLike := models.PurchaseRequestCreateInput{
		StoreID:       input.StoreID,
		DivisionID:    input.DivisionID,
		GLAccountID:   input.GLAccountID,
		SpendType:     strings.ToUpper(strings.TrimSpace(input.SpendType)),
		UrgentLevel:   strings.ToUpper(strings.TrimSpace(input.UrgentLevel)),
		NeededDate:    strings.TrimSpace(input.NeededDate),
		Justification: strings.TrimSpace(input.Justification),
		Action:        "draft",
		Items:         input.Items,
	}

	if err := s.validateEditableInput(createLike); err != nil {
		return err
	}

	totalAmount := 0.0
	normalizedItems := make([]models.PurchaseRequestItemInput, 0, len(input.Items))
	for _, item := range input.Items {
		normalized := models.PurchaseRequestItemInput{
			ItemName:     strings.TrimSpace(item.ItemName),
			Qty:          item.Qty,
			UOM:          strings.TrimSpace(item.UOM),
			EstUnitPrice: item.EstUnitPrice,
			Notes:        strings.TrimSpace(item.Notes),
		}
		totalAmount += normalized.Qty * normalized.EstUnitPrice
		normalizedItems = append(normalizedItems, normalized)
	}

	input.Items = normalizedItems
	input.SpendType = strings.ToUpper(strings.TrimSpace(input.SpendType))
	input.UrgentLevel = strings.ToUpper(strings.TrimSpace(input.UrgentLevel))
	input.NeededDate = strings.TrimSpace(input.NeededDate)
	input.Justification = strings.TrimSpace(input.Justification)

	return s.Repo.UpdateEditable(input, totalAmount)
}

func (s *PurchaseRequestService) validateInput(input models.PurchaseRequestCreateInput) error {
	if input.RequesterUserID <= 0 {
		return errors.New("requester tidak valid")
	}
	exists, err := s.Repo.UserExists(input.RequesterUserID)
	if err != nil {
		return err
	}
	if !exists {
		return errors.New("requester tidak ditemukan")
	}
	return s.validateEditableInput(input)
}

func (s *PurchaseRequestService) validateEditableInput(input models.PurchaseRequestCreateInput) error {
	if input.StoreID <= 0 {
		return errors.New("store wajib dipilih")
	}
	if input.GLAccountID <= 0 {
		return errors.New("GL account wajib dipilih")
	}
	if input.SpendType != "OPEX" && input.SpendType != "CAPEX" {
		return errors.New("spend type harus OPEX atau CAPEX")
	}
	if input.UrgentLevel != "NORMAL" && input.UrgentLevel != "URGENT" && input.UrgentLevel != "EMERGENCY" {
		return errors.New("urgent level tidak valid")
	}
	if input.Action != "draft" && input.Action != "submit" {
		return errors.New("aksi PR tidak valid")
	}
	if len(input.Items) == 0 {
		return errors.New("minimal harus ada 1 item PR")
	}
	if strings.TrimSpace(input.NeededDate) != "" {
		if _, err := time.Parse("2006-01-02", input.NeededDate); err != nil {
			return errors.New("needed date tidak valid")
		}
	}

	exists, err := s.Repo.StoreExists(input.StoreID)
	if err != nil {
		return err
	}
	if !exists {
		return errors.New("store tidak ditemukan")
	}

	if input.DivisionID > 0 {
		exists, err = s.DivisionRepo.ExistsByID(input.DivisionID)
		if err != nil {
			return err
		}
		if !exists {
			return errors.New("division tidak ditemukan")
		}
	}

	exists, err = s.Repo.GLAccountExists(input.GLAccountID)
	if err != nil {
		return err
	}
	if !exists {
		return errors.New("GL account tidak ditemukan")
	}

	glSpendType, err := s.Repo.GLAccountSpendType(input.GLAccountID)
	if err != nil {
		return err
	}
	if strings.ToUpper(strings.TrimSpace(glSpendType)) != input.SpendType {
		return fmt.Errorf("spend type PR harus sama dengan spend type GL account (%s)", glSpendType)
	}

	for i, item := range input.Items {
		name := strings.TrimSpace(item.ItemName)
		uom := strings.TrimSpace(item.UOM)
		if name == "" {
			return fmt.Errorf("nama item pada baris %d wajib diisi", i+1)
		}
		if item.Qty <= 0 {
			return fmt.Errorf("qty pada baris %d harus lebih dari 0", i+1)
		}
		if uom == "" {
			return fmt.Errorf("uom pada baris %d wajib diisi", i+1)
		}
		if item.EstUnitPrice < 0 {
			return fmt.Errorf("estimasi harga pada baris %d tidak boleh negatif", i+1)
		}
	}

	return nil
}
