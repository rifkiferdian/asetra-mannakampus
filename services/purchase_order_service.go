package services

import (
	"errors"
	"fmt"
	"gobase-app/models"
	"gobase-app/repositories"
	"strings"
)

type PurchaseOrderService struct {
	Repo       *repositories.PurchaseOrderRepository
	VendorRepo *repositories.VendorRepository
}

func (s *PurchaseOrderService) GetPurchaseOrders() ([]models.PurchaseOrder, error) {
	return s.Repo.GetAll()
}

func (s *PurchaseOrderService) GetApprovedPRReadyForPO() ([]models.ApprovedPRForPO, error) {
	return s.Repo.GetApprovedPRReadyForPO()
}

func (s *PurchaseOrderService) GetCreateForm(prID int64, userID int) (*models.PurchaseOrderCreateForm, error) {
	if prID <= 0 {
		return nil, errors.New("purchase request tidak valid")
	}
	return s.Repo.GetCreateFormByPRID(prID, userID)
}

func (s *PurchaseOrderService) GetPurchaseOrderDetail(id int64) (*models.PurchaseOrderDetail, error) {
	if id <= 0 {
		return nil, errors.New("purchase order tidak valid")
	}
	return s.Repo.GetDetailByID(id)
}

func (s *PurchaseOrderService) CreateFromPR(input models.PurchaseOrderCreateInput) (int64, error) {
	if input.PRID <= 0 {
		return 0, errors.New("purchase request tidak valid")
	}
	if input.VendorID <= 0 {
		return 0, errors.New("vendor wajib dipilih")
	}
	if input.AuditContext.ActorUserID <= 0 {
		return 0, errors.New("user login tidak valid")
	}

	exists, err := s.VendorRepo.ExistsByID(input.VendorID)
	if err != nil {
		return 0, err
	}
	if !exists {
		return 0, errors.New("vendor tidak ditemukan")
	}

	form, err := s.Repo.GetCreateFormByPRID(input.PRID, 0)
	if err != nil {
		return 0, err
	}

	priceByPRItemID := make(map[int64]float64)
	for _, item := range input.Items {
		if item.PRItemID <= 0 {
			return 0, errors.New("item PR tidak valid")
		}
		if item.UnitPrice < 0 {
			return 0, fmt.Errorf("harga final item tidak boleh negatif")
		}
		priceByPRItemID[item.PRItemID] = item.UnitPrice
	}

	normalizedItems := make([]models.PurchaseOrderItemInput, 0, len(form.PR.Items))
	totalAmount := 0.0
	for _, prItem := range form.PR.Items {
		unitPrice, ok := priceByPRItemID[prItem.ID]
		if !ok {
			return 0, fmt.Errorf("harga final untuk item %s wajib diisi", prItem.ItemName)
		}
		normalized := models.PurchaseOrderItemInput{
			PRItemID:  prItem.ID,
			ItemName:  strings.TrimSpace(prItem.ItemName),
			Qty:       prItem.Qty,
			UOM:       strings.TrimSpace(prItem.UOM),
			UnitPrice: unitPrice,
		}
		totalAmount += normalized.Qty * normalized.UnitPrice
		normalizedItems = append(normalizedItems, normalized)
	}

	input.Items = normalizedItems
	return s.Repo.CreateFromPR(input, totalAmount)
}
