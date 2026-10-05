package services

import (
	"errors"
	"fmt"
	"gobase-app/models"
	"gobase-app/repositories"
	"strings"
)

type PurchaseOrderService struct {
	Repo                   *repositories.PurchaseOrderRepository
	VendorRepo             *repositories.VendorRepository
	POVarianceToleranceBPS int64
}

func (s *PurchaseOrderService) GetPurchaseOrders(scope models.AccessScope) ([]models.PurchaseOrder, error) {
	if scope.UserID <= 0 {
		return nil, errors.New("user login tidak valid")
	}
	return s.Repo.GetAll(scope)
}

func (s *PurchaseOrderService) GetApprovedPRReadyForPO(scope models.AccessScope) ([]models.ApprovedPRForPO, error) {
	if scope.UserID <= 0 {
		return nil, errors.New("user login tidak valid")
	}
	return s.Repo.GetApprovedPRReadyForPO(scope)
}

func (s *PurchaseOrderService) GetCreateForm(prID int64, scope models.AccessScope) (*models.PurchaseOrderCreateForm, error) {
	if prID <= 0 {
		return nil, errors.New("purchase request tidak valid")
	}
	return s.Repo.GetCreateFormByPRID(prID, scope)
}

func (s *PurchaseOrderService) GetPurchaseOrderDetail(id int64, scope models.AccessScope) (*models.PurchaseOrderDetail, error) {
	if id <= 0 {
		return nil, errors.New("purchase order tidak valid")
	}
	return s.Repo.GetDetailByID(id, scope)
}

func (s *PurchaseOrderService) CreateFromPR(input models.PurchaseOrderCreateInput) (int64, error) {
	input.VarianceToleranceBPS = s.POVarianceToleranceBPS
	if input.PRID <= 0 {
		return 0, errors.New("purchase request tidak valid")
	}
	if input.VendorID <= 0 {
		return 0, errors.New("vendor wajib dipilih")
	}
	if input.AuditContext.ActorUserID <= 0 {
		return 0, errors.New("user login tidak valid")
	}
	if input.AccessScope.UserID != input.AuditContext.ActorUserID {
		return 0, errors.New("scope akses user tidak valid")
	}

	exists, err := s.VendorRepo.ExistsByID(input.VendorID)
	if err != nil {
		return 0, err
	}
	if !exists {
		return 0, errors.New("vendor tidak ditemukan")
	}

	form, err := s.Repo.GetCreateFormByPRID(input.PRID, input.AccessScope)
	if err != nil {
		return 0, err
	}

	priceByPRItemID := make(map[int64]models.Money)
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
	var totalAmount models.Money
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
		lineTotal, err := models.MultiplyMoneyByQuantity(normalized.UnitPrice, normalized.Qty)
		if err != nil {
			return 0, fmt.Errorf("total item %s tidak valid: %w", normalized.ItemName, err)
		}
		totalAmount = totalAmount.Add(lineTotal)
		normalizedItems = append(normalizedItems, normalized)
	}
	if models.ExceedsVariance(totalAmount, form.PR.TotalAmount, input.VarianceToleranceBPS) {
		return 0, fmt.Errorf("nilai PO %s melebihi toleransi terhadap nilai PR %s; lakukan approval ulang", totalAmount.FormatIDR(), form.PR.TotalAmount.FormatIDR())
	}

	input.Items = normalizedItems
	return s.Repo.CreateFromPR(input, totalAmount)
}
