package services

import (
	"errors"
	"gobase-app/models"
	"gobase-app/repositories"
	"strings"
	"time"
)

type AssetDepreciationService struct {
	Repo *repositories.AssetDepreciationRepository
}

func (s *AssetDepreciationService) GetAssetDepreciationDetail(assetID int64) (models.AssetDepreciationDetail, error) {
	if assetID <= 0 {
		return models.AssetDepreciationDetail{}, errors.New("aset tidak valid")
	}
	return s.Repo.GetAssetDepreciationDetail(assetID)
}

func (s *AssetDepreciationService) GetPostedDepreciationByAssetID(assetID int64, limit int) ([]models.AssetDepreciationPosting, error) {
	if assetID <= 0 {
		return nil, errors.New("aset tidak valid")
	}
	return s.Repo.GetPostedDepreciationByAssetID(assetID, limit)
}

func (s *AssetDepreciationService) GetDepreciationProfiles(filter models.DepreciationProfileFilter) (models.DepreciationProfileResult, error) {
	filter.Status = strings.ToUpper(strings.TrimSpace(filter.Status))
	if filter.Status == "" {
		filter.Status = "ALL"
	}
	if filter.Status != "ALL" && filter.Status != "ACTIVE" && filter.Status != "PAUSED" && filter.Status != "FINISHED" && filter.Status != "TERMINATED" {
		return models.DepreciationProfileResult{}, errors.New("status profil tidak valid")
	}
	filter.Search = strings.TrimSpace(filter.Search)
	if filter.Page < 1 {
		filter.Page = 1
	}
	if filter.PerPage <= 0 || filter.PerPage > 100 {
		filter.PerPage = 50
	}
	return s.Repo.GetDepreciationProfiles(filter)
}

func (s *AssetDepreciationService) GetDepreciationMethods() ([]models.DepreciationMethodOption, error) {
	return s.Repo.GetDepreciationMethods()
}

func (s *AssetDepreciationService) GetFirstMonthPolicies() ([]models.DepreciationPolicyOption, error) {
	return s.Repo.GetFirstMonthPolicies()
}

func (s *AssetDepreciationService) GetLastMonthPolicies() ([]models.DepreciationPolicyOption, error) {
	return s.Repo.GetLastMonthPolicies()
}

func (s *AssetDepreciationService) GetDepreciationAssetOptions() ([]models.DepreciationAssetOption, error) {
	return s.Repo.GetDepreciationAssetOptions()
}

func (s *AssetDepreciationService) SaveDepreciationProfile(input models.DepreciationProfileInput) error {
	input.Status = strings.ToUpper(strings.TrimSpace(input.Status))
	input.Notes = strings.TrimSpace(input.Notes)
	input.StartDate = strings.TrimSpace(input.StartDate)
	if input.AssetID <= 0 || input.MethodID <= 0 {
		return errors.New("aset dan metode depresiasi wajib dipilih")
	}
	if input.FirstMonthPolicyID <= 0 || input.LastMonthPolicyID <= 0 {
		return errors.New("kebijakan bulan pertama dan bulan terakhir wajib dipilih")
	}
	if input.AuditContext.ActorUserID <= 0 {
		return errors.New("pengguna tidak valid")
	}
	if input.DepreciableBasis < 0 || input.SalvageValue < 0 {
		return errors.New("nilai basis dan nilai residu tidak boleh negatif")
	}
	if input.SalvageValue > input.DepreciableBasis {
		return errors.New("nilai residu tidak boleh lebih besar dari depreciable basis")
	}
	if _, err := time.Parse("2006-01-02", input.StartDate); err != nil {
		return errors.New("tanggal mulai depresiasi tidak valid")
	}
	if input.ID <= 0 {
		input.Status = "ACTIVE"
	}
	return s.Repo.SaveDepreciationProfile(input)
}

func (s *AssetDepreciationService) PauseDepreciationProfile(profileID int64, reason string, auditCtx models.AuditContext) error {
	reason = strings.TrimSpace(reason)
	if profileID <= 0 {
		return errors.New("profil depresiasi tidak valid")
	}
	if reason == "" {
		return errors.New("alasan jeda depresiasi wajib diisi")
	}
	if len(reason) > 1000 {
		return errors.New("alasan jeda maksimal 1000 karakter")
	}
	if auditCtx.ActorUserID <= 0 {
		return errors.New("pengguna tidak valid")
	}
	return s.Repo.PauseDepreciationProfile(profileID, reason, auditCtx)
}

func (s *AssetDepreciationService) ResumeDepreciationProfile(profileID int64, auditCtx models.AuditContext) error {
	if profileID <= 0 {
		return errors.New("profil depresiasi tidak valid")
	}
	if auditCtx.ActorUserID <= 0 {
		return errors.New("pengguna tidak valid")
	}
	return s.Repo.ResumeDepreciationProfile(profileID, auditCtx)
}

func (s *AssetDepreciationService) GetPostingHistory(filter models.DepreciationPostingHistoryFilter) (models.DepreciationPostingHistoryResult, error) {
	if filter.Year < 2000 || filter.Year > 2200 {
		return models.DepreciationPostingHistoryResult{}, errors.New("tahun posting tidak valid")
	}
	if filter.Month < 0 || filter.Month > 12 {
		return models.DepreciationPostingHistoryResult{}, errors.New("bulan posting tidak valid")
	}
	filter.Search = strings.TrimSpace(filter.Search)
	if filter.Page < 1 {
		filter.Page = 1
	}
	if filter.PerPage <= 0 || filter.PerPage > 100 {
		filter.PerPage = 50
	}
	return s.Repo.GetPostingHistory(filter)
}

func (s *AssetDepreciationService) GetMonthlyDepreciation(filter models.MonthlyDepreciationFilter) (models.MonthlyDepreciationResult, error) {
	if filter.Year < 2000 || filter.Year > 2200 {
		return models.MonthlyDepreciationResult{}, errors.New("tahun depresiasi tidak valid")
	}
	if filter.Month < 1 || filter.Month > 12 {
		return models.MonthlyDepreciationResult{}, errors.New("bulan depresiasi tidak valid")
	}
	filter.Status = strings.ToUpper(strings.TrimSpace(filter.Status))
	if filter.Status == "" {
		filter.Status = "ALL"
	}
	if filter.Status != "ALL" && filter.Status != "DRAFT" && filter.Status != "POSTED" && filter.Status != "SKIPPED" && filter.Status != "REVERSED" {
		return models.MonthlyDepreciationResult{}, errors.New("status depresiasi tidak valid")
	}
	filter.Search = strings.TrimSpace(filter.Search)
	if filter.Page < 1 {
		filter.Page = 1
	}
	if filter.PerPage <= 0 || filter.PerPage > 100 {
		filter.PerPage = 50
	}
	return s.Repo.GetMonthlyDepreciation(filter)
}

func (s *AssetDepreciationService) GetDepreciationPeriod(year, month int) (models.DepreciationPeriod, error) {
	if err := validateDepreciationPeriod(year, month, false); err != nil {
		return models.DepreciationPeriod{}, err
	}
	return s.Repo.GetDepreciationPeriod(year, month)
}

func (s *AssetDepreciationService) CloseDepreciationPeriod(year, month int, notes string, auditCtx models.AuditContext) error {
	if err := validateDepreciationPeriod(year, month, true); err != nil {
		return err
	}
	notes = strings.TrimSpace(notes)
	if len(notes) > 1000 {
		return errors.New("catatan penutupan maksimal 1000 karakter")
	}
	if auditCtx.ActorUserID <= 0 {
		return errors.New("pengguna tidak valid")
	}
	return s.Repo.CloseDepreciationPeriod(year, month, notes, auditCtx)
}

func (s *AssetDepreciationService) ReopenDepreciationPeriod(year, month int, reason string, auditCtx models.AuditContext) error {
	if err := validateDepreciationPeriod(year, month, true); err != nil {
		return err
	}
	reason = strings.TrimSpace(reason)
	if reason == "" {
		return errors.New("alasan membuka kembali periode wajib diisi")
	}
	if len(reason) > 1000 {
		return errors.New("alasan membuka kembali periode maksimal 1000 karakter")
	}
	if auditCtx.ActorUserID <= 0 {
		return errors.New("pengguna tidak valid")
	}
	return s.Repo.ReopenDepreciationPeriod(year, month, reason, auditCtx)
}

func (s *AssetDepreciationService) GenerateMonthlySchedules(year, month int, auditCtx models.AuditContext) (int, error) {
	if year < 2000 || year > 2200 {
		return 0, errors.New("tahun depresiasi tidak valid")
	}
	if month < 1 || month > 12 {
		return 0, errors.New("bulan depresiasi tidak valid")
	}
	if auditCtx.ActorUserID <= 0 {
		return 0, errors.New("pengguna tidak valid")
	}
	return s.Repo.GenerateMonthlySchedules(year, month, auditCtx)
}

func (s *AssetDepreciationService) PostSchedules(ids []int64, auditCtx models.AuditContext) (int64, error) {
	if auditCtx.ActorUserID <= 0 {
		return 0, errors.New("pengguna tidak valid")
	}
	return s.Repo.PostSchedules(uniqueDepreciationIDs(ids), auditCtx)
}

func (s *AssetDepreciationService) SkipSchedules(ids []int64, reason string, auditCtx models.AuditContext) (int64, error) {
	reason = strings.TrimSpace(reason)
	if reason == "" {
		return 0, errors.New("alasan melewati depresiasi wajib diisi")
	}
	if len(reason) > 1000 {
		return 0, errors.New("alasan melewati depresiasi maksimal 1000 karakter")
	}
	if auditCtx.ActorUserID <= 0 {
		return 0, errors.New("pengguna tidak valid")
	}
	return s.Repo.SkipSchedules(uniqueDepreciationIDs(ids), reason, auditCtx)
}

func (s *AssetDepreciationService) ReverseSchedule(scheduleID int64, reason string, auditCtx models.AuditContext) (int64, error) {
	reason = strings.TrimSpace(reason)
	if scheduleID <= 0 {
		return 0, errors.New("jadwal depresiasi tidak valid")
	}
	if reason == "" {
		return 0, errors.New("alasan pembatalan posting wajib diisi")
	}
	if len(reason) > 1000 {
		return 0, errors.New("alasan pembatalan posting maksimal 1000 karakter")
	}
	if auditCtx.ActorUserID <= 0 {
		return 0, errors.New("pengguna tidak valid")
	}
	return s.Repo.ReverseSchedule(scheduleID, reason, auditCtx)
}

func (s *AssetDepreciationService) UpdateCorrectionDraft(input models.DepreciationCorrectionInput) error {
	input.Reason = strings.TrimSpace(input.Reason)
	if input.ScheduleID <= 0 {
		return errors.New("draft koreksi tidak valid")
	}
	if input.DepreciationValue <= 0 {
		return errors.New("nilai depresiasi koreksi wajib lebih dari nol")
	}
	if input.Reason == "" {
		return errors.New("alasan koreksi wajib diisi")
	}
	if len(input.Reason) > 1000 {
		return errors.New("alasan koreksi maksimal 1000 karakter")
	}
	if input.AuditContext.ActorUserID <= 0 {
		return errors.New("pengguna tidak valid")
	}
	return s.Repo.UpdateCorrectionDraft(input)
}

func uniqueDepreciationIDs(ids []int64) []int64 {
	unique := make([]int64, 0, len(ids))
	seen := make(map[int64]bool, len(ids))
	for _, id := range ids {
		if id <= 0 || seen[id] {
			continue
		}
		seen[id] = true
		unique = append(unique, id)
	}
	return unique
}

func validateDepreciationPeriod(year, month int, rejectFuture bool) error {
	if year < 2000 || year > 2200 {
		return errors.New("tahun depresiasi tidak valid")
	}
	if month < 1 || month > 12 {
		return errors.New("bulan depresiasi tidak valid")
	}
	if rejectFuture {
		selected := time.Date(year, time.Month(month), 1, 0, 0, 0, 0, time.Local)
		current := time.Date(time.Now().Year(), time.Now().Month(), 1, 0, 0, 0, 0, time.Local)
		if selected.After(current) {
			return errors.New("periode depresiasi masa depan belum dapat ditutup atau dibuka kembali")
		}
	}
	return nil
}
