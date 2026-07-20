package services

import (
	"errors"
	"gobase-app/models"
	"gobase-app/repositories"
	"strings"
	"time"
)

type AssetDisposalService struct {
	Repo *repositories.AssetDisposalRepository
}

func (s *AssetDisposalService) GetPostedDisposalByAssetID(assetID int64) (*models.AssetDisposal, error) {
	if assetID <= 0 {
		return nil, errors.New("aset tidak valid")
	}
	return s.Repo.GetPostedDisposalByAssetID(assetID)
}

func (s *AssetDisposalService) GetDisposalTypes() ([]models.AssetDisposalType, error) {
	return s.Repo.GetDisposalTypes()
}

func (s *AssetDisposalService) SaveDisposalType(input models.AssetDisposalTypeInput) error {
	input.Code = strings.ToUpper(strings.TrimSpace(input.Code))
	input.Name = strings.TrimSpace(input.Name)
	input.Description = strings.TrimSpace(input.Description)
	if input.Code == "" || input.Name == "" {
		return errors.New("kode dan nama jenis disposal wajib diisi")
	}
	if len(input.Code) > 40 || len(input.Name) > 100 || len(input.Description) > 255 {
		return errors.New("data jenis disposal melebihi batas karakter")
	}
	return s.Repo.SaveDisposalType(input)
}

func (s *AssetDisposalService) DeleteDisposalType(id int64) error {
	if id <= 0 {
		return errors.New("jenis disposal tidak valid")
	}
	return s.Repo.DeleteDisposalType(id)
}

func (s *AssetDisposalService) GetDisposals(filter models.AssetDisposalFilter) (models.AssetDisposalResult, error) {
	filter.Status = strings.ToUpper(strings.TrimSpace(filter.Status))
	if filter.Status == "" {
		filter.Status = "ALL"
	}
	validStatuses := map[string]bool{"ALL": true, "DRAFT": true, "IN_APPROVAL": true, "REJECTED": true, "APPROVED": true, "POSTED": true, "CANCELLED": true, "REVERSED": true}
	if !validStatuses[filter.Status] {
		return models.AssetDisposalResult{}, errors.New("status disposal tidak valid")
	}
	filter.Search = strings.TrimSpace(filter.Search)
	if filter.Page < 1 {
		filter.Page = 1
	}
	if filter.PerPage <= 0 || filter.PerPage > 100 {
		filter.PerPage = 50
	}
	return s.Repo.GetDisposals(filter)
}

func (s *AssetDisposalService) GetDisposalAssetOptions() ([]models.AssetDisposalAssetOption, error) {
	return s.Repo.GetDisposalAssetOptions()
}

func (s *AssetDisposalService) SaveDisposal(input models.AssetDisposalInput) error {
	input.BuyerName = strings.TrimSpace(input.BuyerName)
	input.DocumentReference = strings.TrimSpace(input.DocumentReference)
	input.Reason = strings.TrimSpace(input.Reason)
	input.Notes = strings.TrimSpace(input.Notes)
	if input.AssetID <= 0 || input.DisposalTypeID <= 0 {
		return errors.New("aset dan jenis disposal wajib dipilih")
	}
	if input.DisposalValue < 0 {
		return errors.New("nilai disposal tidak boleh negatif")
	}
	if input.Reason == "" {
		return errors.New("alasan disposal wajib diisi")
	}
	if len(input.Reason) > 2000 || len(input.Notes) > 2000 || len(input.BuyerName) > 150 || len(input.DocumentReference) > 100 {
		return errors.New("data disposal melebihi batas karakter")
	}
	date, err := time.Parse("2006-01-02", strings.TrimSpace(input.DisposalDate))
	if err != nil {
		return errors.New("tanggal disposal tidak valid")
	}
	if date.Year() < 2000 || date.Year() > 2200 {
		return errors.New("tahun disposal tidak valid")
	}
	input.DisposalDate = date.Format("2006-01-02")
	if input.AuditContext.ActorUserID <= 0 {
		return errors.New("pengguna tidak valid")
	}
	return s.Repo.SaveDisposal(input)
}

func (s *AssetDisposalService) PostDisposal(id int64, auditCtx models.AuditContext) error {
	if id <= 0 {
		return errors.New("transaksi disposal tidak valid")
	}
	if auditCtx.ActorUserID <= 0 {
		return errors.New("pengguna tidak valid")
	}
	return s.Repo.PostDisposal(id, auditCtx)
}

func (s *AssetDisposalService) CancelDisposal(id int64, reason string, auditCtx models.AuditContext) error {
	reason = strings.TrimSpace(reason)
	if id <= 0 {
		return errors.New("transaksi disposal tidak valid")
	}
	if reason == "" {
		return errors.New("alasan pembatalan disposal wajib diisi")
	}
	if len(reason) > 1000 {
		return errors.New("alasan pembatalan maksimal 1000 karakter")
	}
	if auditCtx.ActorUserID <= 0 {
		return errors.New("pengguna tidak valid")
	}
	return s.Repo.CancelDisposal(id, reason, auditCtx)
}

func (s *AssetDisposalService) ReverseDisposal(id int64, reason string, auditCtx models.AuditContext) error {
	reason = strings.TrimSpace(reason)
	if id <= 0 { return errors.New("transaksi disposal tidak valid") }
	if reason == "" { return errors.New("alasan reversal wajib diisi") }
	if len(reason) > 1000 { return errors.New("alasan reversal maksimal 1000 karakter") }
	if auditCtx.ActorUserID <= 0 { return errors.New("pengguna tidak valid") }
	return s.Repo.ReverseDisposal(id, reason, auditCtx)
}
