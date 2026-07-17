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
	PeriodDate time.Time
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
			adp.status,
			COALESCE(adp.notes, ''),
			COALESCE(schedule.posted_amount, 0),
			COALESCE(schedule.posted_count, 0),
			COALESCE(schedule.draft_count, 0),
			schedule.last_posted_period
		FROM asset_depreciation_profiles adp
		JOIN assets a ON a.id = adp.asset_id
		LEFT JOIN asset_types at ON at.id = a.asset_type_id
		JOIN asset_depreciation_methods adm ON adm.id = adp.depreciation_method_id
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
		var lastPostedPeriod sql.NullTime
		if err := rows.Scan(
			&item.ID, &item.AssetID, &item.AssetCode, &item.AssetName, &item.AssetTypeName,
			&item.MethodID, &item.MethodCode, &item.MethodName, &item.UsefulLifeMonths,
			&salvageValue, &depreciableBasis, &startDate, &item.Status, &item.Notes,
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
			(SELECT COUNT(*) FROM assets a WHERE a.status <> 'DISPOSED' AND NOT EXISTS (
				SELECT 1 FROM asset_depreciation_profiles profile WHERE profile.asset_id = a.id
			))
		FROM asset_depreciation_profiles
	`).Scan(&stats.TotalProfiles, &stats.ActiveProfiles, &stats.PausedProfiles, &stats.FinishedProfiles, &stats.UnconfiguredAssets)
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
				depreciable_basis, start_date, status, notes
			) VALUES (?, ?, ?, ?, ?, ?, ?, ?)
		`, input.AssetID, input.MethodID, nullableUsefulLife(input.UsefulLifeMonths), input.SalvageValue,
			input.DepreciableBasis, input.StartDate, input.Status, nullableString(input.Notes))
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

	var existingAssetID, existingMethodID int64
	var existingLife sql.NullInt64
	var existingSalvage, existingBasis float64
	var existingStart time.Time
	if err := tx.QueryRow(`
		SELECT asset_id, depreciation_method_id, useful_life_months, salvage_value, depreciable_basis, start_date
		FROM asset_depreciation_profiles
		WHERE id = ?
		FOR UPDATE
	`, input.ID).Scan(&existingAssetID, &existingMethodID, &existingLife, &existingSalvage, &existingBasis, &existingStart); err != nil {
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
		existingStart.Format("2006-01-02") != input.StartDate
	if postedCount > 0 && configurationChanged {
		return errors.New("konfigurasi utama tidak dapat diubah karena profil sudah memiliki depresiasi yang diposting")
	}

	if _, err := tx.Exec(`
		UPDATE asset_depreciation_profiles
		SET depreciation_method_id = ?, useful_life_months = ?, salvage_value = ?,
			depreciable_basis = ?, start_date = ?, status = ?, notes = ?
		WHERE id = ?
	`, input.MethodID, nullableUsefulLife(input.UsefulLifeMonths), input.SalvageValue,
		input.DepreciableBasis, input.StartDate, input.Status, nullableString(input.Notes), input.ID); err != nil {
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

	_, err = tx.Exec(`
		INSERT INTO asset_depreciation_schedules (
			profile_id,
			asset_id,
			period_year,
			period_month,
			period_date,
			opening_book_value,
			depreciation_amount,
			accumulated_depreciation,
			closing_book_value,
			status
		)
		SELECT
			adp.id,
			adp.asset_id,
			YEAR(?),
			MONTH(?),
			?,
			GREATEST(
				adp.salvage_value,
				adp.depreciable_basis - COALESCE(posted.posted_amount, 0)
			) AS opening_book_value,
			LEAST(
				ROUND((adp.depreciable_basis - adp.salvage_value) / adp.useful_life_months, 2),
				GREATEST(
					0,
					adp.depreciable_basis - COALESCE(posted.posted_amount, 0) - adp.salvage_value
				)
			) AS depreciation_amount,
			COALESCE(posted.posted_amount, 0) + LEAST(
				ROUND((adp.depreciable_basis - adp.salvage_value) / adp.useful_life_months, 2),
				GREATEST(
					0,
					adp.depreciable_basis - COALESCE(posted.posted_amount, 0) - adp.salvage_value
				)
			) AS accumulated_depreciation,
			GREATEST(
				adp.salvage_value,
				adp.depreciable_basis - COALESCE(posted.posted_amount, 0) - LEAST(
					ROUND((adp.depreciable_basis - adp.salvage_value) / adp.useful_life_months, 2),
					GREATEST(
						0,
						adp.depreciable_basis - COALESCE(posted.posted_amount, 0) - adp.salvage_value
					)
				)
			) AS closing_book_value,
			'DRAFT'
		FROM asset_depreciation_profiles adp
		JOIN asset_depreciation_methods adm ON adm.id = adp.depreciation_method_id
		JOIN assets a ON a.id = adp.asset_id
		LEFT JOIN (
			SELECT
				profile_id,
				SUM(depreciation_amount) AS posted_amount
			FROM asset_depreciation_schedules
			WHERE status = 'POSTED' AND period_date < ?
			GROUP BY profile_id
		) posted ON posted.profile_id = adp.id
		LEFT JOIN (
			SELECT
				profile_id,
				COUNT(*) AS finalized_periods
			FROM asset_depreciation_schedules
			WHERE status IN ('POSTED', 'SKIPPED') AND period_date < ?
			GROUP BY profile_id
		) finalized ON finalized.profile_id = adp.id
		WHERE adp.status = 'ACTIVE'
			AND adm.code = 'STRAIGHT_LINE'
			AND a.status <> 'DISPOSED'
			AND adp.useful_life_months IS NOT NULL
			AND adp.useful_life_months > 0
			AND adp.depreciable_basis > adp.salvage_value
			AND ? >= DATE_SUB(adp.start_date, INTERVAL DAYOFMONTH(adp.start_date) - 1 DAY)
			AND TIMESTAMPDIFF(
				MONTH,
				DATE_SUB(adp.start_date, INTERVAL DAYOFMONTH(adp.start_date) - 1 DAY),
				?
			) < adp.useful_life_months
			AND COALESCE(finalized.finalized_periods, 0) = TIMESTAMPDIFF(
				MONTH,
				DATE_SUB(adp.start_date, INTERVAL DAYOFMONTH(adp.start_date) - 1 DAY),
				?
			)
		ON DUPLICATE KEY UPDATE
			opening_book_value = IF(asset_depreciation_schedules.status = 'DRAFT', VALUES(opening_book_value), asset_depreciation_schedules.opening_book_value),
			depreciation_amount = IF(asset_depreciation_schedules.status = 'DRAFT', VALUES(depreciation_amount), asset_depreciation_schedules.depreciation_amount),
			accumulated_depreciation = IF(asset_depreciation_schedules.status = 'DRAFT', VALUES(accumulated_depreciation), asset_depreciation_schedules.accumulated_depreciation),
			closing_book_value = IF(asset_depreciation_schedules.status = 'DRAFT', VALUES(closing_book_value), asset_depreciation_schedules.closing_book_value),
			updated_at = IF(asset_depreciation_schedules.status = 'DRAFT', CURRENT_TIMESTAMP, asset_depreciation_schedules.updated_at)
	`, periodDate, periodDate, periodDate, periodDate, periodDate, periodDate, periodDate, periodDate)
	if err != nil {
		return 0, err
	}

	var scheduleCount int
	if err := tx.QueryRow(`
		SELECT COUNT(DISTINCT asset_id)
		FROM asset_depreciation_schedules
		WHERE period_year = ? AND period_month = ?
	`, year, month).Scan(&scheduleCount); err != nil {
		return 0, err
	}

	rows, err := tx.Query(`
		SELECT id
		FROM asset_depreciation_schedules
		WHERE period_year = ? AND period_month = ? AND status = 'DRAFT'
		ORDER BY id
	`, year, month)
	if err != nil {
		return 0, err
	}
	scheduleIDs := make([]int64, 0)
	for rows.Next() {
		var scheduleID int64
		if err := rows.Scan(&scheduleID); err != nil {
			rows.Close()
			return 0, err
		}
		scheduleIDs = append(scheduleIDs, scheduleID)
	}
	if err := rows.Close(); err != nil {
		return 0, err
	}
	if err := rows.Err(); err != nil {
		return 0, err
	}
	message := fmt.Sprintf("Jadwal depresiasi periode %s dibuat atau diperbarui sebagai DRAFT", formatDepreciationMonthYearID(periodDate))
	for _, scheduleID := range scheduleIDs {
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
		SELECT id, status, period_date
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
		if err := rows.Scan(&id, &status, &periodDate); err != nil {
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
		records = append(records, depreciationScheduleActionRecord{ID: id, PeriodDate: periodDate})
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
		SELECT id, status, period_date
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
		if err := rows.Scan(&record.ID, &status, &record.PeriodDate); err != nil {
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
	var status string
	if err := tx.QueryRow(`
		SELECT profile_id, asset_id, period_year, period_month, period_date, version_no,
			opening_book_value, depreciation_amount, accumulated_depreciation,
			closing_book_value, status
		FROM asset_depreciation_schedules
		WHERE id = ?
		FOR UPDATE
	`, scheduleID).Scan(
		&profileID, &assetID, &periodYear, &periodMonth, &periodDate, &versionNo,
		&openingValue, &depreciationAmount, &accumulatedValue, &closingValue, &status,
	); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return 0, errors.New("posting depresiasi tidak ditemukan")
		}
		return 0, err
	}
	if status != "POSTED" {
		return 0, errors.New("hanya depresiasi berstatus POSTED yang dapat dibatalkan")
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
