package repositories

import (
	"database/sql"
	"gobase-app/models"
	"math"
	"strconv"
	"strings"
	"time"
)

type AssetDepreciationAnnualReportRepository struct {
	DB *sql.DB
}

func (r *AssetDepreciationAnnualReportRepository) GetReport(filter models.AnnualDepreciationReportFilter) (models.AnnualDepreciationReportResult, error) {
	result := models.AnnualDepreciationReportResult{Years: annualDepreciationYears(filter.YearFrom, filter.YearTo)}
	where, args := annualDepreciationAssetWhere(filter)
	rows, err := r.DB.Query(`
		SELECT asset.id, asset.asset_code, asset.asset_name,
			asset_type.id, asset_type.code, asset_type.name,
			asset.acquisition_date, asset.acquisition_value,
			COALESCE(store.store_name, location_store.store_name, 'Head Office'),
			COALESCE(location.location_name, '-'), asset.status, profile.status,
			profile.depreciable_basis, profile.salvage_value
		FROM asset_depreciation_profiles profile
		JOIN assets asset ON asset.id = profile.asset_id
		JOIN asset_types asset_type ON asset_type.id = asset.asset_type_id
		LEFT JOIN stores store ON store.store_id = asset.store_id
		LEFT JOIN asset_locations location ON location.id = asset.location_id
		LEFT JOIN stores location_store ON location_store.store_id = location.store_id
		WHERE `+where+`
		ORDER BY asset_type.name, asset.asset_name, asset.asset_code
	`, args...)
	if err != nil {
		return result, err
	}
	defer rows.Close()

	items := make([]models.AnnualDepreciationRow, 0)
	for rows.Next() {
		var item models.AnnualDepreciationRow
		var acquisitionDate time.Time
		if err := rows.Scan(
			&item.AssetID, &item.AssetCode, &item.AssetName,
			&item.AssetTypeID, &item.AssetTypeCode, &item.AssetTypeName,
			&acquisitionDate, &item.AcquisitionValue, &item.StoreName,
			&item.LocationName, &item.AssetStatus, &item.ProfileStatus,
			&item.DepreciableBasis, &item.SalvageValue,
		); err != nil {
			return result, err
		}
		item.AcquisitionDate = acquisitionDate.Format("2006-01-02")
		item.AcquisitionDateDisplay = formatDepreciationDateID(acquisitionDate, false)
		item.AcquisitionYear = acquisitionDate.Year()
		item.AcquisitionValueDisplay = formatAssetAmountID(item.AcquisitionValue)
		items = append(items, item)
	}
	if err := rows.Err(); err != nil {
		return result, err
	}

	yearly, err := r.getAnnualDepreciationAmounts(filter)
	if err != nil {
		return result, err
	}
	return buildAnnualDepreciationReport(result, items, yearly), nil
}

func (r *AssetDepreciationAnnualReportRepository) getAnnualDepreciationAmounts(filter models.AnnualDepreciationReportFilter) (map[int64]map[int]float64, error) {
	where, args := annualDepreciationAssetWhere(filter)
	statuses := "'POSTED'"
	if filter.Mode == "PROJECTION" {
		statuses = "'POSTED','DRAFT'"
	}
	args = append(args, filter.YearTo)
	rows, err := r.DB.Query(`
		SELECT asset.id, schedule.period_year, SUM(schedule.depreciation_amount)
		FROM asset_depreciation_profiles profile
		JOIN assets asset ON asset.id = profile.asset_id
		JOIN asset_types asset_type ON asset_type.id = asset.asset_type_id
		LEFT JOIN stores store ON store.store_id = asset.store_id
		LEFT JOIN asset_locations location ON location.id = asset.location_id
		LEFT JOIN stores location_store ON location_store.store_id = location.store_id
		JOIN asset_depreciation_schedules schedule ON schedule.profile_id = profile.id
		WHERE `+where+`
		  AND schedule.period_year <= ?
		  AND schedule.status IN (`+statuses+`)
		  AND schedule.version_no = (
			SELECT MAX(latest.version_no)
			FROM asset_depreciation_schedules latest
			WHERE latest.asset_id = schedule.asset_id
			  AND latest.period_year = schedule.period_year
			  AND latest.period_month = schedule.period_month
			  AND latest.status IN (`+statuses+`)
		  )
		GROUP BY asset.id, schedule.period_year
		ORDER BY asset.id, schedule.period_year
	`, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	result := make(map[int64]map[int]float64)
	for rows.Next() {
		var assetID int64
		var year int
		var amount float64
		if err := rows.Scan(&assetID, &year, &amount); err != nil {
			return nil, err
		}
		if result[assetID] == nil {
			result[assetID] = make(map[int]float64)
		}
		result[assetID][year] = amount
	}
	return result, rows.Err()
}

func annualDepreciationAssetWhere(filter models.AnnualDepreciationReportFilter) (string, []any) {
	clauses := []string{"asset.acquisition_date IS NOT NULL", "asset.acquisition_date <= ?"}
	args := []any{time.Date(filter.YearTo, 12, 31, 0, 0, 0, 0, time.Local)}
	if filter.AssetTypeID > 0 {
		clauses = append(clauses, "asset.asset_type_id = ?")
		args = append(args, filter.AssetTypeID)
	}
	if filter.StoreID > 0 {
		clauses = append(clauses, "COALESCE(asset.store_id, location.store_id) = ?")
		args = append(args, filter.StoreID)
	}
	if filter.LocationID > 0 {
		clauses = append(clauses, "asset.location_id = ?")
		args = append(args, filter.LocationID)
	}
	if filter.AssetStatus != "ALL" {
		clauses = append(clauses, "asset.status = ?")
		args = append(args, filter.AssetStatus)
	}
	if filter.Search != "" {
		term := "%" + filter.Search + "%"
		clauses = append(clauses, "(asset.asset_code LIKE ? OR asset.asset_name LIKE ? OR COALESCE(asset.serial_number, '') LIKE ?)")
		args = append(args, term, term, term)
	}
	return strings.Join(clauses, " AND "), args
}

func buildAnnualDepreciationReport(result models.AnnualDepreciationReportResult, items []models.AnnualDepreciationRow, yearly map[int64]map[int]float64) models.AnnualDepreciationReportResult {
	groupIndexes := make(map[int64]int)
	result.YearTotals = emptyAnnualDepreciationAmounts(result.Years)
	for _, item := range items {
		item.Sequence = result.AssetCount + 1
		item.YearAmounts = annualAmountsForAsset(item, result.Years, yearly[item.AssetID])
		index, exists := groupIndexes[item.AssetTypeID]
		if !exists {
			index = len(result.Groups)
			groupIndexes[item.AssetTypeID] = index
			result.Groups = append(result.Groups, models.AnnualDepreciationGroup{
				AssetTypeID: item.AssetTypeID, AssetTypeCode: item.AssetTypeCode,
				AssetTypeName: item.AssetTypeName, Rows: make([]models.AnnualDepreciationRow, 0),
				YearTotals: emptyAnnualDepreciationAmounts(result.Years),
			})
		}
		group := &result.Groups[index]
		group.Rows = append(group.Rows, item)
		group.AssetCount++
		group.AcquisitionValue += item.AcquisitionValue
		result.AssetCount++
		result.AcquisitionValue += item.AcquisitionValue
		for yearIndex, amount := range item.YearAmounts {
			addAnnualDepreciationAmount(&group.YearTotals[yearIndex], amount)
			addAnnualDepreciationAmount(&result.YearTotals[yearIndex], amount)
		}
	}
	for index := range result.Groups {
		result.Groups[index].AcquisitionValueDisplay = formatAssetAmountID(result.Groups[index].AcquisitionValue)
		formatAnnualDepreciationAmounts(result.Groups[index].YearTotals)
	}
	result.AcquisitionValueDisplay = formatAssetAmountID(result.AcquisitionValue)
	formatAnnualDepreciationAmounts(result.YearTotals)
	if len(result.YearTotals) > 0 {
		latest := result.YearTotals[len(result.YearTotals)-1]
		result.LatestYear = latest.Year
		result.LatestDepreciationDisplay = latest.DepreciationDisplay
		result.LatestAccumulatedDisplay = latest.AccumulatedDepreciationDisplay
		result.LatestBookValueDisplay = latest.BookValueDisplay
	} else {
		result.LatestDepreciationDisplay = formatAssetAmountID(0)
		result.LatestAccumulatedDisplay = formatAssetAmountID(0)
		result.LatestBookValueDisplay = formatAssetAmountID(0)
	}
	return result
}

func annualAmountsForAsset(item models.AnnualDepreciationRow, years []models.AnnualDepreciationYear, yearly map[int]float64) []models.AnnualDepreciationAmount {
	result := make([]models.AnnualDepreciationAmount, 0, len(years))
	for _, year := range years {
		accumulated := 0.0
		for amountYear, amount := range yearly {
			if amountYear <= year.Year {
				accumulated += amount
			}
		}
		bookValue := math.Max(item.SalvageValue, item.DepreciableBasis-accumulated)
		if accumulated <= 0 && year.Year < item.AcquisitionYear {
			bookValue = 0
		}
		result = append(result, models.AnnualDepreciationAmount{
			Year: year.Year, Depreciation: yearly[year.Year], AccumulatedDepreciation: accumulated, BookValue: bookValue,
		})
	}
	formatAnnualDepreciationAmounts(result)
	return result
}

func annualDepreciationYears(from, to int) []models.AnnualDepreciationYear {
	result := make([]models.AnnualDepreciationYear, 0, to-from+1)
	for year := from; year <= to; year++ {
		result = append(result, models.AnnualDepreciationYear{Year: year, AsOfLabel: "S/D 31/12/" + strconv.Itoa(year)})
	}
	return result
}

func emptyAnnualDepreciationAmounts(years []models.AnnualDepreciationYear) []models.AnnualDepreciationAmount {
	result := make([]models.AnnualDepreciationAmount, len(years))
	for index, year := range years {
		result[index].Year = year.Year
	}
	return result
}

func addAnnualDepreciationAmount(target *models.AnnualDepreciationAmount, amount models.AnnualDepreciationAmount) {
	target.Depreciation += amount.Depreciation
	target.AccumulatedDepreciation += amount.AccumulatedDepreciation
	target.BookValue += amount.BookValue
}

func formatAnnualDepreciationAmounts(items []models.AnnualDepreciationAmount) {
	for index := range items {
		items[index].DepreciationDisplay = formatAssetAmountID(items[index].Depreciation)
		items[index].AccumulatedDepreciationDisplay = formatAssetAmountID(items[index].AccumulatedDepreciation)
		items[index].BookValueDisplay = formatAssetAmountID(items[index].BookValue)
	}
}
