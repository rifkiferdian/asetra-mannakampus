package repositories

import (
	"database/sql"
	"gobase-app/models"
	"math"
	"strings"
	"time"
)

type AssetDisposalReportRepository struct {
	DB *sql.DB
}

func (r *AssetDisposalReportRepository) GetReport(filter models.AssetDisposalReportFilter) (models.AssetDisposalReportResult, error) {
	result := models.AssetDisposalReportResult{}
	where, args := assetDisposalReportWhere(filter)

	if err := r.DB.QueryRow(`
		SELECT COUNT(*),
			COALESCE(SUM(disposal.status = 'POSTED'), 0),
			COALESCE(SUM(disposal.status = 'REVERSED'), 0),
			COALESCE(SUM(CASE WHEN disposal.status = 'POSTED' THEN disposal.acquisition_value ELSE 0 END), 0),
			COALESCE(SUM(CASE WHEN disposal.status = 'POSTED' THEN disposal.accumulated_depreciation ELSE 0 END), 0),
			COALESCE(SUM(CASE WHEN disposal.status = 'POSTED' THEN disposal.book_value ELSE 0 END), 0),
			COALESCE(SUM(CASE WHEN disposal.status = 'POSTED' THEN disposal.disposal_value ELSE 0 END), 0),
			COALESCE(SUM(CASE WHEN disposal.status = 'POSTED' AND disposal.gain_loss_amount > 0 THEN disposal.gain_loss_amount ELSE 0 END), 0),
			COALESCE(SUM(CASE WHEN disposal.status = 'POSTED' AND disposal.gain_loss_amount < 0 THEN ABS(disposal.gain_loss_amount) ELSE 0 END), 0),
			COALESCE(SUM(CASE WHEN disposal.status = 'POSTED' THEN disposal.gain_loss_amount ELSE 0 END), 0)
		FROM asset_disposals disposal
		JOIN assets asset ON asset.id = disposal.asset_id
		LEFT JOIN asset_types asset_type ON asset_type.id = asset.asset_type_id
		LEFT JOIN stores store ON store.store_id = asset.store_id
		JOIN asset_disposal_types disposal_type ON disposal_type.id = disposal.disposal_type_id
		WHERE `+where, args...).Scan(
		&result.TotalRows, &result.Summary.PostedCount, &result.Summary.ReversedCount,
		&result.Summary.AcquisitionValue, &result.Summary.AccumulatedDepreciation,
		&result.Summary.BookValue, &result.Summary.DisposalValue,
		&result.Summary.ProfitAmount, &result.Summary.LossAmount, &result.Summary.NetGainLoss,
	); err != nil {
		return result, err
	}
	result.Summary.TransactionCount = result.TotalRows
	formatAssetDisposalReportSummary(&result.Summary)

	result.TotalPages = 1
	if result.TotalRows > 0 {
		result.TotalPages = (result.TotalRows + filter.PerPage - 1) / filter.PerPage
	}
	if filter.Page > result.TotalPages {
		filter.Page = result.TotalPages
	}

	queryArgs := append(append([]any{}, args...), filter.PerPage, (filter.Page-1)*filter.PerPage)
	items, err := r.queryReportRows(where, queryArgs, true)
	if err != nil {
		return result, err
	}
	result.Items = items
	return result, nil
}

func (r *AssetDisposalReportRepository) GetExportRows(filter models.AssetDisposalReportFilter) ([]models.AssetDisposalReportRow, error) {
	where, args := assetDisposalReportWhere(filter)
	return r.queryReportRows(where, args, false)
}

func (r *AssetDisposalReportRepository) queryReportRows(where string, args []any, paginated bool) ([]models.AssetDisposalReportRow, error) {
	query := `
		SELECT disposal.id, disposal.disposal_number, disposal.disposal_date,
			asset.id, asset.asset_code, asset.asset_name, COALESCE(asset_type.name, '-'),
			COALESCE(store.store_name, 'Head Office'), disposal_type.name,
			COALESCE(disposal.buyer_name, ''), COALESCE(disposal.document_reference, ''),
			disposal.acquisition_value, disposal.accumulated_depreciation, disposal.book_value,
			disposal.disposal_value, disposal.gain_loss_amount, disposal.status,
			disposal.posted_at, COALESCE(poster.name, ''), COALESCE(disposal.reversal_reason, '')
		FROM asset_disposals disposal
		JOIN assets asset ON asset.id = disposal.asset_id
		LEFT JOIN asset_types asset_type ON asset_type.id = asset.asset_type_id
		LEFT JOIN stores store ON store.store_id = asset.store_id
		JOIN asset_disposal_types disposal_type ON disposal_type.id = disposal.disposal_type_id
		LEFT JOIN users poster ON poster.id = disposal.posted_by
		WHERE ` + where + `
		ORDER BY disposal.disposal_date DESC, disposal.id DESC`
	if paginated {
		query += ` LIMIT ? OFFSET ?`
	}
	rows, err := r.DB.Query(query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	items := make([]models.AssetDisposalReportRow, 0)
	for rows.Next() {
		var item models.AssetDisposalReportRow
		var disposalDate time.Time
		var postedAt sql.NullTime
		if err := rows.Scan(
			&item.ID, &item.DisposalNumber, &disposalDate,
			&item.AssetID, &item.AssetCode, &item.AssetName, &item.AssetTypeName,
			&item.StoreName, &item.DisposalTypeName, &item.BuyerName, &item.DocumentReference,
			&item.AcquisitionValue, &item.AccumulatedDepreciation, &item.BookValue,
			&item.DisposalValue, &item.GainLossAmount, &item.Status,
			&postedAt, &item.PostedByName, &item.ReversalReason,
		); err != nil {
			return nil, err
		}
		item.DisposalDate = disposalDate.Format("2006-01-02")
		item.DisposalDateDisplay = formatDepreciationDateID(disposalDate, false)
		item.AcquisitionValueDisplay = formatAssetAmountID(item.AcquisitionValue)
		item.AccumulatedDepreciationDisplay = formatAssetAmountID(item.AccumulatedDepreciation)
		item.BookValueDisplay = formatAssetAmountID(item.BookValue)
		item.DisposalValueDisplay = formatAssetAmountID(item.DisposalValue)
		item.GainLossAmountDisplay = formatAssetAmountID(math.Abs(item.GainLossAmount))
		item.GainLossLabel = assetDisposalGainLossLabel(item.GainLossAmount)
		if postedAt.Valid {
			item.PostedAtDisplay = formatDepreciationDateID(postedAt.Time, true)
		}
		items = append(items, item)
	}
	return items, rows.Err()
}

func assetDisposalReportWhere(filter models.AssetDisposalReportFilter) (string, []any) {
	clauses := []string{"disposal.status IN ('POSTED','REVERSED')", "disposal.disposal_date BETWEEN ? AND ?"}
	args := []any{filter.DateFrom, filter.DateTo}
	if filter.DisposalTypeID > 0 {
		clauses = append(clauses, "disposal.disposal_type_id = ?")
		args = append(args, filter.DisposalTypeID)
	}
	if filter.AssetTypeID > 0 {
		clauses = append(clauses, "asset.asset_type_id = ?")
		args = append(args, filter.AssetTypeID)
	}
	if filter.StoreID > 0 {
		clauses = append(clauses, "asset.store_id = ?")
		args = append(args, filter.StoreID)
	}
	if filter.Status != "ALL" {
		clauses = append(clauses, "disposal.status = ?")
		args = append(args, filter.Status)
	}
	switch filter.Result {
	case "PROFIT":
		clauses = append(clauses, "disposal.gain_loss_amount > 0")
	case "LOSS":
		clauses = append(clauses, "disposal.gain_loss_amount < 0")
	case "BREAK_EVEN":
		clauses = append(clauses, "ABS(disposal.gain_loss_amount) <= 0.005")
	}
	if filter.Search != "" {
		term := "%" + filter.Search + "%"
		clauses = append(clauses, `(disposal.disposal_number LIKE ? OR asset.asset_code LIKE ?
			OR asset.asset_name LIKE ? OR COALESCE(disposal.buyer_name, '') LIKE ?
			OR COALESCE(disposal.document_reference, '') LIKE ?)`)
		args = append(args, term, term, term, term, term)
	}
	return strings.Join(clauses, " AND "), args
}

func formatAssetDisposalReportSummary(summary *models.AssetDisposalReportSummary) {
	summary.AcquisitionValueDisplay = formatAssetAmountID(summary.AcquisitionValue)
	summary.AccumulatedDepreciationDisplay = formatAssetAmountID(summary.AccumulatedDepreciation)
	summary.BookValueDisplay = formatAssetAmountID(summary.BookValue)
	summary.DisposalValueDisplay = formatAssetAmountID(summary.DisposalValue)
	summary.ProfitAmountDisplay = formatAssetAmountID(summary.ProfitAmount)
	summary.LossAmountDisplay = formatAssetAmountID(summary.LossAmount)
	summary.NetGainLossDisplay = formatAssetAmountID(math.Abs(summary.NetGainLoss))
	summary.NetGainLossLabel = assetDisposalGainLossLabel(summary.NetGainLoss)
}

func assetDisposalGainLossLabel(value float64) string {
	if value > 0.005 {
		return "LABA"
	}
	if value < -0.005 {
		return "RUGI"
	}
	return "IMPAS"
}
