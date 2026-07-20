package controllers

import (
	"encoding/csv"
	"fmt"
	"gobase-app/config"
	"gobase-app/models"
	"gobase-app/repositories"
	"gobase-app/services"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
)

func AssetDisposalReportIndex(c *gin.Context) {
	filter := assetDisposalReportFilter(c)
	result, err := assetDisposalReportService().GetReport(filter)
	if err != nil {
		c.String(http.StatusBadRequest, err.Error())
		return
	}
	types, err := assetDisposalService().GetDisposalTypes()
	if err != nil {
		c.String(http.StatusInternalServerError, err.Error())
		return
	}
	assetTypes, err := assetService().GetAssetTypes()
	if err != nil {
		c.String(http.StatusInternalServerError, err.Error())
		return
	}
	stores, err := (&repositories.StoreRepository{DB: config.DB}).GetAll()
	if err != nil {
		c.String(http.StatusInternalServerError, err.Error())
		return
	}

	Render(c, "asset_disposal_report.html", gin.H{
		"Title": "Laporan Disposal Aset", "Page": "asset_disposal_report",
		"Items": result.Items, "Summary": result.Summary,
		"Filter": filter, "FilterQuery": assetDisposalReportQuery(filter),
		"Types": types, "AssetTypes": assetTypes, "Stores": stores,
		"Pagination": assetDisposalReportPagination(filter, result),
	})
}

func AssetDisposalReportExportCSV(c *gin.Context) {
	filter := assetDisposalReportFilter(c)
	rows, err := assetDisposalReportService().GetExportRows(filter)
	if err != nil {
		c.String(http.StatusBadRequest, err.Error())
		return
	}

	filename := fmt.Sprintf("laporan-disposal-%s-%s.csv", filter.DateFrom, filter.DateTo)
	c.Header("Content-Type", "text/csv; charset=utf-8")
	c.Header("Content-Disposition", fmt.Sprintf(`attachment; filename="%s"`, filename))
	c.Status(http.StatusOK)
	_, _ = c.Writer.Write([]byte{0xEF, 0xBB, 0xBF})
	w := csv.NewWriter(c.Writer)
	_ = w.Write([]string{
		"Nomor Disposal", "Tanggal Disposal", "Kode Aset", "Nama Aset", "Tipe Aset", "Store",
		"Jenis Disposal", "Pembeli", "Referensi Dokumen", "Nilai Perolehan",
		"Akumulasi Depresiasi", "Nilai Buku", "Nilai Disposal", "Hasil", "Laba/Rugi",
		"Status", "Tanggal Posting", "Diposting Oleh", "Alasan Reversal",
	})
	for _, item := range rows {
		_ = w.Write([]string{
			assetDisposalReportCSVText(item.DisposalNumber), item.DisposalDate,
			assetDisposalReportCSVText(item.AssetCode), assetDisposalReportCSVText(item.AssetName),
			assetDisposalReportCSVText(item.AssetTypeName), assetDisposalReportCSVText(item.StoreName),
			assetDisposalReportCSVText(item.DisposalTypeName), assetDisposalReportCSVText(item.BuyerName),
			assetDisposalReportCSVText(item.DocumentReference), reportFloat(item.AcquisitionValue),
			reportFloat(item.AccumulatedDepreciation), reportFloat(item.BookValue), reportFloat(item.DisposalValue),
			item.GainLossLabel, reportFloat(item.GainLossAmount), item.Status,
			item.PostedAtDisplay, assetDisposalReportCSVText(item.PostedByName),
			assetDisposalReportCSVText(item.ReversalReason),
		})
	}
	w.Flush()
	if err := w.Error(); err != nil {
		return
	}
}

func assetDisposalReportFilter(c *gin.Context) models.AssetDisposalReportFilter {
	now := time.Now()
	dateFrom := strings.TrimSpace(c.Query("date_from"))
	if dateFrom == "" {
		dateFrom = time.Date(now.Year(), 1, 1, 0, 0, 0, 0, now.Location()).Format("2006-01-02")
	}
	dateTo := strings.TrimSpace(c.Query("date_to"))
	if dateTo == "" {
		dateTo = now.Format("2006-01-02")
	}
	page, _ := strconv.Atoi(c.Query("page"))
	if page < 1 {
		page = 1
	}
	status := strings.ToUpper(strings.TrimSpace(c.Query("status")))
	if status == "" {
		status = "POSTED"
	}
	result := strings.ToUpper(strings.TrimSpace(c.Query("result")))
	if result == "" {
		result = "ALL"
	}
	return models.AssetDisposalReportFilter{
		DateFrom: dateFrom, DateTo: dateTo,
		DisposalTypeID: queryInt64(c, "disposal_type_id"), AssetTypeID: queryInt64(c, "asset_type_id"),
		StoreID: int(queryInt64(c, "store_id")), Result: result, Status: status,
		Search: strings.TrimSpace(c.Query("search")), Page: page, PerPage: 50,
	}
}

func assetDisposalReportQuery(filter models.AssetDisposalReportFilter) string {
	values := url.Values{}
	values.Set("date_from", filter.DateFrom)
	values.Set("date_to", filter.DateTo)
	values.Set("disposal_type_id", strconv.FormatInt(filter.DisposalTypeID, 10))
	values.Set("asset_type_id", strconv.FormatInt(filter.AssetTypeID, 10))
	values.Set("store_id", strconv.Itoa(filter.StoreID))
	values.Set("result", filter.Result)
	values.Set("status", filter.Status)
	values.Set("search", filter.Search)
	return values.Encode()
}

func assetDisposalReportPagination(filter models.AssetDisposalReportFilter, result models.AssetDisposalReportResult) assetPaginationMeta {
	page := filter.Page
	if page > result.TotalPages {
		page = result.TotalPages
	}
	start, end := 0, 0
	if result.TotalRows > 0 {
		start = (page-1)*filter.PerPage + 1
		end = start + len(result.Items) - 1
	}
	return assetPaginationMeta{
		CurrentPage: page, PrevPage: page - 1, NextPage: page + 1, TotalPages: result.TotalPages,
		PageSize: filter.PerPage, PageStart: start, PageEnd: end, TotalRows: result.TotalRows,
		HasPrev: page > 1, HasNext: page < result.TotalPages,
	}
}

func assetDisposalReportService() *services.AssetDisposalReportService {
	return &services.AssetDisposalReportService{Repo: &repositories.AssetDisposalReportRepository{DB: config.DB}}
}

func queryInt64(c *gin.Context, key string) int64 {
	value, _ := strconv.ParseInt(strings.TrimSpace(c.Query(key)), 10, 64)
	return value
}

func assetDisposalReportCSVText(value string) string {
	value = strings.TrimSpace(value)
	if strings.HasPrefix(value, "=") || strings.HasPrefix(value, "+") ||
		strings.HasPrefix(value, "-") || strings.HasPrefix(value, "@") {
		return "'" + value
	}
	return value
}

func reportFloat(value float64) string {
	return strconv.FormatFloat(value, 'f', 2, 64)
}
