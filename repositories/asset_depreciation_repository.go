package repositories

import (
	"database/sql"
	"errors"
	"fmt"
	"gobase-app/models"
	"math"
	"strconv"
	"strings"
	"time"
)

type AssetDepreciationRepository struct {
	DB *sql.DB
}

type depreciationScheduleActionRecord struct {
	ID         int64
	ProfileID  int64
	PeriodDate time.Time
}

type depreciationGenerationCandidate struct {
	ProfileID       int64
	AssetID         int64
	UsefulLife      int
	Basis           float64
	Salvage         float64
	StartDate       time.Time
	FirstPolicyCode string
	LastPolicyCode  string
	DisposalDate    sql.NullTime
	PostedAmount    float64
	PriorDrafts     int
	LaterSchedules  int
}

func (r *AssetDepreciationRepository) GetAssetDepreciationDetail(assetID int64) (models.AssetDepreciationDetail, error) {
	var detail models.AssetDepreciationDetail
	var acquisitionValue float64
	var profileID sql.NullInt64
	var methodCode, methodName, profileStatus sql.NullString
	var usefulLife sql.NullInt64
	var salvageValue, depreciableBasis sql.NullFloat64
	var startDate sql.NullTime

	err := r.DB.QueryRow(`
		SELECT
			a.acquisition_value,
			adp.id,
			adm.code,
			adm.name,
			adp.useful_life_months,
			adp.salvage_value,
			adp.depreciable_basis,
			adp.start_date,
			adp.status
		FROM assets a
		LEFT JOIN asset_depreciation_profiles adp ON adp.asset_id = a.id
		LEFT JOIN asset_depreciation_methods adm ON adm.id = adp.depreciation_method_id
		WHERE a.id = ?
		LIMIT 1
	`, assetID).Scan(
		&acquisitionValue,
		&profileID,
		&methodCode,
		&methodName,
		&usefulLife,
		&salvageValue,
		&depreciableBasis,
		&startDate,
		&profileStatus,
	)
	if err != nil {
		return detail, err
	}

	detail.CurrentBookValueDisplay = formatAssetAmountID(acquisitionValue)
	if !profileID.Valid {
		return detail, nil
	}

	detail.Configured = true
	detail.ProfileID = profileID.Int64
	detail.MethodCode = methodCode.String
	detail.MethodName = methodName.String
	detail.UsefulLifeMonths = int(usefulLife.Int64)
	detail.ProfileStatus = profileStatus.String
	detail.SalvageValueDisplay = formatAssetAmountID(salvageValue.Float64)
	detail.DepreciableBasisDisplay = formatAssetAmountID(depreciableBasis.Float64)
	if startDate.Valid {
		detail.StartDateDisplay = formatDepreciationDateID(startDate.Time, false)
	}

	var postedDepreciation, monthlyDepreciation float64
	var lastPostedPeriod, nextDraftPeriod sql.NullTime
	err = r.DB.QueryRow(`
		SELECT
			COALESCE(SUM(CASE WHEN status = 'POSTED' THEN depreciation_amount ELSE 0 END), 0),
			COALESCE(SUM(CASE WHEN status = 'POSTED' THEN 1 ELSE 0 END), 0),
			COALESCE(SUM(CASE WHEN status = 'DRAFT' THEN 1 ELSE 0 END), 0),
			MAX(CASE WHEN status = 'POSTED' THEN period_date END),
			MIN(CASE WHEN status = 'DRAFT' THEN period_date END),
			COALESCE((
				SELECT schedule.depreciation_amount
				FROM asset_depreciation_schedules schedule
				WHERE schedule.profile_id = ?
				ORDER BY CASE WHEN schedule.status = 'DRAFT' THEN 0 ELSE 1 END, schedule.period_date ASC
				LIMIT 1
			), 0)
		FROM asset_depreciation_schedules
		WHERE profile_id = ?
	`, detail.ProfileID, detail.ProfileID).Scan(
		&postedDepreciation,
		&detail.PostedScheduleCount,
		&detail.DraftScheduleCount,
		&lastPostedPeriod,
		&nextDraftPeriod,
		&monthlyDepreciation,
	)
	if err != nil {
		return detail, err
	}

	currentBookValue := depreciableBasis.Float64 - postedDepreciation
	if currentBookValue < salvageValue.Float64 {
		currentBookValue = salvageValue.Float64
	}
	detail.MonthlyDepreciationDisplay = formatAssetAmountID(monthlyDepreciation)
	detail.PostedDepreciationDisplay = formatAssetAmountID(postedDepreciation)
	detail.CurrentBookValueDisplay = formatAssetAmountID(currentBookValue)

	depreciableAmount := depreciableBasis.Float64 - salvageValue.Float64
	if depreciableAmount > 0 {
		detail.ProgressPercent = math.Min(100, math.Max(0, postedDepreciation/depreciableAmount*100))
	}
	detail.ProgressPercentDisplay = strconv.FormatFloat(detail.ProgressPercent, 'f', 1, 64) + "%"
	if lastPostedPeriod.Valid {
		detail.LastPostedPeriodDisplay = formatDepreciationMonthYearID(lastPostedPeriod.Time)
	}
	if nextDraftPeriod.Valid {
		detail.NextDraftPeriodDisplay = formatDepreciationMonthYearID(nextDraftPeriod.Time)
	}
	return detail, nil
}

func (r *AssetDepreciationRepository) GetPostedDepreciationByAssetID(assetID int64, limit int) ([]models.AssetDepreciationPosting, error) {
	if limit <= 0 {
		limit = 12
	}
	rows, err := r.DB.Query(`
		SELECT
			id,
			period_date,
			opening_book_value,
			depreciation_amount,
			accumulated_depreciation,
			closing_book_value,
			posted_at
		FROM asset_depreciation_schedules
		WHERE asset_id = ? AND status = 'POSTED'
		ORDER BY period_date DESC, id DESC
		LIMIT ?
	`, assetID, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	postings := make([]models.AssetDepreciationPosting, 0)
	for rows.Next() {
		var posting models.AssetDepreciationPosting
		var periodDate time.Time
		var postedAt sql.NullTime
		var openingValue, depreciationAmount, accumulatedValue, closingValue float64
		if err := rows.Scan(
			&posting.ID,
			&periodDate,
			&openingValue,
			&depreciationAmount,
			&accumulatedValue,
			&closingValue,
			&postedAt,
		); err != nil {
			return nil, err
		}
		posting.PeriodDisplay = formatDepreciationMonthYearID(periodDate)
		posting.OpeningBookValueDisplay = formatAssetAmountID(openingValue)
		posting.DepreciationAmountDisplay = formatAssetAmountID(depreciationAmount)
		posting.AccumulatedDepreciationDisplay = formatAssetAmountID(accumulatedValue)
		posting.ClosingBookValueDisplay = formatAssetAmountID(closingValue)
		if postedAt.Valid {
			posting.PostedAtDisplay = formatDepreciationDateID(postedAt.Time, true)
		}
		postings = append(postings, posting)
	}
	return postings, rows.Err()
}

func (r *AssetDepreciationRepository) GetDepreciationProfiles(filter models.DepreciationProfileFilter) (models.DepreciationProfileResult, error) {
	result := models.DepreciationProfileResult{}
	if err := r.loadDepreciationProfileStats(&result.Stats); err != nil {
		return result, err
	}

	where, args := depreciationProfileWhere(filter)
	if err := r.DB.QueryRow(`
		SELECT COUNT(*)
		FROM asset_depreciation_profiles adp
		JOIN assets a ON a.id = adp.asset_id
		LEFT JOIN asset_types at ON at.id = a.asset_type_id
		JOIN asset_depreciation_methods adm ON adm.id = adp.depreciation_method_id
		WHERE `+where, args...).Scan(&result.TotalRows); err != nil {
		return result, err
	}
	result.TotalPages = 1
	if result.TotalRows > 0 {
		result.TotalPages = (result.TotalRows + filter.PerPage - 1) / filter.PerPage
	}
	if filter.Page > result.TotalPages {
		filter.Page = result.TotalPages
	}
	offset := (filter.Page - 1) * filter.PerPage

	query := `
		SELECT
			adp.id,
			adp.asset_id,
			a.asset_code,
			a.asset_name,
			COALESCE(at.name, ''),
			adm.id,
			adm.code,
			adm.name,
			COALESCE(adp.useful_life_months, 0),
			adp.salvage_value,
			adp.depreciable_basis,
			adp.start_date,
			adp.first_month_policy_id,
			first_policy.code,
			first_policy.name,
			adp.last_month_policy_id,
			last_policy.code,
			last_policy.name,
			adp.status,
			COALESCE(adp.notes, ''),
			adp.paused_at,
			COALESCE(paused_user.name, ''),
			COALESCE(adp.pause_reason, ''),
			adp.resumed_at,
			COALESCE(resumed_user.name, ''),
			adp.finished_at,
			adp.terminated_at,
			COALESCE(terminated_user.name, ''),
			COALESCE(adp.termination_reason, ''),
			COALESCE(schedule.posted_amount, 0),
			COALESCE(schedule.posted_count, 0),
			COALESCE(schedule.draft_count, 0),
			schedule.last_posted_period
		FROM asset_depreciation_profiles adp
		JOIN assets a ON a.id = adp.asset_id
		LEFT JOIN asset_types at ON at.id = a.asset_type_id
		JOIN asset_depreciation_methods adm ON adm.id = adp.depreciation_method_id
		JOIN asset_depreciation_first_month_policies first_policy ON first_policy.id = adp.first_month_policy_id
		JOIN asset_depreciation_last_month_policies last_policy ON last_policy.id = adp.last_month_policy_id
		LEFT JOIN users paused_user ON paused_user.id = adp.paused_by
		LEFT JOIN users resumed_user ON resumed_user.id = adp.resumed_by
		LEFT JOIN users terminated_user ON terminated_user.id = adp.terminated_by
		LEFT JOIN (
			SELECT
				profile_id,
				SUM(CASE WHEN status = 'POSTED' THEN depreciation_amount ELSE 0 END) AS posted_amount,
				SUM(CASE WHEN status = 'POSTED' THEN 1 ELSE 0 END) AS posted_count,
				SUM(CASE WHEN status = 'DRAFT' THEN 1 ELSE 0 END) AS draft_count,
				MAX(CASE WHEN status = 'POSTED' THEN period_date END) AS last_posted_period
			FROM asset_depreciation_schedules
			GROUP BY profile_id
		) schedule ON schedule.profile_id = adp.id
		WHERE ` + where + `
		ORDER BY a.asset_code ASC
		LIMIT ? OFFSET ?`
	queryArgs := append(append([]any{}, args...), filter.PerPage, offset)
	rows, err := r.DB.Query(query, queryArgs...)
	if err != nil {
		return result, err
	}
	defer rows.Close()

	for rows.Next() {
		var item models.AssetDepreciationProfile
		var salvageValue, depreciableBasis, postedAmount float64
		var startDate time.Time
		var pausedAt, resumedAt, finishedAt, terminatedAt, lastPostedPeriod sql.NullTime
		if err := rows.Scan(
			&item.ID, &item.AssetID, &item.AssetCode, &item.AssetName, &item.AssetTypeName,
			&item.MethodID, &item.MethodCode, &item.MethodName, &item.UsefulLifeMonths,
			&salvageValue, &depreciableBasis, &startDate,
			&item.FirstMonthPolicyID, &item.FirstMonthPolicyCode, &item.FirstMonthPolicyName,
			&item.LastMonthPolicyID, &item.LastMonthPolicyCode, &item.LastMonthPolicyName,
			&item.Status, &item.Notes,
			&pausedAt, &item.PausedByName, &item.PauseReason,
			&resumedAt, &item.ResumedByName, &finishedAt,
			&terminatedAt, &item.TerminatedByName, &item.TerminationReason,
			&postedAmount, &item.PostedScheduleCount, &item.DraftScheduleCount, &lastPostedPeriod,
		); err != nil {
			return result, err
		}
		monthlyAmount := 0.0
		if item.MethodCode == "STRAIGHT_LINE" && item.UsefulLifeMonths > 0 && depreciableBasis > salvageValue {
			monthlyAmount = math.Round((depreciableBasis-salvageValue)/float64(item.UsefulLifeMonths)*100) / 100
		}
		currentBookValue := math.Max(salvageValue, depreciableBasis-postedAmount)
		item.SalvageValueInput = formatNumberInput(salvageValue)
		item.SalvageValueDisplay = formatAssetAmountID(salvageValue)
		item.DepreciableBasisInput = formatNumberInput(depreciableBasis)
		item.DepreciableBasisDisplay = formatAssetAmountID(depreciableBasis)
		item.MonthlyDepreciationDisplay = formatAssetAmountID(monthlyAmount)
		item.PostedDepreciationDisplay = formatAssetAmountID(postedAmount)
		item.CurrentBookValueDisplay = formatAssetAmountID(currentBookValue)
		item.StartDate = startDate.Format("2006-01-02")
		item.ConfigurationLocked = item.PostedScheduleCount > 0
		if pausedAt.Valid {
			item.PausedAtDisplay = formatDepreciationDateID(pausedAt.Time, true)
		}
		if resumedAt.Valid {
			item.ResumedAtDisplay = formatDepreciationDateID(resumedAt.Time, true)
		}
		if finishedAt.Valid {
			item.FinishedAtDisplay = formatDepreciationDateID(finishedAt.Time, true)
		}
		if terminatedAt.Valid {
			item.TerminatedAtDisplay = formatDepreciationDateID(terminatedAt.Time, true)
		}
		if lastPostedPeriod.Valid {
			item.LastPostedPeriodDisplay = formatDepreciationMonthYearID(lastPostedPeriod.Time)
		}
		result.Items = append(result.Items, item)
	}
	return result, rows.Err()
}

func (r *AssetDepreciationRepository) loadDepreciationProfileStats(stats *models.DepreciationProfileStats) error {
	return r.DB.QueryRow(`
		SELECT
			COUNT(*),
			COALESCE(SUM(status = 'ACTIVE'), 0),
			COALESCE(SUM(status = 'PAUSED'), 0),
			COALESCE(SUM(status = 'FINISHED'), 0),
			COALESCE(SUM(status = 'TERMINATED'), 0),
			(SELECT COUNT(*) FROM assets a WHERE a.status <> 'DISPOSED' AND NOT EXISTS (
				SELECT 1 FROM asset_depreciation_profiles profile WHERE profile.asset_id = a.id
			))
		FROM asset_depreciation_profiles
	`).Scan(&stats.TotalProfiles, &stats.ActiveProfiles, &stats.PausedProfiles, &stats.FinishedProfiles, &stats.TerminatedProfiles, &stats.UnconfiguredAssets)
}

func (r *AssetDepreciationRepository) GetDepreciationMethods() ([]models.DepreciationMethodOption, error) {
	rows, err := r.DB.Query(`
		SELECT id, code, name
		FROM asset_depreciation_methods
		WHERE is_active = 1
		ORDER BY CASE WHEN code = 'STRAIGHT_LINE' THEN 0 ELSE 1 END, name ASC
	`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	methods := make([]models.DepreciationMethodOption, 0)
	for rows.Next() {
		var method models.DepreciationMethodOption
		if err := rows.Scan(&method.ID, &method.Code, &method.Name); err != nil {
			return nil, err
		}
		methods = append(methods, method)
	}
	return methods, rows.Err()
}

func (r *AssetDepreciationRepository) GetFirstMonthPolicies() ([]models.DepreciationPolicyOption, error) {
	return r.getDepreciationPolicies("asset_depreciation_first_month_policies")
}

func (r *AssetDepreciationRepository) GetLastMonthPolicies() ([]models.DepreciationPolicyOption, error) {
	return r.getDepreciationPolicies("asset_depreciation_last_month_policies")
}

func (r *AssetDepreciationRepository) getDepreciationPolicies(table string) ([]models.DepreciationPolicyOption, error) {
	if table != "asset_depreciation_first_month_policies" && table != "asset_depreciation_last_month_policies" {
		return nil, errors.New("master kebijakan depresiasi tidak valid")
	}
	rows, err := r.DB.Query(`SELECT id, code, name, COALESCE(description, '') FROM ` + table + ` WHERE is_active = 1 ORDER BY id`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	items := make([]models.DepreciationPolicyOption, 0)
	for rows.Next() {
		var item models.DepreciationPolicyOption
		if err := rows.Scan(&item.ID, &item.Code, &item.Name, &item.Description); err != nil {
			return nil, err
		}
		items = append(items, item)
	}
	return items, rows.Err()
}

func (r *AssetDepreciationRepository) GetDepreciationAssetOptions() ([]models.DepreciationAssetOption, error) {
	rows, err := r.DB.Query(`
		SELECT
			a.id,
			a.asset_code,
			a.asset_name,
			COALESCE(at.name, ''),
			a.acquisition_date,
			a.acquisition_value,
			CASE WHEN adp.id IS NULL THEN 0 ELSE 1 END
		FROM assets a
		LEFT JOIN asset_types at ON at.id = a.asset_type_id
		LEFT JOIN asset_depreciation_profiles adp ON adp.asset_id = a.id
		WHERE a.status <> 'DISPOSED'
		ORDER BY a.asset_code ASC
	`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	assets := make([]models.DepreciationAssetOption, 0)
	for rows.Next() {
		var asset models.DepreciationAssetOption
		var acquisitionDate sql.NullTime
		var acquisitionValue float64
		var hasProfile int
		if err := rows.Scan(&asset.ID, &asset.AssetCode, &asset.AssetName, &asset.AssetTypeName, &acquisitionDate, &acquisitionValue, &hasProfile); err != nil {
			return nil, err
		}
		if acquisitionDate.Valid {
			asset.AcquisitionDate = acquisitionDate.Time.Format("2006-01-02")
		}
		asset.AcquisitionValueInput = formatNumberInput(acquisitionValue)
		asset.HasProfile = hasProfile == 1
		assets = append(assets, asset)
	}
	return assets, rows.Err()
}

func (r *AssetDepreciationRepository) SaveDepreciationProfile(input models.DepreciationProfileInput) error {
	tx, err := r.DB.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	var methodCode string
	if err := tx.QueryRow(`SELECT code FROM asset_depreciation_methods WHERE id = ? AND is_active = 1`, input.MethodID).Scan(&methodCode); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return errors.New("metode depresiasi tidak ditemukan atau tidak aktif")
		}
		return err
	}
	if methodCode == "NONE" {
		input.UsefulLifeMonths = 0
		input.Status = "FINISHED"
	} else if input.UsefulLifeMonths <= 0 {
		return errors.New("umur manfaat wajib lebih dari 0 bulan")
	} else if input.DepreciableBasis <= input.SalvageValue {
		return errors.New("depreciable basis harus lebih besar dari nilai residu")
	}
	var policyCount int
	if err := tx.QueryRow(`
		SELECT
			(SELECT COUNT(*) FROM asset_depreciation_first_month_policies WHERE id = ? AND is_active = 1) +
			(SELECT COUNT(*) FROM asset_depreciation_last_month_policies WHERE id = ? AND is_active = 1)
	`, input.FirstMonthPolicyID, input.LastMonthPolicyID).Scan(&policyCount); err != nil {
		return err
	}
	if policyCount != 2 {
		return errors.New("kebijakan bulan pertama atau bulan terakhir tidak valid")
	}

	var acquisitionDate sql.NullTime
	var assetStatus string
	if err := tx.QueryRow(`SELECT acquisition_date, status FROM assets WHERE id = ?`, input.AssetID).Scan(&acquisitionDate, &assetStatus); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return errors.New("asset tidak ditemukan")
		}
		return err
	}
	if assetStatus == "DISPOSED" {
		return errors.New("asset disposed tidak dapat memiliki profile depresiasi aktif")
	}
	if acquisitionDate.Valid && input.StartDate < acquisitionDate.Time.Format("2006-01-02") {
		return errors.New("tanggal mulai depresiasi tidak boleh sebelum tanggal perolehan asset")
	}

	if input.ID <= 0 {
		var existingID int64
		err := tx.QueryRow(`SELECT id FROM asset_depreciation_profiles WHERE asset_id = ?`, input.AssetID).Scan(&existingID)
		if err == nil {
			return errors.New("asset sudah memiliki depreciation profile")
		}
		if !errors.Is(err, sql.ErrNoRows) {
			return err
		}
		result, err := tx.Exec(`
			INSERT INTO asset_depreciation_profiles (
				asset_id, depreciation_method_id, useful_life_months, salvage_value,
				depreciable_basis, start_date, first_month_policy_id, last_month_policy_id,
				status, finished_at, notes
			) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, IF(? = 'FINISHED', CURRENT_TIMESTAMP, NULL), ?)
		`, input.AssetID, input.MethodID, nullableUsefulLife(input.UsefulLifeMonths), input.SalvageValue,
			input.DepreciableBasis, input.StartDate, input.FirstMonthPolicyID, input.LastMonthPolicyID,
			input.Status, input.Status, nullableString(input.Notes))
		if err != nil {
			return err
		}
		profileID, err := result.LastInsertId()
		if err != nil {
			return err
		}
		if err := insertAuditLogTx(tx, "DEPRECIATION_PROFILE", profileID, "CREATE_PROFILE",
			fmt.Sprintf("Profil depresiasi dibuat untuk aset ID %d dengan status %s", input.AssetID, input.Status), input.AuditContext); err != nil {
			return err
		}
		return tx.Commit()
	}

	var existingAssetID, existingMethodID, existingFirstPolicyID, existingLastPolicyID int64
	var existingLife sql.NullInt64
	var existingSalvage, existingBasis float64
	var existingStart time.Time
	var existingStatus string
	if err := tx.QueryRow(`
		SELECT asset_id, depreciation_method_id, useful_life_months, salvage_value, depreciable_basis,
			start_date, first_month_policy_id, last_month_policy_id, status
		FROM asset_depreciation_profiles
		WHERE id = ?
		FOR UPDATE
	`, input.ID).Scan(&existingAssetID, &existingMethodID, &existingLife, &existingSalvage, &existingBasis,
		&existingStart, &existingFirstPolicyID, &existingLastPolicyID, &existingStatus); err != nil {
		return err
	}
	if existingAssetID != input.AssetID {
		return errors.New("aset pada profil depresiasi tidak dapat diubah")
	}

	var postedCount int
	if err := tx.QueryRow(`SELECT COUNT(*) FROM asset_depreciation_schedules WHERE profile_id = ? AND status = 'POSTED'`, input.ID).Scan(&postedCount); err != nil {
		return err
	}
	configurationChanged := existingMethodID != input.MethodID || int(existingLife.Int64) != input.UsefulLifeMonths ||
		math.Abs(existingSalvage-input.SalvageValue) > 0.005 || math.Abs(existingBasis-input.DepreciableBasis) > 0.005 ||
		existingStart.Format("2006-01-02") != input.StartDate || existingFirstPolicyID != input.FirstMonthPolicyID ||
		existingLastPolicyID != input.LastMonthPolicyID
	if postedCount > 0 && configurationChanged {
		return errors.New("konfigurasi utama tidak dapat diubah karena profil sudah memiliki depresiasi yang diposting")
	}

	input.Status = existingStatus
	if methodCode == "NONE" {
		input.Status = "FINISHED"
	} else if postedCount == 0 && existingStatus == "FINISHED" && existingMethodID != input.MethodID {
		input.Status = "ACTIVE"
	}
	if _, err := tx.Exec(`
		UPDATE asset_depreciation_profiles
		SET depreciation_method_id = ?, useful_life_months = ?, salvage_value = ?,
			depreciable_basis = ?, start_date = ?, first_month_policy_id = ?, last_month_policy_id = ?,
			status = ?, finished_at = IF(? = 'FINISHED', COALESCE(finished_at, CURRENT_TIMESTAMP), NULL), notes = ?
		WHERE id = ?
	`, input.MethodID, nullableUsefulLife(input.UsefulLifeMonths), input.SalvageValue,
		input.DepreciableBasis, input.StartDate, input.FirstMonthPolicyID, input.LastMonthPolicyID,
		input.Status, input.Status, nullableString(input.Notes), input.ID); err != nil {
		return err
	}
	if configurationChanged {
		if _, err := tx.Exec(`DELETE FROM asset_depreciation_schedules WHERE profile_id = ? AND status = 'DRAFT'`, input.ID); err != nil {
			return err
		}
	}
	if err := insertAuditLogTx(tx, "DEPRECIATION_PROFILE", input.ID, "UPDATE_PROFILE",
		fmt.Sprintf("Profil depresiasi aset ID %d diperbarui dengan status %s", input.AssetID, input.Status), input.AuditContext); err != nil {
		return err
	}
	return tx.Commit()
}

func (r *AssetDepreciationRepository) PauseDepreciationProfile(profileID int64, reason string, auditCtx models.AuditContext) error {
	tx, err := r.DB.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	var status, assetCode string
	if err := tx.QueryRow(`
		SELECT adp.status, a.asset_code
		FROM asset_depreciation_profiles adp
		JOIN assets a ON a.id = adp.asset_id
		WHERE adp.id = ?
		FOR UPDATE
	`, profileID).Scan(&status, &assetCode); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return errors.New("profil depresiasi tidak ditemukan")
		}
		return err
	}
	if status != "ACTIVE" {
		return fmt.Errorf("hanya profil ACTIVE yang dapat dijeda, status saat ini %s", status)
	}
	var draftCount int
	if err := tx.QueryRow(`SELECT COUNT(*) FROM asset_depreciation_schedules WHERE profile_id = ? AND status = 'DRAFT'`, profileID).Scan(&draftCount); err != nil {
		return err
	}
	if draftCount > 0 {
		return errors.New("profil masih memiliki draft depresiasi; posting atau lewati seluruh draft sebelum menjeda")
	}
	if _, err := tx.Exec(`
		UPDATE asset_depreciation_profiles
		SET status = 'PAUSED', paused_at = CURRENT_TIMESTAMP, paused_by = ?, pause_reason = ?,
			resumed_at = NULL, resumed_by = NULL, updated_at = CURRENT_TIMESTAMP
		WHERE id = ? AND status = 'ACTIVE'
	`, auditCtx.ActorUserID, reason, profileID); err != nil {
		return err
	}
	if err := insertAuditLogTx(tx, "DEPRECIATION_PROFILE", profileID, "PAUSE_PROFILE",
		fmt.Sprintf("Depresiasi aset %s dijeda. Alasan: %s", assetCode, reason), auditCtx); err != nil {
		return err
	}
	return tx.Commit()
}

func (r *AssetDepreciationRepository) ResumeDepreciationProfile(profileID int64, auditCtx models.AuditContext) error {
	tx, err := r.DB.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	var status, assetStatus, assetCode string
	var assetID int64
	if err := tx.QueryRow(`
		SELECT adp.status, adp.asset_id, a.status, a.asset_code
		FROM asset_depreciation_profiles adp
		JOIN assets a ON a.id = adp.asset_id
		WHERE adp.id = ?
		FOR UPDATE
	`, profileID).Scan(&status, &assetID, &assetStatus, &assetCode); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return errors.New("profil depresiasi tidak ditemukan")
		}
		return err
	}
	if status != "PAUSED" {
		return fmt.Errorf("hanya profil PAUSED yang dapat dilanjutkan, status saat ini %s", status)
	}
	if assetStatus == "DISPOSED" {
		return errors.New("profil aset yang sudah disposed tidak dapat dilanjutkan")
	}
	var postedDisposalCount int
	if err := tx.QueryRow(`SELECT COUNT(*) FROM asset_disposals WHERE asset_id = ? AND status = 'POSTED'`, assetID).Scan(&postedDisposalCount); err != nil {
		return err
	}
	if postedDisposalCount > 0 {
		return errors.New("profil tidak dapat dilanjutkan karena aset sudah memiliki disposal yang diposting")
	}
	if _, err := tx.Exec(`
		UPDATE asset_depreciation_profiles
		SET status = 'ACTIVE', resumed_at = CURRENT_TIMESTAMP, resumed_by = ?, updated_at = CURRENT_TIMESTAMP
		WHERE id = ? AND status = 'PAUSED'
	`, auditCtx.ActorUserID, profileID); err != nil {
		return err
	}
	if err := insertAuditLogTx(tx, "DEPRECIATION_PROFILE", profileID, "RESUME_PROFILE",
		fmt.Sprintf("Depresiasi aset %s dilanjutkan kembali", assetCode), auditCtx); err != nil {
		return err
	}
	return tx.Commit()
}

func (r *AssetDepreciationRepository) GetPostingHistory(filter models.DepreciationPostingHistoryFilter) (models.DepreciationPostingHistoryResult, error) {
	result := models.DepreciationPostingHistoryResult{}
	where, args := postingHistoryWhere(filter)
	if err := r.loadPostingHistoryStats(where, args, &result.Stats); err != nil {
		return result, err
	}
	if err := r.DB.QueryRow(`
		SELECT COUNT(*)
		FROM asset_depreciation_schedules ads
		JOIN assets a ON a.id = ads.asset_id
		LEFT JOIN asset_types at ON at.id = a.asset_type_id
		WHERE `+where, args...).Scan(&result.TotalRows); err != nil {
		return result, err
	}
	result.TotalPages = 1
	if result.TotalRows > 0 {
		result.TotalPages = (result.TotalRows + filter.PerPage - 1) / filter.PerPage
	}
	if filter.Page > result.TotalPages {
		filter.Page = result.TotalPages
	}
	offset := (filter.Page - 1) * filter.PerPage
	query := `
		SELECT
			ads.id,
			ads.asset_id,
			a.asset_code,
			a.asset_name,
			COALESCE(at.name, ''),
			COALESCE(adm.code, ''),
			COALESCE(adm.name, ''),
			ads.version_no,
			COALESCE(ads.original_schedule_id, 0),
			ads.status,
			COALESCE(period.status, 'OPEN'),
			ads.period_date,
			ads.opening_book_value,
			ads.depreciation_amount,
			ads.accumulated_depreciation,
			ads.closing_book_value,
			ads.posted_at,
			COALESCE(posted_user.name, ''),
			ads.reversed_at,
			COALESCE(reversed_user.name, ''),
			COALESCE(ads.reversal_reason, '')
		FROM asset_depreciation_schedules ads
		JOIN assets a ON a.id = ads.asset_id
		LEFT JOIN asset_types at ON at.id = a.asset_type_id
		LEFT JOIN asset_depreciation_profiles adp ON adp.id = ads.profile_id
		LEFT JOIN asset_depreciation_methods adm ON adm.id = adp.depreciation_method_id
		LEFT JOIN users posted_user ON posted_user.id = ads.posted_by
		LEFT JOIN users reversed_user ON reversed_user.id = ads.reversed_by
		LEFT JOIN asset_depreciation_periods period ON period.period_year = ads.period_year AND period.period_month = ads.period_month
		WHERE ` + where + `
		ORDER BY ads.period_date DESC, COALESCE(ads.reversed_at, ads.posted_at) DESC, ads.version_no DESC
		LIMIT ? OFFSET ?`
	queryArgs := append(append([]any{}, args...), filter.PerPage, offset)
	rows, err := r.DB.Query(query, queryArgs...)
	if err != nil {
		return result, err
	}
	defer rows.Close()
	for rows.Next() {
		var item models.DepreciationPostingHistoryItem
		var periodDate time.Time
		var postedAt, reversedAt sql.NullTime
		var openingValue, depreciationAmount, accumulatedValue, closingValue float64
		if err := rows.Scan(
			&item.ID, &item.AssetID, &item.AssetCode, &item.AssetName, &item.AssetTypeName, &item.MethodCode, &item.MethodName,
			&item.VersionNo, &item.OriginalScheduleID, &item.Status, &item.PeriodStatus, &periodDate, &openingValue, &depreciationAmount,
			&accumulatedValue, &closingValue, &postedAt, &item.PostedByName, &reversedAt, &item.ReversedByName, &item.ReversalReason,
		); err != nil {
			return result, err
		}
		item.PeriodDisplay = formatDepreciationMonthYearID(periodDate)
		item.PeriodYear = periodDate.Year()
		item.PeriodMonth = int(periodDate.Month())
		item.OpeningBookValueDisplay = formatAssetAmountID(openingValue)
		item.DepreciationAmountDisplay = formatAssetAmountID(depreciationAmount)
		item.AccumulatedDepreciationDisplay = formatAssetAmountID(accumulatedValue)
		item.ClosingBookValueDisplay = formatAssetAmountID(closingValue)
		if postedAt.Valid {
			item.PostedAtDisplay = formatDepreciationDateID(postedAt.Time, true)
		}
		if reversedAt.Valid {
			item.ReversedAtDisplay = formatDepreciationDateID(reversedAt.Time, true)
		}
		result.Items = append(result.Items, item)
	}
	return result, rows.Err()
}

func (r *AssetDepreciationRepository) loadPostingHistoryStats(where string, args []any, stats *models.DepreciationPostingHistoryStats) error {
	var totalAmount float64
	var latestPosting sql.NullTime
	err := r.DB.QueryRow(`
		SELECT COUNT(*), COUNT(DISTINCT ads.asset_id),
			COALESCE(SUM(CASE WHEN ads.status = 'POSTED' THEN ads.depreciation_amount ELSE 0 END), 0),
			MAX(ads.posted_at)
		FROM asset_depreciation_schedules ads
		JOIN assets a ON a.id = ads.asset_id
		LEFT JOIN asset_types at ON at.id = a.asset_type_id
		WHERE `+where, args...).Scan(&stats.TotalPostings, &stats.TotalAssets, &totalAmount, &latestPosting)
	if err != nil {
		return err
	}
	stats.TotalAmountDisplay = formatAssetAmountID(totalAmount)
	if latestPosting.Valid {
		stats.LatestPostingDisplay = formatDepreciationDateID(latestPosting.Time, true)
	}
	return nil
}

func depreciationProfileWhere(filter models.DepreciationProfileFilter) (string, []any) {
	clauses := []string{"1 = 1"}
	args := make([]any, 0)
	if filter.Status != "" && filter.Status != "ALL" {
		clauses = append(clauses, "adp.status = ?")
		args = append(args, filter.Status)
	}
	if filter.Search != "" {
		clauses = append(clauses, "(a.asset_code LIKE ? OR a.asset_name LIKE ? OR at.name LIKE ? OR adm.name LIKE ?)")
		term := "%" + filter.Search + "%"
		args = append(args, term, term, term, term)
	}
	return strings.Join(clauses, " AND "), args
}

func postingHistoryWhere(filter models.DepreciationPostingHistoryFilter) (string, []any) {
	clauses := []string{"ads.status IN ('POSTED', 'REVERSED')", "ads.period_year = ?"}
	args := []any{filter.Year}
	if filter.Month > 0 {
		clauses = append(clauses, "ads.period_month = ?")
		args = append(args, filter.Month)
	}
	if filter.Search != "" {
		clauses = append(clauses, "(a.asset_code LIKE ? OR a.asset_name LIKE ? OR at.name LIKE ?)")
		term := "%" + filter.Search + "%"
		args = append(args, term, term, term)
	}
	return strings.Join(clauses, " AND "), args
}

func nullableUsefulLife(value int) any {
	if value <= 0 {
		return nil
	}
	return value
}

func formatDepreciationMonthYearID(value time.Time) string {
	months := []string{"Jan", "Feb", "Mar", "Apr", "Mei", "Jun", "Jul", "Agu", "Sep", "Okt", "Nov", "Des"}
	return fmt.Sprintf("%s %d", months[int(value.Month())-1], value.Year())
}

func formatDepreciationDateID(value time.Time, withTime bool) string {
	months := []string{"Jan", "Feb", "Mar", "Apr", "Mei", "Jun", "Jul", "Agu", "Sep", "Okt", "Nov", "Des"}
	result := fmt.Sprintf("%02d %s %d", value.Day(), months[int(value.Month())-1], value.Year())
	if withTime {
		result += value.Format(" 15:04")
	}
	return result
}

func (r *AssetDepreciationRepository) GetMonthlyDepreciation(filter models.MonthlyDepreciationFilter) (models.MonthlyDepreciationResult, error) {
	result := models.MonthlyDepreciationResult{}
	if err := r.loadMonthlyStats(filter.Year, filter.Month, &result.Stats); err != nil {
		return result, err
	}

	where, args := monthlyDepreciationWhere(filter)
	countQuery := `
		SELECT COUNT(*)
		FROM asset_depreciation_schedules ads
		JOIN assets a ON a.id = ads.asset_id
		LEFT JOIN asset_types at ON at.id = a.asset_type_id
		WHERE ` + where
	if err := r.DB.QueryRow(countQuery, args...).Scan(&result.TotalRows); err != nil {
		return result, err
	}

	result.TotalPages = 1
	if result.TotalRows > 0 {
		result.TotalPages = (result.TotalRows + filter.PerPage - 1) / filter.PerPage
	}
	if filter.Page > result.TotalPages {
		filter.Page = result.TotalPages
	}
	offset := (filter.Page - 1) * filter.PerPage

	query := `
		SELECT
			ads.id,
			ads.asset_id,
			a.asset_code,
			a.asset_name,
			COALESCE(at.name, ''),
			COALESCE(adm.code, ''),
			COALESCE(adm.name, ''),
			COALESCE(adp.useful_life_months, 0),
			ads.period_date,
			ads.version_no,
			COALESCE(ads.original_schedule_id, 0),
			COALESCE(ads.correction_reason, ''),
			a.acquisition_value,
			ads.opening_book_value,
			ads.depreciation_amount,
			ads.accumulated_depreciation,
			ads.closing_book_value,
			ads.status,
			ads.posted_at,
			ads.skipped_at,
			ads.reversed_at,
			COALESCE(posted_user.name, skipped_user.name, reversed_user.name, ''),
			COALESCE(ads.skip_reason, ''),
			COALESCE(ads.reversal_reason, '')
		FROM asset_depreciation_schedules ads
		JOIN assets a ON a.id = ads.asset_id
		LEFT JOIN asset_types at ON at.id = a.asset_type_id
		LEFT JOIN asset_depreciation_profiles adp ON adp.id = ads.profile_id
		LEFT JOIN asset_depreciation_methods adm ON adm.id = adp.depreciation_method_id
		LEFT JOIN users posted_user ON posted_user.id = ads.posted_by
		LEFT JOIN users skipped_user ON skipped_user.id = ads.skipped_by
		LEFT JOIN users reversed_user ON reversed_user.id = ads.reversed_by
		WHERE ` + where + `
		ORDER BY a.asset_code ASC, ads.id ASC
		LIMIT ? OFFSET ?`
	queryArgs := append(append([]any{}, args...), filter.PerPage, offset)
	rows, err := r.DB.Query(query, queryArgs...)
	if err != nil {
		return result, err
	}
	defer rows.Close()

	for rows.Next() {
		var item models.MonthlyDepreciationItem
		var periodDate time.Time
		var postedAt, skippedAt, reversedAt sql.NullTime
		var acquisitionValue, openingValue, depreciationAmount, accumulatedValue, closingValue float64
		if err := rows.Scan(
			&item.ID,
			&item.AssetID,
			&item.AssetCode,
			&item.AssetName,
			&item.AssetTypeName,
			&item.MethodCode,
			&item.MethodName,
			&item.UsefulLifeMonths,
			&periodDate,
			&item.VersionNo,
			&item.OriginalScheduleID,
			&item.CorrectionReason,
			&acquisitionValue,
			&openingValue,
			&depreciationAmount,
			&accumulatedValue,
			&closingValue,
			&item.Status,
			&postedAt,
			&skippedAt,
			&reversedAt,
			&item.ActionByName,
			&item.SkipReason,
			&item.ReversalReason,
		); err != nil {
			return result, err
		}

		item.PeriodDate = formatDepreciationDateID(periodDate, false)
		item.AcquisitionValueDisplay = formatAssetAmountID(acquisitionValue)
		item.OpeningBookValueDisplay = formatAssetAmountID(openingValue)
		item.DepreciationAmountDisplay = formatAssetAmountID(depreciationAmount)
		item.DepreciationAmountInput = formatNumberInput(depreciationAmount)
		item.IsCorrection = item.OriginalScheduleID > 0
		item.AccumulatedDepreciationDisplay = formatAssetAmountID(accumulatedValue)
		item.ClosingBookValueDisplay = formatAssetAmountID(closingValue)
		if postedAt.Valid {
			item.ActionAtDisplay = formatDepreciationDateID(postedAt.Time, true)
		} else if skippedAt.Valid {
			item.ActionAtDisplay = formatDepreciationDateID(skippedAt.Time, true)
		} else if reversedAt.Valid {
			item.ActionAtDisplay = formatDepreciationDateID(reversedAt.Time, true)
		}
		result.Items = append(result.Items, item)
	}

	return result, rows.Err()
}

func (r *AssetDepreciationRepository) loadMonthlyStats(year, month int, stats *models.MonthlyDepreciationStats) error {
	var totalAmount, draftAmount, postedAmount float64
	err := r.DB.QueryRow(`
		SELECT
			COUNT(DISTINCT asset_id),
			COALESCE(SUM(CASE WHEN status = 'DRAFT' THEN 1 ELSE 0 END), 0),
			COALESCE(SUM(CASE WHEN status = 'POSTED' THEN 1 ELSE 0 END), 0),
			COALESCE(SUM(CASE WHEN status = 'SKIPPED' THEN 1 ELSE 0 END), 0),
			COALESCE(SUM(CASE WHEN status = 'REVERSED' THEN 1 ELSE 0 END), 0),
			COALESCE(SUM(CASE WHEN status <> 'SKIPPED' THEN depreciation_amount ELSE 0 END), 0),
			COALESCE(SUM(CASE WHEN status = 'DRAFT' THEN depreciation_amount ELSE 0 END), 0),
			COALESCE(SUM(CASE WHEN status = 'POSTED' THEN depreciation_amount ELSE 0 END), 0)
		FROM asset_depreciation_schedules
		WHERE period_year = ? AND period_month = ?
	`, year, month).Scan(
		&stats.TotalAssets,
		&stats.DraftCount,
		&stats.PostedCount,
		&stats.SkippedCount,
		&stats.ReversedCount,
		&totalAmount,
		&draftAmount,
		&postedAmount,
	)
	if err != nil {
		return err
	}
	stats.TotalDepreciationDisplay = formatAssetAmountID(totalAmount)
	stats.DraftDepreciationDisplay = formatAssetAmountID(draftAmount)
	stats.PostedDepreciationDisplay = formatAssetAmountID(postedAmount)
	return nil
}

func (r *AssetDepreciationRepository) GetDepreciationPeriod(year, month int) (models.DepreciationPeriod, error) {
	period := models.DepreciationPeriod{Year: year, Month: month, Status: "OPEN"}
	var generatedAt, postedAt, closedAt, reopenedAt sql.NullTime
	var closedByName, closingNotes, reopenedByName, reopenReason sql.NullString
	err := r.DB.QueryRow(`
		SELECT p.id, p.status, p.generated_at, p.posted_at, p.closed_at,
			closed_user.name, p.closing_notes, p.reopened_at, reopened_user.name, p.reopen_reason
		FROM asset_depreciation_periods p
		LEFT JOIN users closed_user ON closed_user.id = p.closed_by
		LEFT JOIN users reopened_user ON reopened_user.id = p.reopened_by
		WHERE p.period_year = ? AND p.period_month = ?
	`, year, month).Scan(
		&period.ID, &period.Status, &generatedAt, &postedAt, &closedAt,
		&closedByName, &closingNotes, &reopenedAt, &reopenedByName, &reopenReason,
	)
	if errors.Is(err, sql.ErrNoRows) {
		return period, nil
	}
	if err != nil {
		return period, err
	}
	period.CanClose = period.Status == "POSTED"
	if generatedAt.Valid {
		period.GeneratedAtDisplay = formatDepreciationDateID(generatedAt.Time, true)
	}
	if postedAt.Valid {
		period.PostedAtDisplay = formatDepreciationDateID(postedAt.Time, true)
	}
	if closedAt.Valid {
		period.ClosedAtDisplay = formatDepreciationDateID(closedAt.Time, true)
	}
	if reopenedAt.Valid {
		period.ReopenedAtDisplay = formatDepreciationDateID(reopenedAt.Time, true)
	}
	period.ClosedByName = closedByName.String
	period.ClosingNotes = closingNotes.String
	period.ReopenedByName = reopenedByName.String
	period.ReopenReason = reopenReason.String
	return period, nil
}

func (r *AssetDepreciationRepository) CloseDepreciationPeriod(year, month int, notes string, auditCtx models.AuditContext) error {
	tx, err := r.DB.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()
	periodID, status, err := ensureDepreciationPeriodTx(tx, year, month)
	if err != nil {
		return err
	}
	if status == "CLOSED" {
		return errors.New("periode depresiasi sudah ditutup")
	}
	calculatedStatus, totalAssets, err := calculateDepreciationPeriodStatusTx(tx, year, month)
	if err != nil {
		return err
	}
	if totalAssets == 0 {
		return errors.New("periode tanpa jadwal depresiasi tidak dapat ditutup")
	}
	if calculatedStatus != "POSTED" {
		return errors.New("periode belum dapat ditutup karena masih memiliki draft atau koreksi yang belum diselesaikan")
	}
	if _, err := tx.Exec(`
		UPDATE asset_depreciation_periods
		SET status = 'CLOSED', closed_at = CURRENT_TIMESTAMP, closed_by = ?, closing_notes = ?,
			updated_at = CURRENT_TIMESTAMP
		WHERE id = ?
	`, auditCtx.ActorUserID, nullableString(notes), periodID); err != nil {
		return err
	}
	if err := insertAuditLogTx(tx, "DEPRECIATION_PERIOD", periodID, "CLOSE_PERIOD",
		fmt.Sprintf("Periode depresiasi %s ditutup. Catatan: %s", formatDepreciationMonthYearID(time.Date(year, time.Month(month), 1, 0, 0, 0, 0, time.Local)), notes), auditCtx); err != nil {
		return err
	}
	return tx.Commit()
}

func (r *AssetDepreciationRepository) ReopenDepreciationPeriod(year, month int, reason string, auditCtx models.AuditContext) error {
	tx, err := r.DB.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()
	periodID, status, err := ensureDepreciationPeriodTx(tx, year, month)
	if err != nil {
		return err
	}
	if status != "CLOSED" {
		return errors.New("hanya periode CLOSED yang dapat dibuka kembali")
	}
	calculatedStatus, _, err := calculateDepreciationPeriodStatusTx(tx, year, month)
	if err != nil {
		return err
	}
	if _, err := tx.Exec(`
		UPDATE asset_depreciation_periods
		SET status = ?, reopened_at = CURRENT_TIMESTAMP, reopened_by = ?, reopen_reason = ?,
			closed_at = NULL, closed_by = NULL, closing_notes = NULL, updated_at = CURRENT_TIMESTAMP
		WHERE id = ?
	`, calculatedStatus, auditCtx.ActorUserID, reason, periodID); err != nil {
		return err
	}
	if err := insertAuditLogTx(tx, "DEPRECIATION_PERIOD", periodID, "REOPEN_PERIOD",
		fmt.Sprintf("Periode depresiasi %s dibuka kembali. Alasan: %s", formatDepreciationMonthYearID(time.Date(year, time.Month(month), 1, 0, 0, 0, 0, time.Local)), reason), auditCtx); err != nil {
		return err
	}
	return tx.Commit()
}

func (r *AssetDepreciationRepository) GenerateMonthlySchedules(year, month int, auditCtx models.AuditContext) (int, error) {
	periodDate := time.Date(year, time.Month(month), 1, 0, 0, 0, 0, time.Local)
	tx, err := r.DB.Begin()
	if err != nil {
		return 0, err
	}
	defer tx.Rollback()
	periodID, err := assertDepreciationPeriodOpenTx(tx, year, month)
	if err != nil {
		return 0, err
	}
	if _, err := tx.Exec(`
		UPDATE asset_depreciation_periods
		SET status = 'GENERATED', generated_at = CURRENT_TIMESTAMP, generated_by = ?, updated_at = CURRENT_TIMESTAMP
		WHERE id = ?
	`, auditCtx.ActorUserID, periodID); err != nil {
		return 0, err
	}

	rows, err := tx.Query(`
		SELECT
			adp.id,
			adp.asset_id,
			adp.useful_life_months,
			adp.depreciable_basis,
			adp.salvage_value,
			adp.start_date,
			first_policy.code,
			last_policy.code,
			disposal.disposal_date,
			COALESCE((
				SELECT SUM(schedule.depreciation_amount)
				FROM asset_depreciation_schedules schedule
				WHERE schedule.profile_id = adp.id
				  AND schedule.status = 'POSTED'
				  AND schedule.period_date < ?
			), 0),
			(
				SELECT COUNT(*)
				FROM asset_depreciation_schedules schedule
				WHERE schedule.profile_id = adp.id
				  AND schedule.status = 'DRAFT'
				  AND schedule.period_date < ?
			),
			(
				SELECT COUNT(*)
				FROM asset_depreciation_schedules schedule
				WHERE schedule.profile_id = adp.id
				  AND schedule.status IN ('DRAFT', 'POSTED', 'SKIPPED')
				  AND schedule.period_date > ?
			)
		FROM asset_depreciation_profiles adp
		JOIN asset_depreciation_methods method ON method.id = adp.depreciation_method_id
		JOIN asset_depreciation_first_month_policies first_policy ON first_policy.id = adp.first_month_policy_id
		JOIN asset_depreciation_last_month_policies last_policy ON last_policy.id = adp.last_month_policy_id
		JOIN assets asset ON asset.id = adp.asset_id
		LEFT JOIN asset_disposals disposal ON disposal.id = (
			SELECT MAX(candidate.id)
			FROM asset_disposals candidate
			WHERE candidate.asset_id = adp.asset_id AND candidate.status = 'DRAFT'
		)
		WHERE adp.status = 'ACTIVE'
		  AND method.code = 'STRAIGHT_LINE'
		  AND asset.status <> 'DISPOSED'
		  AND adp.useful_life_months > 0
		  AND adp.depreciable_basis > adp.salvage_value
		FOR UPDATE
	`, periodDate, periodDate, periodDate)
	if err != nil {
		return 0, err
	}
	candidates := make([]depreciationGenerationCandidate, 0)
	for rows.Next() {
		var candidate depreciationGenerationCandidate
		if err := rows.Scan(
			&candidate.ProfileID, &candidate.AssetID, &candidate.UsefulLife,
			&candidate.Basis, &candidate.Salvage, &candidate.StartDate,
			&candidate.FirstPolicyCode, &candidate.LastPolicyCode, &candidate.DisposalDate,
			&candidate.PostedAmount, &candidate.PriorDrafts, &candidate.LaterSchedules,
		); err != nil {
			rows.Close()
			return 0, err
		}
		candidates = append(candidates, candidate)
	}
	if err := rows.Close(); err != nil {
		return 0, err
	}
	if err := rows.Err(); err != nil {
		return 0, err
	}

	changedScheduleIDs := make([]int64, 0)
	for _, candidate := range candidates {
		if candidate.PriorDrafts > 0 || candidate.LaterSchedules > 0 {
			continue
		}
		effectiveStart := time.Date(candidate.StartDate.Year(), candidate.StartDate.Month(), 1, 0, 0, 0, 0, time.Local)
		if candidate.FirstPolicyCode == "NEXT_MONTH" {
			effectiveStart = effectiveStart.AddDate(0, 1, 0)
		}
		if periodDate.Before(effectiveStart) {
			continue
		}

		startDay := 1
		endDay := daysInDepreciationMonth(periodDate)
		startMonth := time.Date(candidate.StartDate.Year(), candidate.StartDate.Month(), 1, 0, 0, 0, 0, time.Local)
		if candidate.FirstPolicyCode == "PRORATE_DAILY" && periodDate.Equal(startMonth) {
			startDay = candidate.StartDate.Day()
		}
		if candidate.DisposalDate.Valid {
			disposalMonth := time.Date(candidate.DisposalDate.Time.Year(), candidate.DisposalDate.Time.Month(), 1, 0, 0, 0, 0, time.Local)
			if periodDate.After(disposalMonth) || (periodDate.Equal(disposalMonth) && candidate.LastPolicyCode == "NO_DEPRECIATION") {
				continue
			}
			if periodDate.Equal(disposalMonth) && candidate.LastPolicyCode == "PRORATE_DAILY" {
				endDay = candidate.DisposalDate.Time.Day()
			}
		}
		if endDay < startDay {
			continue
		}

		remaining := candidate.Basis - candidate.PostedAmount - candidate.Salvage
		if remaining <= 0.005 {
			if err := finishDepreciationProfileTx(tx, candidate.ProfileID, auditCtx); err != nil {
				return 0, err
			}
			continue
		}
		monthlyAmount := roundDepreciationAmount((candidate.Basis - candidate.Salvage) / float64(candidate.UsefulLife))
		factor := float64(endDay-startDay+1) / float64(daysInDepreciationMonth(periodDate))
		depreciationAmount := math.Min(remaining, roundDepreciationAmount(monthlyAmount*factor))
		if depreciationAmount <= 0.005 {
			continue
		}
		openingValue := math.Max(candidate.Salvage, candidate.Basis-candidate.PostedAmount)
		accumulatedValue := candidate.PostedAmount + depreciationAmount
		closingValue := math.Max(candidate.Salvage, openingValue-depreciationAmount)

		var existingID, originalScheduleID int64
		var existingStatus string
		err := tx.QueryRow(`
			SELECT id, status, COALESCE(original_schedule_id, 0)
			FROM asset_depreciation_schedules
			WHERE asset_id = ? AND period_year = ? AND period_month = ?
			ORDER BY version_no DESC
			LIMIT 1
			FOR UPDATE
		`, candidate.AssetID, year, month).Scan(&existingID, &existingStatus, &originalScheduleID)
		switch {
		case errors.Is(err, sql.ErrNoRows):
			result, err := tx.Exec(`
				INSERT INTO asset_depreciation_schedules (
					profile_id, asset_id, period_year, period_month, period_date,
					opening_book_value, depreciation_amount, accumulated_depreciation,
					closing_book_value, status
				) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, 'DRAFT')
			`, candidate.ProfileID, candidate.AssetID, year, month, periodDate,
				openingValue, depreciationAmount, accumulatedValue, closingValue)
			if err != nil {
				return 0, err
			}
			existingID, err = result.LastInsertId()
			if err != nil {
				return 0, err
			}
			changedScheduleIDs = append(changedScheduleIDs, existingID)
		case err != nil:
			return 0, err
		case existingStatus == "DRAFT" && originalScheduleID == 0:
			if _, err := tx.Exec(`
				UPDATE asset_depreciation_schedules
				SET opening_book_value = ?, depreciation_amount = ?, accumulated_depreciation = ?,
					closing_book_value = ?, updated_at = CURRENT_TIMESTAMP
				WHERE id = ? AND status = 'DRAFT'
			`, openingValue, depreciationAmount, accumulatedValue, closingValue, existingID); err != nil {
				return 0, err
			}
			changedScheduleIDs = append(changedScheduleIDs, existingID)
		}
	}

	var scheduleCount int
	if err := tx.QueryRow(`
		SELECT COUNT(DISTINCT asset_id)
		FROM asset_depreciation_schedules
		WHERE period_year = ? AND period_month = ?
	`, year, month).Scan(&scheduleCount); err != nil {
		return 0, err
	}

	message := fmt.Sprintf("Jadwal depresiasi periode %s dibuat atau diperbarui sebagai DRAFT", formatDepreciationMonthYearID(periodDate))
	for _, scheduleID := range changedScheduleIDs {
		if err := insertAuditLogTx(tx, "ASSET_DEPRECIATION", scheduleID, "GENERATE", message, auditCtx); err != nil {
			return 0, err
		}
	}
	if err := syncDepreciationPeriodStatusTx(tx, year, month, auditCtx.ActorUserID); err != nil {
		return 0, err
	}
	if err := tx.Commit(); err != nil {
		return 0, err
	}
	return scheduleCount, nil
}

func daysInDepreciationMonth(period time.Time) int {
	return time.Date(period.Year(), period.Month()+1, 0, 0, 0, 0, 0, period.Location()).Day()
}

func roundDepreciationAmount(value float64) float64 {
	return math.Round(value*100) / 100
}

func finishDepreciationProfileTx(tx *sql.Tx, profileID int64, auditCtx models.AuditContext) error {
	result, err := tx.Exec(`
		UPDATE asset_depreciation_profiles
		SET status = 'FINISHED', finished_at = COALESCE(finished_at, CURRENT_TIMESTAMP), updated_at = CURRENT_TIMESTAMP
		WHERE id = ? AND status = 'ACTIVE'
	`, profileID)
	if err != nil {
		return err
	}
	affected, err := result.RowsAffected()
	if err != nil || affected == 0 {
		return err
	}
	return insertAuditLogTx(tx, "DEPRECIATION_PROFILE", profileID, "FINISH_PROFILE",
		"Profil otomatis selesai karena nilai buku telah mencapai nilai residu", auditCtx)
}

func finishCompletedDepreciationProfilesTx(tx *sql.Tx, records []depreciationScheduleActionRecord, auditCtx models.AuditContext) error {
	seen := make(map[int64]bool, len(records))
	for _, record := range records {
		if record.ProfileID <= 0 || seen[record.ProfileID] {
			continue
		}
		seen[record.ProfileID] = true
		var status string
		var basis, salvage, posted float64
		if err := tx.QueryRow(`
			SELECT adp.status, adp.depreciable_basis, adp.salvage_value,
				COALESCE(SUM(CASE WHEN schedule.status = 'POSTED' THEN schedule.depreciation_amount ELSE 0 END), 0)
			FROM asset_depreciation_profiles adp
			LEFT JOIN asset_depreciation_schedules schedule ON schedule.profile_id = adp.id
			WHERE adp.id = ?
			GROUP BY adp.id, adp.status, adp.depreciable_basis, adp.salvage_value
			FOR UPDATE
		`, record.ProfileID).Scan(&status, &basis, &salvage, &posted); err != nil {
			return err
		}
		if status == "ACTIVE" && basis-posted <= salvage+0.005 {
			if err := finishDepreciationProfileTx(tx, record.ProfileID, auditCtx); err != nil {
				return err
			}
		}
	}
	return nil
}

func (r *AssetDepreciationRepository) PostSchedules(ids []int64, auditCtx models.AuditContext) (int64, error) {
	if len(ids) == 0 {
		return 0, errors.New("pilih minimal satu depresiasi yang akan diposting")
	}

	tx, err := r.DB.Begin()
	if err != nil {
		return 0, err
	}
	defer tx.Rollback()

	placeholders := make([]string, len(ids))
	args := make([]any, len(ids))
	for i, id := range ids {
		placeholders[i] = "?"
		args[i] = id
	}
	query := `
		SELECT id, profile_id, status, period_date
		FROM asset_depreciation_schedules
		WHERE id IN (` + strings.Join(placeholders, ",") + `)
		FOR UPDATE`
	rows, err := tx.Query(query, args...)
	if err != nil {
		return 0, err
	}

	currentPeriod := time.Date(time.Now().Year(), time.Now().Month(), 1, 0, 0, 0, 0, time.Local)
	found := 0
	records := make([]depreciationScheduleActionRecord, 0, len(ids))
	for rows.Next() {
		var id int64
		var status string
		var periodDate time.Time
		var profileID int64
		if err := rows.Scan(&id, &profileID, &status, &periodDate); err != nil {
			rows.Close()
			return 0, err
		}
		found++
		if status != "DRAFT" {
			rows.Close()
			return 0, fmt.Errorf("depresiasi ID %d sudah berstatus %s", id, status)
		}
		period := time.Date(periodDate.Year(), periodDate.Month(), 1, 0, 0, 0, 0, time.Local)
		if period.After(currentPeriod) {
			rows.Close()
			return 0, fmt.Errorf("depresiasi periode %s belum dapat diposting", formatDepreciationMonthYearID(periodDate))
		}
		records = append(records, depreciationScheduleActionRecord{ID: id, ProfileID: profileID, PeriodDate: periodDate})
	}
	if err := rows.Close(); err != nil {
		return 0, err
	}
	if found != len(ids) {
		return 0, errors.New("sebagian data depresiasi tidak ditemukan")
	}
	if err := assertDepreciationActionPeriodsOpenTx(tx, records); err != nil {
		return 0, err
	}

	updateQuery := `
		UPDATE asset_depreciation_schedules
		SET status = 'POSTED', posted_at = CURRENT_TIMESTAMP, posted_by = ?,
			skipped_at = NULL, skipped_by = NULL, skip_reason = NULL, updated_at = CURRENT_TIMESTAMP
		WHERE status = 'DRAFT' AND id IN (` + strings.Join(placeholders, ",") + `)`
	updateArgs := append([]any{auditCtx.ActorUserID}, args...)
	result, err := tx.Exec(updateQuery, updateArgs...)
	if err != nil {
		return 0, err
	}
	affected, err := result.RowsAffected()
	if err != nil {
		return 0, err
	}
	if affected != int64(len(ids)) {
		return 0, errors.New("posting dibatalkan karena ada data yang berubah")
	}
	for _, record := range records {
		message := fmt.Sprintf("Depresiasi periode %s diposting", formatDepreciationMonthYearID(record.PeriodDate))
		if err := insertAuditLogTx(tx, "ASSET_DEPRECIATION", record.ID, "POST", message, auditCtx); err != nil {
			return 0, err
		}
	}
	if err := finishCompletedDepreciationProfilesTx(tx, records, auditCtx); err != nil {
		return 0, err
	}
	if err := syncDepreciationActionPeriodsTx(tx, records, auditCtx.ActorUserID); err != nil {
		return 0, err
	}
	if err := tx.Commit(); err != nil {
		return 0, err
	}
	return affected, nil
}

func (r *AssetDepreciationRepository) SkipSchedules(ids []int64, reason string, auditCtx models.AuditContext) (int64, error) {
	if len(ids) == 0 {
		return 0, errors.New("pilih minimal satu depresiasi yang akan dilewati")
	}

	tx, err := r.DB.Begin()
	if err != nil {
		return 0, err
	}
	defer tx.Rollback()

	placeholders := make([]string, len(ids))
	args := make([]any, len(ids))
	for i, id := range ids {
		placeholders[i] = "?"
		args[i] = id
	}
	rows, err := tx.Query(`
		SELECT id, profile_id, status, period_date
		FROM asset_depreciation_schedules
		WHERE id IN (`+strings.Join(placeholders, ",")+`)
		FOR UPDATE
	`, args...)
	if err != nil {
		return 0, err
	}

	currentPeriod := time.Date(time.Now().Year(), time.Now().Month(), 1, 0, 0, 0, 0, time.Local)
	records := make([]depreciationScheduleActionRecord, 0, len(ids))
	for rows.Next() {
		var record depreciationScheduleActionRecord
		var status string
		if err := rows.Scan(&record.ID, &record.ProfileID, &status, &record.PeriodDate); err != nil {
			rows.Close()
			return 0, err
		}
		if status != "DRAFT" {
			rows.Close()
			return 0, fmt.Errorf("depresiasi ID %d sudah berstatus %s", record.ID, status)
		}
		period := time.Date(record.PeriodDate.Year(), record.PeriodDate.Month(), 1, 0, 0, 0, 0, time.Local)
		if period.After(currentPeriod) {
			rows.Close()
			return 0, fmt.Errorf("depresiasi periode %s belum dapat dilewati", formatDepreciationMonthYearID(record.PeriodDate))
		}
		records = append(records, record)
	}
	if err := rows.Close(); err != nil {
		return 0, err
	}
	if err := rows.Err(); err != nil {
		return 0, err
	}
	if len(records) != len(ids) {
		return 0, errors.New("sebagian data depresiasi tidak ditemukan")
	}
	if err := assertDepreciationActionPeriodsOpenTx(tx, records); err != nil {
		return 0, err
	}

	updateArgs := []any{auditCtx.ActorUserID, reason}
	updateArgs = append(updateArgs, args...)
	result, err := tx.Exec(`
		UPDATE asset_depreciation_schedules
		SET status = 'SKIPPED',
			accumulated_depreciation = GREATEST(0, accumulated_depreciation - depreciation_amount),
			closing_book_value = opening_book_value,
			depreciation_amount = 0,
			posted_at = NULL, posted_by = NULL,
			skipped_at = CURRENT_TIMESTAMP, skipped_by = ?, skip_reason = ?,
			updated_at = CURRENT_TIMESTAMP
		WHERE status = 'DRAFT' AND id IN (`+strings.Join(placeholders, ",")+`)
	`, updateArgs...)
	if err != nil {
		return 0, err
	}
	affected, err := result.RowsAffected()
	if err != nil {
		return 0, err
	}
	if affected != int64(len(ids)) {
		return 0, errors.New("proses lewati dibatalkan karena ada data yang berubah")
	}
	for _, record := range records {
		message := fmt.Sprintf("Depresiasi periode %s dilewati. Alasan: %s", formatDepreciationMonthYearID(record.PeriodDate), reason)
		if err := insertAuditLogTx(tx, "ASSET_DEPRECIATION", record.ID, "SKIP", message, auditCtx); err != nil {
			return 0, err
		}
	}
	if err := syncDepreciationActionPeriodsTx(tx, records, auditCtx.ActorUserID); err != nil {
		return 0, err
	}
	if err := tx.Commit(); err != nil {
		return 0, err
	}
	return affected, nil
}

func (r *AssetDepreciationRepository) ReverseSchedule(scheduleID int64, reason string, auditCtx models.AuditContext) (int64, error) {
	tx, err := r.DB.Begin()
	if err != nil {
		return 0, err
	}
	defer tx.Rollback()

	var profileID, assetID int64
	var periodYear, periodMonth, versionNo int
	var periodDate time.Time
	var openingValue, depreciationAmount, accumulatedValue, closingValue float64
	var status, assetStatus string
	if err := tx.QueryRow(`
		SELECT profile_id, asset_id, period_year, period_month, period_date, version_no,
			opening_book_value, depreciation_amount, accumulated_depreciation,
			closing_book_value, schedule.status, asset.status
		FROM asset_depreciation_schedules schedule
		JOIN assets asset ON asset.id = schedule.asset_id
		WHERE schedule.id = ?
		FOR UPDATE
	`, scheduleID).Scan(
		&profileID, &assetID, &periodYear, &periodMonth, &periodDate, &versionNo,
		&openingValue, &depreciationAmount, &accumulatedValue, &closingValue, &status, &assetStatus,
	); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return 0, errors.New("posting depresiasi tidak ditemukan")
		}
		return 0, err
	}
	if status != "POSTED" {
		return 0, errors.New("hanya depresiasi berstatus POSTED yang dapat dibatalkan")
	}
	if assetStatus == "DISPOSED" {
		return 0, errors.New("batalkan disposal aset terlebih dahulu sebelum melakukan reversal depresiasi")
	}
	if _, err := assertDepreciationPeriodOpenTx(tx, periodYear, periodMonth); err != nil {
		return 0, err
	}

	var laterScheduleCount int
	if err := tx.QueryRow(`
		SELECT COUNT(*)
		FROM asset_depreciation_schedules
		WHERE asset_id = ? AND period_date > ?
			AND status IN ('DRAFT', 'POSTED', 'SKIPPED')
	`, assetID, periodDate).Scan(&laterScheduleCount); err != nil {
		return 0, err
	}
	if laterScheduleCount > 0 {
		return 0, errors.New("batalkan periode depresiasi terbaru terlebih dahulu sebelum membatalkan periode ini")
	}

	var samePeriodOpenCount, maxVersion int
	if err := tx.QueryRow(`
		SELECT
			SUM(CASE WHEN id <> ? AND status IN ('DRAFT', 'POSTED') THEN 1 ELSE 0 END),
			MAX(version_no)
		FROM asset_depreciation_schedules
		WHERE asset_id = ? AND period_year = ? AND period_month = ?
	`, scheduleID, assetID, periodYear, periodMonth).Scan(&samePeriodOpenCount, &maxVersion); err != nil {
		return 0, err
	}
	if samePeriodOpenCount > 0 {
		return 0, errors.New("periode ini sudah memiliki draft atau posting versi lain")
	}

	if _, err := tx.Exec(`
		UPDATE asset_depreciation_schedules
		SET status = 'REVERSED', reversed_at = CURRENT_TIMESTAMP, reversed_by = ?,
			reversal_reason = ?, updated_at = CURRENT_TIMESTAMP
		WHERE id = ? AND status = 'POSTED'
	`, auditCtx.ActorUserID, reason, scheduleID); err != nil {
		return 0, err
	}

	result, err := tx.Exec(`
		INSERT INTO asset_depreciation_schedules (
			profile_id, asset_id, period_year, period_month, period_date, version_no,
			original_schedule_id, correction_reason, opening_book_value,
			depreciation_amount, accumulated_depreciation, closing_book_value, status
		) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, 'DRAFT')
	`, profileID, assetID, periodYear, periodMonth, periodDate, maxVersion+1,
		scheduleID, reason, openingValue, depreciationAmount, accumulatedValue, closingValue)
	if err != nil {
		return 0, err
	}
	correctionID, err := result.LastInsertId()
	if err != nil {
		return 0, err
	}

	periodLabel := formatDepreciationMonthYearID(periodDate)
	if err := insertAuditLogTx(tx, "ASSET_DEPRECIATION", scheduleID, "REVERSE",
		fmt.Sprintf("Posting depresiasi periode %s dibatalkan. Alasan: %s", periodLabel, reason), auditCtx); err != nil {
		return 0, err
	}
	if err := insertAuditLogTx(tx, "ASSET_DEPRECIATION", correctionID, "CREATE_CORRECTION_DRAFT",
		fmt.Sprintf("Draft koreksi versi %d dibuat dari jadwal ID %d", maxVersion+1, scheduleID), auditCtx); err != nil {
		return 0, err
	}
	result, err = tx.Exec(`
		UPDATE asset_depreciation_profiles
		SET status = 'ACTIVE', finished_at = NULL, updated_at = CURRENT_TIMESTAMP
		WHERE id = ? AND status = 'FINISHED'
	`, profileID)
	if err != nil {
		return 0, err
	}
	if affected, err := result.RowsAffected(); err != nil {
		return 0, err
	} else if affected > 0 {
		if err := insertAuditLogTx(tx, "DEPRECIATION_PROFILE", profileID, "REOPEN_AFTER_REVERSAL",
			"Profil diaktifkan kembali karena posting depresiasi terakhir dibatalkan", auditCtx); err != nil {
			return 0, err
		}
	}
	if err := syncDepreciationPeriodStatusTx(tx, periodYear, periodMonth, auditCtx.ActorUserID); err != nil {
		return 0, err
	}
	if err := tx.Commit(); err != nil {
		return 0, err
	}
	return correctionID, nil
}

func (r *AssetDepreciationRepository) UpdateCorrectionDraft(input models.DepreciationCorrectionInput) error {
	tx, err := r.DB.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	var profileID, assetID, originalScheduleID int64
	var periodDate time.Time
	var openingValue, salvageValue float64
	var status string
	if err := tx.QueryRow(`
		SELECT ads.profile_id, ads.asset_id, COALESCE(ads.original_schedule_id, 0),
			ads.period_date, ads.opening_book_value, adp.salvage_value, ads.status
		FROM asset_depreciation_schedules ads
		JOIN asset_depreciation_profiles adp ON adp.id = ads.profile_id
		WHERE ads.id = ?
		FOR UPDATE
	`, input.ScheduleID).Scan(
		&profileID, &assetID, &originalScheduleID, &periodDate, &openingValue, &salvageValue, &status,
	); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return errors.New("draft koreksi tidak ditemukan")
		}
		return err
	}
	if status != "DRAFT" || originalScheduleID <= 0 {
		return errors.New("hanya draft hasil reversal yang dapat dikoreksi")
	}
	if _, err := assertDepreciationPeriodOpenTx(tx, periodDate.Year(), int(periodDate.Month())); err != nil {
		return err
	}
	maxDepreciation := math.Max(0, openingValue-salvageValue)
	if input.DepreciationValue > maxDepreciation+0.005 {
		return errors.New("nilai depresiasi koreksi melebihi nilai yang masih dapat disusutkan")
	}

	var laterScheduleCount int
	if err := tx.QueryRow(`
		SELECT COUNT(*)
		FROM asset_depreciation_schedules
		WHERE asset_id = ? AND period_date > ?
			AND status IN ('DRAFT', 'POSTED', 'SKIPPED')
	`, assetID, periodDate).Scan(&laterScheduleCount); err != nil {
		return err
	}
	if laterScheduleCount > 0 {
		return errors.New("draft koreksi tidak dapat diubah karena terdapat periode depresiasi yang lebih baru")
	}

	var previousAccumulated float64
	if err := tx.QueryRow(`
		SELECT COALESCE(SUM(depreciation_amount), 0)
		FROM asset_depreciation_schedules
		WHERE profile_id = ? AND period_date < ? AND status = 'POSTED'
	`, profileID, periodDate).Scan(&previousAccumulated); err != nil {
		return err
	}
	accumulatedValue := previousAccumulated + input.DepreciationValue
	closingValue := math.Max(salvageValue, openingValue-input.DepreciationValue)
	if _, err := tx.Exec(`
		UPDATE asset_depreciation_schedules
		SET depreciation_amount = ?, accumulated_depreciation = ?, closing_book_value = ?,
			correction_reason = ?, updated_at = CURRENT_TIMESTAMP
		WHERE id = ? AND status = 'DRAFT'
	`, input.DepreciationValue, accumulatedValue, closingValue, input.Reason, input.ScheduleID); err != nil {
		return err
	}
	if err := insertAuditLogTx(tx, "ASSET_DEPRECIATION", input.ScheduleID, "UPDATE_CORRECTION",
		fmt.Sprintf("Draft koreksi diperbarui menjadi %s. Alasan: %s", formatAssetAmountID(input.DepreciationValue), input.Reason), input.AuditContext); err != nil {
		return err
	}
	return tx.Commit()
}

func ensureDepreciationPeriodTx(tx *sql.Tx, year, month int) (int64, string, error) {
	if _, err := tx.Exec(`
		INSERT INTO asset_depreciation_periods (period_year, period_month, status)
		VALUES (?, ?, 'OPEN')
		ON DUPLICATE KEY UPDATE id = id
	`, year, month); err != nil {
		return 0, "", err
	}
	var periodID int64
	var status string
	if err := tx.QueryRow(`
		SELECT id, status
		FROM asset_depreciation_periods
		WHERE period_year = ? AND period_month = ?
		FOR UPDATE
	`, year, month).Scan(&periodID, &status); err != nil {
		return 0, "", err
	}
	return periodID, status, nil
}

func assertDepreciationPeriodOpenTx(tx *sql.Tx, year, month int) (int64, error) {
	periodID, status, err := ensureDepreciationPeriodTx(tx, year, month)
	if err != nil {
		return 0, err
	}
	if status == "CLOSED" {
		return 0, fmt.Errorf("periode depresiasi %s sudah CLOSED dan tidak dapat diubah", formatDepreciationMonthYearID(time.Date(year, time.Month(month), 1, 0, 0, 0, 0, time.Local)))
	}
	return periodID, nil
}

func calculateDepreciationPeriodStatusTx(tx *sql.Tx, year, month int) (string, int, error) {
	var totalAssets, unresolved int
	err := tx.QueryRow(`
		SELECT
			COUNT(*),
			COALESCE(SUM(latest.status NOT IN ('POSTED', 'SKIPPED')), 0)
		FROM asset_depreciation_schedules latest
		WHERE latest.period_year = ? AND latest.period_month = ?
			AND latest.version_no = (
				SELECT MAX(candidate.version_no)
				FROM asset_depreciation_schedules candidate
				WHERE candidate.asset_id = latest.asset_id
					AND candidate.period_year = latest.period_year
					AND candidate.period_month = latest.period_month
			)
	`, year, month).Scan(&totalAssets, &unresolved)
	if err != nil {
		return "", 0, err
	}
	if totalAssets == 0 {
		return "OPEN", 0, nil
	}
	if unresolved > 0 {
		return "GENERATED", totalAssets, nil
	}
	return "POSTED", totalAssets, nil
}

func syncDepreciationPeriodStatusTx(tx *sql.Tx, year, month, actorUserID int) error {
	periodID, currentStatus, err := ensureDepreciationPeriodTx(tx, year, month)
	if err != nil {
		return err
	}
	if currentStatus == "CLOSED" {
		return errors.New("periode CLOSED tidak dapat diperbarui")
	}
	status, _, err := calculateDepreciationPeriodStatusTx(tx, year, month)
	if err != nil {
		return err
	}
	if status == "OPEN" && currentStatus == "GENERATED" {
		status = "GENERATED"
	}
	if _, err := tx.Exec(`
		UPDATE asset_depreciation_periods
		SET status = ?,
			posted_at = CASE WHEN ? = 'POSTED' THEN CURRENT_TIMESTAMP ELSE NULL END,
			posted_by = CASE WHEN ? = 'POSTED' THEN ? ELSE NULL END,
			updated_at = CURRENT_TIMESTAMP
		WHERE id = ?
	`, status, status, status, actorUserID, periodID); err != nil {
		return err
	}
	return nil
}

func depreciationActionPeriods(records []depreciationScheduleActionRecord) [][2]int {
	periods := make([][2]int, 0)
	seen := make(map[[2]int]bool)
	for _, record := range records {
		period := [2]int{record.PeriodDate.Year(), int(record.PeriodDate.Month())}
		if seen[period] {
			continue
		}
		seen[period] = true
		periods = append(periods, period)
	}
	return periods
}

func assertDepreciationActionPeriodsOpenTx(tx *sql.Tx, records []depreciationScheduleActionRecord) error {
	for _, period := range depreciationActionPeriods(records) {
		if _, err := assertDepreciationPeriodOpenTx(tx, period[0], period[1]); err != nil {
			return err
		}
	}
	return nil
}

func syncDepreciationActionPeriodsTx(tx *sql.Tx, records []depreciationScheduleActionRecord, actorUserID int) error {
	for _, period := range depreciationActionPeriods(records) {
		if err := syncDepreciationPeriodStatusTx(tx, period[0], period[1], actorUserID); err != nil {
			return err
		}
	}
	return nil
}

func monthlyDepreciationWhere(filter models.MonthlyDepreciationFilter) (string, []any) {
	clauses := []string{"ads.period_year = ?", "ads.period_month = ?"}
	args := []any{filter.Year, filter.Month}
	if filter.Status != "" && filter.Status != "ALL" {
		clauses = append(clauses, "ads.status = ?")
		args = append(args, filter.Status)
	}
	if filter.Search != "" {
		clauses = append(clauses, "(a.asset_code LIKE ? OR a.asset_name LIKE ? OR at.name LIKE ?)")
		term := "%" + filter.Search + "%"
		args = append(args, term, term, term)
	}
	return strings.Join(clauses, " AND "), args
}
