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

func (s *PurchaseRequestService) GetPurchaseRequests(scope models.AccessScope) ([]models.PurchaseRequest, error) {
	if scope.UserID <= 0 {
		return nil, errors.New("user login tidak valid")
	}
	return s.Repo.GetAll(scope)
}

func (s *PurchaseRequestService) GetPurchaseRequestDetail(id int64, scope models.AccessScope) (*models.PurchaseRequestDetail, error) {
	if id <= 0 {
		return nil, errors.New("purchase request tidak valid")
	}
	if scope.UserID <= 0 {
		return nil, errors.New("user login tidak valid")
	}
	return s.Repo.GetDetailByID(id, scope)
}

func (s *PurchaseRequestService) CreatePurchaseRequest(input models.PurchaseRequestCreateInput) error {
	input = normalizePRCreateInput(input)
	spendType, err := s.Repo.GLAccountSpendType(input.GLAccountID)
	if err != nil {
		return err
	}
	input.SpendType = strings.ToUpper(strings.TrimSpace(spendType))
	if err := s.validateInput(input); err != nil {
		return err
	}
	totalAmount := calculatePRTotal(input.Items)
	if err := s.validateBudgetException(input, totalAmount); err != nil {
		return err
	}

	_, err = s.Repo.Create(input, totalAmount)
	return err
}

func (s *PurchaseRequestService) GetFormCheck(storeID, divisionID, glAccountID int, neededDate string, amount float64, urgentLevel string) (*models.PurchaseRequestFormCheck, error) {
	if storeID <= 0 || glAccountID <= 0 || amount < 0 {
		return nil, errors.New("parameter pengecekan PR tidak valid")
	}
	urgentLevel = strings.ToUpper(strings.TrimSpace(urgentLevel))
	if urgentLevel == "" {
		urgentLevel = "NORMAL"
	}
	return s.Repo.GetFormCheck(storeID, divisionID, glAccountID, neededDate, amount, urgentLevel)
}

func (s *PurchaseRequestService) RegenerateApprovalFlow(prID int64, scope models.AccessScope, auditCtx models.AuditContext) error {
	if prID <= 0 {
		return errors.New("purchase request tidak valid")
	}
	if auditCtx.ActorUserID <= 0 {
		return errors.New("user login tidak valid")
	}
	if scope.UserID != auditCtx.ActorUserID {
		return errors.New("scope akses user tidak valid")
	}
	return s.Repo.RegenerateApprovalFlow(prID, scope, auditCtx)
}

func (s *PurchaseRequestService) UpdatePurchaseRequest(input models.PurchaseRequestUpdateInput) error {
	if input.ID <= 0 {
		return errors.New("purchase request tidak valid")
	}
	if input.AccessScope.UserID <= 0 || input.AccessScope.UserID != input.AuditContext.ActorUserID {
		return errors.New("scope akses user tidak valid")
	}

	createLike := models.PurchaseRequestCreateInput{
		StoreID:                    input.StoreID,
		DivisionID:                 input.DivisionID,
		GLAccountID:                input.GLAccountID,
		SpendType:                  strings.ToUpper(strings.TrimSpace(input.SpendType)),
		UrgentLevel:                strings.ToUpper(strings.TrimSpace(input.UrgentLevel)),
		NeededDate:                 strings.TrimSpace(input.NeededDate),
		Justification:              strings.TrimSpace(input.Justification),
		RequestTitle:               input.RequestTitle,
		RequestCategory:            input.RequestCategory,
		DeliveryLocation:           input.DeliveryLocation,
		ImpactIfNotApproved:        input.ImpactIfNotApproved,
		UrgencyReason:              input.UrgencyReason,
		RecommendedVendorID:        input.RecommendedVendorID,
		VendorRecommendationReason: input.VendorRecommendationReason,
		IsSingleSource:             input.IsSingleSource,
		SingleSourceReason:         input.SingleSourceReason,
		BudgetExceptionReason:      input.BudgetExceptionReason,
		AssetRequestType:           input.AssetRequestType,
		ExistingAssetCode:          input.ExistingAssetCode,
		AssetLocation:              input.AssetLocation,
		AssetPIC:                   input.AssetPIC,
		Action:                     "draft",
		Items:                      input.Items,
	}

	createLike = normalizePRCreateInput(createLike)
	spendType, err := s.Repo.GLAccountSpendType(createLike.GLAccountID)
	if err != nil {
		return err
	}
	createLike.SpendType = strings.ToUpper(strings.TrimSpace(spendType))
	if err := s.validateEditableInput(createLike); err != nil {
		return err
	}
	input.RequestTitle = createLike.RequestTitle
	input.RequestCategory = createLike.RequestCategory
	input.StoreID = createLike.StoreID
	input.DivisionID = createLike.DivisionID
	input.GLAccountID = createLike.GLAccountID
	input.SpendType = createLike.SpendType
	input.UrgentLevel = createLike.UrgentLevel
	input.NeededDate = createLike.NeededDate
	input.DeliveryLocation = createLike.DeliveryLocation
	input.Justification = createLike.Justification
	input.ImpactIfNotApproved = createLike.ImpactIfNotApproved
	input.UrgencyReason = createLike.UrgencyReason
	input.RecommendedVendorID = createLike.RecommendedVendorID
	input.VendorRecommendationReason = createLike.VendorRecommendationReason
	input.IsSingleSource = createLike.IsSingleSource
	input.SingleSourceReason = createLike.SingleSourceReason
	input.BudgetExceptionReason = createLike.BudgetExceptionReason
	input.AssetRequestType = createLike.AssetRequestType
	input.ExistingAssetCode = createLike.ExistingAssetCode
	input.AssetLocation = createLike.AssetLocation
	input.AssetPIC = createLike.AssetPIC
	input.Items = createLike.Items
	totalAmount := calculatePRTotal(input.Items)

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
	if strings.TrimSpace(input.RequestTitle) == "" {
		return errors.New("judul PR wajib diisi")
	}
	validCategories := map[string]bool{"GOODS": true, "SERVICE": true, "MAINTENANCE": true, "RENTAL": true, "PROJECT": true, "ASSET_REPLACEMENT": true}
	if !validCategories[input.RequestCategory] {
		return errors.New("kategori kebutuhan tidak valid")
	}
	if strings.TrimSpace(input.NeededDate) != "" {
		parsedDate, err := time.Parse("2006-01-02", input.NeededDate)
		if err != nil {
			return errors.New("needed date tidak valid")
		}
		if input.Action == "submit" {
			today := time.Now()
			today = time.Date(today.Year(), today.Month(), today.Day(), 0, 0, 0, 0, today.Location())
			if parsedDate.Before(today) {
				return errors.New("tanggal dibutuhkan tidak boleh sebelum hari ini")
			}
		}
	}
	if input.Action == "submit" {
		if strings.TrimSpace(input.NeededDate) == "" {
			return errors.New("tanggal dibutuhkan wajib diisi sebelum submit")
		}
		if strings.TrimSpace(input.DeliveryLocation) == "" {
			return errors.New("lokasi penggunaan/pengiriman wajib diisi sebelum submit")
		}
		if strings.TrimSpace(input.Justification) == "" {
			return errors.New("justifikasi wajib diisi sebelum submit")
		}
		if strings.TrimSpace(input.ImpactIfNotApproved) == "" {
			return errors.New("dampak jika tidak disetujui wajib diisi sebelum submit")
		}
	}
	if input.UrgentLevel != "NORMAL" && strings.TrimSpace(input.UrgencyReason) == "" {
		return errors.New("alasan urgency wajib diisi untuk PR urgent/emergency")
	}
	if input.IsSingleSource && strings.TrimSpace(input.SingleSourceReason) == "" {
		return errors.New("alasan single source wajib diisi")
	}
	if input.IsSingleSource && input.RecommendedVendorID <= 0 {
		return errors.New("vendor rekomendasi wajib dipilih untuk pengadaan single source")
	}
	if input.SpendType == "CAPEX" && input.Action == "submit" {
		if input.AssetRequestType != "NEW" && input.AssetRequestType != "REPLACEMENT" {
			return errors.New("jenis permintaan asset wajib dipilih")
		}
		if input.AssetRequestType == "REPLACEMENT" && strings.TrimSpace(input.ExistingAssetCode) == "" {
			return errors.New("kode asset lama wajib diisi untuk penggantian asset")
		}
		if strings.TrimSpace(input.AssetLocation) == "" || strings.TrimSpace(input.AssetPIC) == "" {
			return errors.New("lokasi asset dan PIC wajib diisi untuk CAPEX")
		}
	}

	exists, err := s.Repo.StoreExists(input.StoreID)
	if err != nil {
		return err
	}
	if !exists {
		return errors.New("store tidak ditemukan")
	}
	if input.RecommendedVendorID > 0 {
		exists, err = s.Repo.VendorExists(input.RecommendedVendorID)
		if err != nil {
			return err
		}
		if !exists {
			return errors.New("vendor rekomendasi tidak ditemukan")
		}
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
		if input.Action == "submit" && item.EstUnitPrice <= 0 {
			return fmt.Errorf("estimasi harga pada baris %d harus lebih dari 0 sebelum submit", i+1)
		}
		if input.Action == "submit" && strings.TrimSpace(item.PriceSource) == "" {
			return fmt.Errorf("sumber estimasi harga pada baris %d wajib diisi", i+1)
		}
	}

	return nil
}

func (s *PurchaseRequestService) validateBudgetException(input models.PurchaseRequestCreateInput, totalAmount float64) error {
	if input.Action != "submit" {
		return nil
	}
	check, err := s.Repo.GetBudgetCheck(input.StoreID, input.DivisionID, input.GLAccountID, input.NeededDate, totalAmount)
	if err != nil {
		return err
	}
	if (check.Status == "INSUFFICIENT" || check.Status == "UNBUDGETED") && strings.TrimSpace(input.BudgetExceptionReason) == "" {
		return errors.New("alasan pengecualian budget wajib diisi karena budget tidak tersedia atau tidak mencukupi")
	}
	return nil
}

func normalizePRCreateInput(input models.PurchaseRequestCreateInput) models.PurchaseRequestCreateInput {
	input.RequestTitle = strings.TrimSpace(input.RequestTitle)
	input.RequestCategory = strings.ToUpper(strings.TrimSpace(input.RequestCategory))
	input.UrgentLevel = strings.ToUpper(strings.TrimSpace(input.UrgentLevel))
	input.NeededDate = strings.TrimSpace(input.NeededDate)
	input.DeliveryLocation = strings.TrimSpace(input.DeliveryLocation)
	input.Justification = strings.TrimSpace(input.Justification)
	input.ImpactIfNotApproved = strings.TrimSpace(input.ImpactIfNotApproved)
	input.UrgencyReason = strings.TrimSpace(input.UrgencyReason)
	input.VendorRecommendationReason = strings.TrimSpace(input.VendorRecommendationReason)
	input.SingleSourceReason = strings.TrimSpace(input.SingleSourceReason)
	input.BudgetExceptionReason = strings.TrimSpace(input.BudgetExceptionReason)
	input.AssetRequestType = strings.ToUpper(strings.TrimSpace(input.AssetRequestType))
	input.ExistingAssetCode = strings.TrimSpace(input.ExistingAssetCode)
	input.AssetLocation = strings.TrimSpace(input.AssetLocation)
	input.AssetPIC = strings.TrimSpace(input.AssetPIC)
	input.Action = strings.ToLower(strings.TrimSpace(input.Action))
	for i := range input.Items {
		input.Items[i].ItemName = strings.TrimSpace(input.Items[i].ItemName)
		input.Items[i].UOM = strings.TrimSpace(input.Items[i].UOM)
		input.Items[i].Specification = strings.TrimSpace(input.Items[i].Specification)
		input.Items[i].PriceSource = strings.TrimSpace(input.Items[i].PriceSource)
		input.Items[i].Notes = strings.TrimSpace(input.Items[i].Notes)
	}
	return input
}

func calculatePRTotal(items []models.PurchaseRequestItemInput) float64 {
	total := 0.0
	for _, item := range items {
		total += item.Qty * item.EstUnitPrice
	}
	return total
}
