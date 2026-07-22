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

func AssetDepreciationAnnualReportIndex(c *gin.Context) {
	filter := annualDepreciationReportFilter(c)
	result, err := annualDepreciationReportService().GetReport(filter)
	if err != nil {
		c.String(http.StatusBadRequest, err.Error())
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
	locations, err := assetService().GetLocations()
	if err != nil {
		c.String(http.StatusInternalServerError, err.Error())
		return
	}
	Render(c, "asset_depreciation_annual_report.html", gin.H{
		"Title": "Laporan Penyusutan Tahunan", "Page": "asset_depreciation_annual_report",
		"Report": result, "Filter": filter, "FilterQuery": annualDepreciationReportQuery(filter),
		"AssetTypes": assetTypes, "Stores": stores, "Locations": locations,
	})
}

func AssetDepreciationAnnualReportExportCSV(c *gin.Context) {
	filter := annualDepreciationReportFilter(c)
	result, err := annualDepreciationReportService().GetReport(filter)
	if err != nil {
		c.String(http.StatusBadRequest, err.Error())
		return
	}
	filename := fmt.Sprintf("laporan-penyusutan-%d-%d.csv", filter.YearFrom, filter.YearTo)
	c.Header("Content-Type", "text/csv; charset=utf-8")
	c.Header("Content-Disposition", fmt.Sprintf(`attachment; filename="%s"`, filename))
	c.Status(http.StatusOK)
	_, _ = c.Writer.Write([]byte{0xEF, 0xBB, 0xBF})
	w := csv.NewWriter(c.Writer)
	header := []string{"No", "Jenis Aktiva", "Kode Aset", "Tanggal", "Tahun Perolehan", "Harga Perolehan"}
	for _, year := range result.Years {
		header = append(header,
			fmt.Sprintf("Penyusutan %d", year.Year),
			fmt.Sprintf("Ak. Penyusutan s/d 31/12/%d", year.Year),
			fmt.Sprintf("Nilai Buku s/d 31/12/%d", year.Year),
		)
	}
	_ = w.Write(header)
	for _, group := range result.Groups {
		for _, item := range group.Rows {
			record := []string{
				strconv.Itoa(item.Sequence), assetDisposalReportCSVText(item.AssetName), assetDisposalReportCSVText(item.AssetCode),
				item.AcquisitionDate, strconv.Itoa(item.AcquisitionYear), reportFloat(item.AcquisitionValue),
			}
			for _, amount := range item.YearAmounts {
				record = append(record, reportFloat(amount.Depreciation), reportFloat(amount.AccumulatedDepreciation), reportFloat(amount.BookValue))
			}
			_ = w.Write(record)
		}
		subtotal := []string{"", "Sub Jumlah " + assetDisposalReportCSVText(group.AssetTypeName), "", "", "", reportFloat(group.AcquisitionValue)}
		for _, amount := range group.YearTotals {
			subtotal = append(subtotal, reportFloat(amount.Depreciation), reportFloat(amount.AccumulatedDepreciation), reportFloat(amount.BookValue))
		}
		_ = w.Write(subtotal)
	}
	grandTotal := []string{"", "GRAND TOTAL", "", "", "", reportFloat(result.AcquisitionValue)}
	for _, amount := range result.YearTotals {
		grandTotal = append(grandTotal, reportFloat(amount.Depreciation), reportFloat(amount.AccumulatedDepreciation), reportFloat(amount.BookValue))
	}
	_ = w.Write(grandTotal)
	w.Flush()
}

func annualDepreciationReportFilter(c *gin.Context) models.AnnualDepreciationReportFilter {
	currentYear := time.Now().Year()
	yearFrom, _ := strconv.Atoi(strings.TrimSpace(c.Query("year_from")))
	if yearFrom == 0 {
		yearFrom = currentYear - 2
	}
	yearTo, _ := strconv.Atoi(strings.TrimSpace(c.Query("year_to")))
	if yearTo == 0 {
		yearTo = currentYear
	}
	mode := strings.ToUpper(strings.TrimSpace(c.Query("mode")))
	if mode == "" {
		mode = "ACTUAL"
	}
	status := strings.ToUpper(strings.TrimSpace(c.Query("asset_status")))
	if status == "" {
		status = "ALL"
	}
	return models.AnnualDepreciationReportFilter{
		YearFrom: yearFrom, YearTo: yearTo, Mode: mode,
		AssetTypeID: queryInt64(c, "asset_type_id"), StoreID: int(queryInt64(c, "store_id")),
		LocationID: queryInt64(c, "location_id"), AssetStatus: status,
		Search: strings.TrimSpace(c.Query("search")),
	}
}

func annualDepreciationReportQuery(filter models.AnnualDepreciationReportFilter) string {
	values := url.Values{}
	values.Set("year_from", strconv.Itoa(filter.YearFrom))
	values.Set("year_to", strconv.Itoa(filter.YearTo))
	values.Set("mode", filter.Mode)
	values.Set("asset_type_id", strconv.FormatInt(filter.AssetTypeID, 10))
	values.Set("store_id", strconv.Itoa(filter.StoreID))
	values.Set("location_id", strconv.FormatInt(filter.LocationID, 10))
	values.Set("asset_status", filter.AssetStatus)
	values.Set("search", filter.Search)
	return values.Encode()
}

func annualDepreciationReportService() *services.AssetDepreciationAnnualReportService {
	return &services.AssetDepreciationAnnualReportService{Repo: &repositories.AssetDepreciationAnnualReportRepository{DB: config.DB}}
}
