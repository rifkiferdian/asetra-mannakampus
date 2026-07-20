package services

import (
	"errors"
	"gobase-app/models"
	"gobase-app/repositories"
	"sort"
	"strconv"
	"strings"
	"time"
)

type AssetDisposalApprovalService struct {
	Repo *repositories.AssetDisposalApprovalRepository
}

func (s *AssetDisposalApprovalService) GetRules() ([]models.AssetDisposalApprovalRule, error) {
	return s.Repo.GetRules()
}

func (s *AssetDisposalApprovalService) SaveRule(input models.AssetDisposalApprovalRuleInput) error {
	input.Name = strings.TrimSpace(input.Name)
	if input.Name == "" {
		return errors.New("nama aturan wajib diisi")
	}
	if len(input.Name) > 150 {
		return errors.New("nama aturan maksimal 150 karakter")
	}
	if input.MinBookValue < 0 {
		return errors.New("nilai buku minimum tidak boleh negatif")
	}
	if input.MaxBookValue != nil && *input.MaxBookValue < input.MinBookValue {
		return errors.New("nilai buku maksimum harus sama atau lebih besar dari minimum")
	}
	if input.Priority < 1 {
		return errors.New("prioritas minimal 1")
	}
	if len(input.Steps) == 0 {
		return errors.New("minimal satu tahap approval wajib diisi")
	}
	if err := validateApprovalEffectiveDates(input.EffectiveFrom, input.EffectiveUntil); err != nil {
		return err
	}
	seen := map[string]bool{}
	for i := range input.Steps {
		step := &input.Steps[i]
		step.Scope = strings.ToUpper(strings.TrimSpace(step.Scope))
		if step.StepOrder < 1 || step.RoleID <= 0 {
			return errors.New("urutan dan role pada setiap tahap wajib diisi")
		}
		if step.Scope != "STORE" && step.Scope != "HO" && step.Scope != "ANY" {
			return errors.New("scope tahap approval tidak valid")
		}
		key := strconv.Itoa(step.StepOrder) + ":" + strconv.FormatInt(step.RoleID, 10)
		if seen[key] {
			return errors.New("role yang sama tidak boleh berulang pada urutan tahap yang sama")
		}
		seen[key] = true
	}
	sort.SliceStable(input.Steps, func(i, j int) bool { return input.Steps[i].StepOrder < input.Steps[j].StepOrder })
	return s.Repo.SaveRule(input)
}

func (s *AssetDisposalApprovalService) DeleteRule(id int64) error {
	if id <= 0 {
		return errors.New("aturan approval tidak valid")
	}
	return s.Repo.DeleteRule(id)
}

func (s *AssetDisposalApprovalService) GetApprovers() ([]models.AssetDisposalApprover, error) {
	return s.Repo.GetApprovers()
}

func (s *AssetDisposalApprovalService) SaveApprover(input models.AssetDisposalApproverInput) error {
	input.Scope = strings.ToUpper(strings.TrimSpace(input.Scope))
	if input.Scope != "STORE" && input.Scope != "HO" {
		return errors.New("scope approver tidak valid")
	}
	if input.Scope == "STORE" && input.StoreID <= 0 {
		return errors.New("store wajib dipilih untuk scope STORE")
	}
	if input.Scope == "HO" {
		input.StoreID = 0
	}
	if input.RoleID <= 0 || input.UserID <= 0 {
		return errors.New("role dan user approver wajib dipilih")
	}
	return s.Repo.SaveApprover(input)
}

func (s *AssetDisposalApprovalService) DeleteApprover(id int64) error {
	if id <= 0 {
		return errors.New("pemetaan approver tidak valid")
	}
	return s.Repo.DeleteApprover(id)
}

func (s *AssetDisposalApprovalService) Submit(id int64, auditCtx models.AuditContext) error {
	if id <= 0 || auditCtx.ActorUserID <= 0 {
		return errors.New("transaksi atau pengguna tidak valid")
	}
	return s.Repo.Submit(id, auditCtx)
}

func (s *AssetDisposalApprovalService) GetInbox(userID int, filter models.AssetDisposalApprovalInboxFilter) (models.AssetDisposalApprovalInboxResult, error) {
	if userID <= 0 {
		return models.AssetDisposalApprovalInboxResult{}, errors.New("pengguna tidak valid")
	}
	filter.Status = strings.ToUpper(strings.TrimSpace(filter.Status))
	if filter.Status == "" {
		filter.Status = "ALL"
	}
	valid := map[string]bool{"ALL": true, "WAITING": true, "PENDING": true, "APPROVED": true, "REJECTED": true, "CANCELLED": true, "SKIPPED": true}
	if !valid[filter.Status] {
		return models.AssetDisposalApprovalInboxResult{}, errors.New("status tugas tidak valid")
	}
	filter.Search = strings.TrimSpace(filter.Search)
	if filter.Page < 1 {
		filter.Page = 1
	}
	if filter.PerPage <= 0 || filter.PerPage > 100 {
		filter.PerPage = 50
	}
	return s.Repo.GetInbox(userID, filter)
}

func (s *AssetDisposalApprovalService) ApproveTask(id int64, comment string, auditCtx models.AuditContext) error {
	comment = strings.TrimSpace(comment)
	if id <= 0 || auditCtx.ActorUserID <= 0 {
		return errors.New("tugas atau pengguna tidak valid")
	}
	if len(comment) > 2000 {
		return errors.New("catatan maksimal 2000 karakter")
	}
	return s.Repo.ApproveTask(id, comment, auditCtx)
}

func (s *AssetDisposalApprovalService) RejectTask(id int64, reason string, auditCtx models.AuditContext) error {
	reason = strings.TrimSpace(reason)
	if id <= 0 || auditCtx.ActorUserID <= 0 {
		return errors.New("tugas atau pengguna tidak valid")
	}
	if reason == "" {
		return errors.New("alasan penolakan wajib diisi")
	}
	if len(reason) > 2000 {
		return errors.New("alasan maksimal 2000 karakter")
	}
	return s.Repo.RejectTask(id, reason, auditCtx)
}

func validateApprovalEffectiveDates(from, until string) error {
	var start, end time.Time
	var err error
	if strings.TrimSpace(from) != "" {
		start, err = time.Parse("2006-01-02", from)
		if err != nil {
			return errors.New("tanggal mulai berlaku tidak valid")
		}
	}
	if strings.TrimSpace(until) != "" {
		end, err = time.Parse("2006-01-02", until)
		if err != nil {
			return errors.New("tanggal akhir berlaku tidak valid")
		}
	}
	if !start.IsZero() && !end.IsZero() && end.Before(start) {
		return errors.New("tanggal akhir berlaku tidak boleh sebelum tanggal mulai")
	}
	return nil
}
