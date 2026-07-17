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
	periodLabel := depreciationPeriodLabel(filter.Year, filter.Month)
	result, err := assetDepreciationService().GetMonthlyDepreciation(filter)
	if err != nil {
		Render(c, "monthly_depreciation.html", gin.H{
			"Title": "Depresiasi Bulanan", "Page": "monthly_depreciation",
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
		"Title":       "Depresiasi Bulanan",
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
	count, err := assetDepreciationService().GenerateMonthlySchedules(filter.Year, filter.Month, depreciationAuditContext(c))
	if err != nil {
		redirectMonthlyDepreciation(c, filter, "", depreciationErrorMessage(err))
		return
	}
	periodLabel := depreciationPeriodLabel(filter.Year, filter.Month)
	message := strconv.Itoa(count) + " jadwal depresiasi tersedia untuk periode " + periodLabel
	if count == 0 {
		message = "Tidak ada jadwal yang dapat dibuat untuk periode " + periodLabel + ". Pastikan profil aktif dan periode sebelumnya sudah DIPOSTING atau DILEWATI."
	}
	redirectMonthlyDepreciation(c, filter, message, "")
}

func MonthlyDepreciationPost(c *gin.Context) {
	filter := monthlyDepreciationFilter(c)
	count, err := assetDepreciationService().PostSchedules(depreciationScheduleIDs(c), depreciationAuditContext(c))
	if err != nil {
		redirectMonthlyDepreciation(c, filter, "", depreciationErrorMessage(err))
		return
	}
	redirectMonthlyDepreciation(c, filter, strconv.FormatInt(count, 10)+" depresiasi berhasil diposting", "")
}

func MonthlyDepreciationSkip(c *gin.Context) {
	filter := monthlyDepreciationFilter(c)
	count, err := assetDepreciationService().SkipSchedules(
		depreciationScheduleIDs(c),
		c.PostForm("skip_reason"),
		depreciationAuditContext(c),
	)
	if err != nil {
		redirectMonthlyDepreciation(c, filter, "", depreciationErrorMessage(err))
		return
	}
	redirectMonthlyDepreciation(c, filter, strconv.FormatInt(count, 10)+" depresiasi berhasil dilewati", "")
}

func DepreciationProfileIndex(c *gin.Context) {
	filter := depreciationProfileFilter(c)
	service := assetDepreciationService()
	result, err := service.GetDepreciationProfiles(filter)
	if err != nil {
		c.String(http.StatusInternalServerError, depreciationErrorMessage(err))
		return
	}
	methods, err := service.GetDepreciationMethods()
	if err != nil {
		c.String(http.StatusInternalServerError, depreciationErrorMessage(err))
		return
	}
	assets, err := service.GetDepreciationAssetOptions()
	if err != nil {
		c.String(http.StatusInternalServerError, depreciationErrorMessage(err))
		return
	}
	Render(c, "depreciation_profiles.html", gin.H{
		"Title":      "Profil Depresiasi",
		"Page":       "depreciation_profiles",
		"Items":      result.Items,
		"Stats":      result.Stats,
		"Methods":    methods,
		"Assets":     assets,
		"Filter":     filter,
		"Pagination": depreciationProfilePagination(filter, result),
		"Success":    strings.TrimSpace(c.Query("success")),
		"Error":      strings.TrimSpace(c.Query("error")),
	})
}

func DepreciationProfileSave(c *gin.Context) {
	input := models.DepreciationProfileInput{
		ID:               parseInt64Form(c, "id"),
		AssetID:          parseInt64Form(c, "asset_id"),
		MethodID:         parseInt64Form(c, "depreciation_method_id"),
		UsefulLifeMonths: parseIntForm(c, "useful_life_months"),
		SalvageValue:     parseFloatForm(c, "salvage_value"),
		DepreciableBasis: parseFloatForm(c, "depreciable_basis"),
		StartDate:        c.PostForm("start_date"),
		Status:           c.PostForm("status"),
		Notes:            c.PostForm("notes"),
		AuditContext:     depreciationAuditContext(c),
	}
	if err := assetDepreciationService().SaveDepreciationProfile(input); err != nil {
		c.Redirect(http.StatusSeeOther, "/asset-depreciation/profiles?error="+url.QueryEscape(depreciationErrorMessage(err)))
		return
	}
	message := "Profil depresiasi berhasil disimpan"
	c.Redirect(http.StatusSeeOther, "/asset-depreciation/profiles?success="+url.QueryEscape(message))
}

func DepreciationPostingHistoryIndex(c *gin.Context) {
	filter := depreciationPostingHistoryFilter(c)
	result, err := assetDepreciationService().GetPostingHistory(filter)
	if err != nil {
		c.String(http.StatusInternalServerError, depreciationErrorMessage(err))
		return
	}
	Render(c, "depreciation_posting_history.html", gin.H{
		"Title":      "Riwayat Posting Depresiasi",
		"Page":       "depreciation_posting_history",
		"Items":      result.Items,
		"Stats":      result.Stats,
		"Filter":     filter,
		"Months":     depreciationMonthOptions(),
		"Years":      depreciationYearOptions(filter.Year),
		"Pagination": depreciationPostingHistoryPagination(filter, result),
	})
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

func depreciationProfileFilter(c *gin.Context) models.DepreciationProfileFilter {
	page, _ := strconv.Atoi(c.Query("page"))
	if page < 1 {
		page = 1
	}
	status := strings.ToUpper(strings.TrimSpace(c.Query("status")))
	if status == "" {
		status = "ALL"
	}
	return models.DepreciationProfileFilter{
		Status: status, Search: strings.TrimSpace(c.Query("search")), Page: page, PerPage: monthlyDepreciationPageSize,
	}
}

func depreciationPostingHistoryFilter(c *gin.Context) models.DepreciationPostingHistoryFilter {
	year, _ := strconv.Atoi(c.Query("year"))
	month, _ := strconv.Atoi(c.Query("month"))
	page, _ := strconv.Atoi(c.Query("page"))
	if year == 0 {
		year = time.Now().Year()
	}
	if page < 1 {
		page = 1
	}
	return models.DepreciationPostingHistoryFilter{
		Year: year, Month: month, Search: strings.TrimSpace(c.Query("search")), Page: page, PerPage: monthlyDepreciationPageSize,
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

func depreciationProfilePagination(filter models.DepreciationProfileFilter, result models.DepreciationProfileResult) models.DepreciationPagination {
	page := filter.Page
	if page > result.TotalPages {
		page = result.TotalPages
	}
	start, end := paginationBounds(page, filter.PerPage, result.TotalRows, len(result.Items))
	return models.DepreciationPagination{
		CurrentPage: page, TotalPages: result.TotalPages, PageStart: start, PageEnd: end,
		TotalRows: result.TotalRows, PageSize: filter.PerPage, HasPrev: page > 1, HasNext: page < result.TotalPages,
		PrevURL: depreciationProfileURL(filter, page-1), NextURL: depreciationProfileURL(filter, page+1),
	}
}

func depreciationPostingHistoryPagination(filter models.DepreciationPostingHistoryFilter, result models.DepreciationPostingHistoryResult) models.DepreciationPagination {
	page := filter.Page
	if page > result.TotalPages {
		page = result.TotalPages
	}
	start, end := paginationBounds(page, filter.PerPage, result.TotalRows, len(result.Items))
	return models.DepreciationPagination{
		CurrentPage: page, TotalPages: result.TotalPages, PageStart: start, PageEnd: end,
		TotalRows: result.TotalRows, PageSize: filter.PerPage, HasPrev: page > 1, HasNext: page < result.TotalPages,
		PrevURL: depreciationPostingHistoryURL(filter, page-1), NextURL: depreciationPostingHistoryURL(filter, page+1),
	}
}

func paginationBounds(page, pageSize, totalRows, itemCount int) (int, int) {
	if totalRows == 0 {
		return 0, 0
	}
	start := (page-1)*pageSize + 1
	return start, start + itemCount - 1
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

func depreciationProfileURL(filter models.DepreciationProfileFilter, page int) string {
	values := url.Values{}
	values.Set("status", filter.Status)
	values.Set("page", strconv.Itoa(page))
	if filter.Search != "" {
		values.Set("search", filter.Search)
	}
	return "/asset-depreciation/profiles?" + values.Encode()
}

func depreciationPostingHistoryURL(filter models.DepreciationPostingHistoryFilter, page int) string {
	values := url.Values{}
	values.Set("year", strconv.Itoa(filter.Year))
	values.Set("month", strconv.Itoa(filter.Month))
	values.Set("page", strconv.Itoa(page))
	if filter.Search != "" {
		values.Set("search", filter.Search)
	}
	return "/asset-depreciation/posting-history?" + values.Encode()
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
	labels := []string{"Januari", "Februari", "Maret", "April", "Mei", "Juni", "Juli", "Agustus", "September", "Oktober", "November", "Desember"}
	months := make([]models.DepreciationMonthOption, 0, 12)
	for month := 1; month <= 12; month++ {
		months = append(months, models.DepreciationMonthOption{Value: month, Label: labels[month-1]})
	}
	return months
}

func depreciationPeriodLabel(year, month int) string {
	labels := []string{"Januari", "Februari", "Maret", "April", "Mei", "Juni", "Juli", "Agustus", "September", "Oktober", "November", "Desember"}
	if month < 1 || month > len(labels) {
		return strconv.Itoa(year)
	}
	return labels[month-1] + " " + strconv.Itoa(year)
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

func depreciationScheduleIDs(c *gin.Context) []int64 {
	ids := make([]int64, 0)
	for _, value := range c.PostFormArray("schedule_ids") {
		id, err := strconv.ParseInt(strings.TrimSpace(value), 10, 64)
		if err == nil && id > 0 {
			ids = append(ids, id)
		}
	}
	return ids
}

func depreciationAuditContext(c *gin.Context) models.AuditContext {
	return models.AuditContext{
		ActorUserID: currentSessionUserID(c),
		IPAddress:   c.ClientIP(),
		UserAgent:   c.Request.UserAgent(),
	}
}

func depreciationErrorMessage(err error) string {
	message := err.Error()
	lower := strings.ToLower(message)
	if strings.Contains(lower, "doesn't exist") && strings.Contains(lower, "asset_depreciation") {
		return "Tabel depresiasi belum tersedia. Jalankan SQL depresiasi aset terlebih dahulu."
	}
	if strings.Contains(lower, "procedure") && strings.Contains(lower, "does not exist") {
		return "Stored procedure generator depresiasi belum tersedia. Jalankan SQL procedure depresiasi terlebih dahulu."
	}
	return message
}
