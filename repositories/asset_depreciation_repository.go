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
		detail.StartDateDisplay = startDate.Time.Format("02 Jan 2006")
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
		detail.LastPostedPeriodDisplay = lastPostedPeriod.Time.Format("Jan 2006")
	}
	if nextDraftPeriod.Valid {
		detail.NextDraftPeriodDisplay = nextDraftPeriod.Time.Format("Jan 2006")
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
		posting.PeriodDisplay = periodDate.Format("Jan 2006")
		posting.OpeningBookValueDisplay = formatAssetAmountID(openingValue)
		posting.DepreciationAmountDisplay = formatAssetAmountID(depreciationAmount)
		posting.AccumulatedDepreciationDisplay = formatAssetAmountID(accumulatedValue)
		posting.ClosingBookValueDisplay = formatAssetAmountID(closingValue)
		if postedAt.Valid {
			posting.PostedAtDisplay = postedAt.Time.Format("02 Jan 2006 15:04")
		}
		postings = append(postings, posting)
	}
	return postings, rows.Err()
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
			COALESCE(adm.name, ''),
			COALESCE(adp.useful_life_months, 0),
			ads.period_date,
			a.acquisition_value,
			ads.opening_book_value,
			ads.depreciation_amount,
			ads.accumulated_depreciation,
			ads.closing_book_value,
			ads.status,
			ads.posted_at
		FROM asset_depreciation_schedules ads
		JOIN assets a ON a.id = ads.asset_id
		LEFT JOIN asset_types at ON at.id = a.asset_type_id
		LEFT JOIN asset_depreciation_profiles adp ON adp.id = ads.profile_id
		LEFT JOIN asset_depreciation_methods adm ON adm.id = adp.depreciation_method_id
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
		var postedAt sql.NullTime
		var acquisitionValue, openingValue, depreciationAmount, accumulatedValue, closingValue float64
		if err := rows.Scan(
			&item.ID,
			&item.AssetID,
			&item.AssetCode,
			&item.AssetName,
			&item.AssetTypeName,
			&item.MethodName,
			&item.UsefulLifeMonths,
			&periodDate,
			&acquisitionValue,
			&openingValue,
			&depreciationAmount,
			&accumulatedValue,
			&closingValue,
			&item.Status,
			&postedAt,
		); err != nil {
			return result, err
		}

		item.PeriodDate = periodDate.Format("02 Jan 2006")
		item.AcquisitionValueDisplay = formatAssetAmountID(acquisitionValue)
		item.OpeningBookValueDisplay = formatAssetAmountID(openingValue)
		item.DepreciationAmountDisplay = formatAssetAmountID(depreciationAmount)
		item.AccumulatedDepreciationDisplay = formatAssetAmountID(accumulatedValue)
		item.ClosingBookValueDisplay = formatAssetAmountID(closingValue)
		if postedAt.Valid {
			item.PostedAtDisplay = postedAt.Time.Format("02 Jan 2006 15:04")
		}
		result.Items = append(result.Items, item)
	}

	return result, rows.Err()
}

func (r *AssetDepreciationRepository) loadMonthlyStats(year, month int, stats *models.MonthlyDepreciationStats) error {
	var totalAmount, draftAmount, postedAmount float64
	err := r.DB.QueryRow(`
		SELECT
			COUNT(*),
			COALESCE(SUM(CASE WHEN status = 'DRAFT' THEN 1 ELSE 0 END), 0),
			COALESCE(SUM(CASE WHEN status = 'POSTED' THEN 1 ELSE 0 END), 0),
			COALESCE(SUM(CASE WHEN status = 'SKIPPED' THEN 1 ELSE 0 END), 0),
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

func (r *AssetDepreciationRepository) GenerateAllSchedules() error {
	_, err := r.DB.Exec(`CALL sp_generate_all_asset_depreciation_schedules()`)
	return err
}

func (r *AssetDepreciationRepository) PostSchedules(ids []int64) (int64, error) {
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
			return 0, fmt.Errorf("depresiasi periode %s belum dapat diposting", periodDate.Format("January 2006"))
		}
	}
	if err := rows.Close(); err != nil {
		return 0, err
	}
	if found != len(ids) {
		return 0, errors.New("sebagian data depresiasi tidak ditemukan")
	}

	updateQuery := `
		UPDATE asset_depreciation_schedules
		SET status = 'POSTED', posted_at = CURRENT_TIMESTAMP, updated_at = CURRENT_TIMESTAMP
		WHERE status = 'DRAFT' AND id IN (` + strings.Join(placeholders, ",") + `)`
	result, err := tx.Exec(updateQuery, args...)
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
	if err := tx.Commit(); err != nil {
		return 0, err
	}
	return affected, nil
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
