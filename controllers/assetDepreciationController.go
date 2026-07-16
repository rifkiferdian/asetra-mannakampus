package controllers

import (
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

const monthlyDepreciationPageSize = 50

func MonthlyDepreciationIndex(c *gin.Context) {
	filter := monthlyDepreciationFilter(c)
	periodLabel := time.Date(filter.Year, time.Month(filter.Month), 1, 0, 0, 0, 0, time.Local).Format("January 2006")
	result, err := assetDepreciationService().GetMonthlyDepreciation(filter)
	if err != nil {
		Render(c, "monthly_depreciation.html", gin.H{
			"Title": "Monthly Depreciation", "Page": "monthly_depreciation",
			"Filter": filter, "Error": depreciationErrorMessage(err),
			"Months": depreciationMonthOptions(), "Years": depreciationYearOptions(filter.Year),
			"PeriodLabel": periodLabel,
			"Stats":       models.MonthlyDepreciationStats{},
			"Pagination": models.DepreciationPagination{
				CurrentPage: 1, TotalPages: 1, PageSize: monthlyDepreciationPageSize,
			},
		})
		return
	}

	Render(c, "monthly_depreciation.html", gin.H{
		"Title":       "Monthly Depreciation",
		"Page":        "monthly_depreciation",
		"Items":       result.Items,
		"Stats":       result.Stats,
		"Filter":      filter,
		"Months":      depreciationMonthOptions(),
		"Years":       depreciationYearOptions(filter.Year),
		"PeriodLabel": periodLabel,
		"Pagination":  monthlyDepreciationPagination(filter, result),
		"Success":     strings.TrimSpace(c.Query("success")),
		"Error":       strings.TrimSpace(c.Query("error")),
	})
}

func MonthlyDepreciationGenerate(c *gin.Context) {
	filter := monthlyDepreciationFilter(c)
	if err := assetDepreciationService().GenerateSchedules(); err != nil {
		redirectMonthlyDepreciation(c, filter, "", depreciationErrorMessage(err))
		return
	}
	redirectMonthlyDepreciation(c, filter, "Jadwal depresiasi berhasil diperbarui", "")
}

func MonthlyDepreciationPost(c *gin.Context) {
	filter := monthlyDepreciationFilter(c)
	ids := make([]int64, 0)
	for _, value := range c.PostFormArray("schedule_ids") {
		id, err := strconv.ParseInt(strings.TrimSpace(value), 10, 64)
		if err == nil && id > 0 {
			ids = append(ids, id)
		}
	}

	count, err := assetDepreciationService().PostSchedules(ids)
	if err != nil {
		redirectMonthlyDepreciation(c, filter, "", depreciationErrorMessage(err))
		return
	}
	redirectMonthlyDepreciation(c, filter, strconv.FormatInt(count, 10)+" depresiasi berhasil diposting", "")
}

func assetDepreciationService() *services.AssetDepreciationService {
	return &services.AssetDepreciationService{Repo: &repositories.AssetDepreciationRepository{DB: config.DB}}
}

func monthlyDepreciationFilter(c *gin.Context) models.MonthlyDepreciationFilter {
	now := time.Now()
	year, _ := strconv.Atoi(firstNonEmpty(c.PostForm("year"), c.Query("year")))
	month, _ := strconv.Atoi(firstNonEmpty(c.PostForm("month"), c.Query("month")))
	page, _ := strconv.Atoi(firstNonEmpty(c.PostForm("page"), c.Query("page")))
	if year == 0 {
		year = now.Year()
	}
	if month == 0 {
		month = int(now.Month())
	}
	if page < 1 {
		page = 1
	}
	status := strings.ToUpper(strings.TrimSpace(firstNonEmpty(c.PostForm("status"), c.Query("status"))))
	if status == "" {
		status = "ALL"
	}
	return models.MonthlyDepreciationFilter{
		Year: year, Month: month, Status: status,
		Search: strings.TrimSpace(firstNonEmpty(c.PostForm("search"), c.Query("search"))),
		Page:   page, PerPage: monthlyDepreciationPageSize,
	}
}

func monthlyDepreciationPagination(filter models.MonthlyDepreciationFilter, result models.MonthlyDepreciationResult) models.DepreciationPagination {
	page := filter.Page
	if page > result.TotalPages {
		page = result.TotalPages
	}
	start, end := 0, 0
	if result.TotalRows > 0 {
		start = (page-1)*filter.PerPage + 1
		end = start + len(result.Items) - 1
	}
	return models.DepreciationPagination{
		CurrentPage: page, TotalPages: result.TotalPages, PageStart: start, PageEnd: end,
		TotalRows: result.TotalRows, PageSize: filter.PerPage, HasPrev: page > 1, HasNext: page < result.TotalPages,
		PrevURL: monthlyDepreciationURL(filter, page-1), NextURL: monthlyDepreciationURL(filter, page+1),
	}
}

func monthlyDepreciationURL(filter models.MonthlyDepreciationFilter, page int) string {
	values := url.Values{}
	values.Set("year", strconv.Itoa(filter.Year))
	values.Set("month", strconv.Itoa(filter.Month))
	values.Set("status", filter.Status)
	if filter.Search != "" {
		values.Set("search", filter.Search)
	}
	values.Set("page", strconv.Itoa(page))
	return "/asset-depreciation/monthly?" + values.Encode()
}

func redirectMonthlyDepreciation(c *gin.Context, filter models.MonthlyDepreciationFilter, success, errorMessage string) {
	values := url.Values{}
	values.Set("year", strconv.Itoa(filter.Year))
	values.Set("month", strconv.Itoa(filter.Month))
	values.Set("status", filter.Status)
	if filter.Search != "" {
		values.Set("search", filter.Search)
	}
	if success != "" {
		values.Set("success", success)
	}
	if errorMessage != "" {
		values.Set("error", errorMessage)
	}
	c.Redirect(http.StatusSeeOther, "/asset-depreciation/monthly?"+values.Encode())
}

func depreciationMonthOptions() []models.DepreciationMonthOption {
	months := make([]models.DepreciationMonthOption, 0, 12)
	for month := 1; month <= 12; month++ {
		months = append(months, models.DepreciationMonthOption{Value: month, Label: time.Month(month).String()})
	}
	return months
}

func depreciationYearOptions(selected int) []int {
	current := time.Now().Year()
	start, end := current-5, current+2
	if selected < start {
		start = selected
	}
	if selected > end {
		end = selected
	}
	years := make([]int, 0, end-start+1)
	for year := end; year >= start; year-- {
		years = append(years, year)
	}
	return years
}

func firstNonEmpty(values ...string) string {
	for _, value := range values {
		if strings.TrimSpace(value) != "" {
			return value
		}
	}
	return ""
}

func depreciationErrorMessage(err error) string {
	message := err.Error()
	lower := strings.ToLower(message)
	if strings.Contains(lower, "doesn't exist") && strings.Contains(lower, "asset_depreciation") {
		return "Tabel depresiasi belum tersedia. Jalankan SQL asset depreciation terlebih dahulu."
	}
	if strings.Contains(lower, "procedure") && strings.Contains(lower, "does not exist") {
		return "Stored procedure generator depresiasi belum tersedia. Jalankan SQL procedure depresiasi terlebih dahulu."
	}
	return message
}
