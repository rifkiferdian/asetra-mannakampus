package services

import (
	"errors"
	"gobase-app/models"
	"gobase-app/repositories"
	"strings"
	"time"
)

type AssetDepreciationAnnualReportService struct {
	Repo *repositories.AssetDepreciationAnnualReportRepository
}

func (s *AssetDepreciationAnnualReportService) GetReport(filter models.AnnualDepreciationReportFilter) (models.AnnualDepreciationReportResult, error) {
	filter, err := normalizeAnnualDepreciationFilter(filter)
	if err != nil {
		return models.AnnualDepreciationReportResult{}, err
	}
	return s.Repo.GetReport(filter)
}

func normalizeAnnualDepreciationFilter(filter models.AnnualDepreciationReportFilter) (models.AnnualDepreciationReportFilter, error) {
	currentYear := time.Now().Year()
	if filter.YearFrom == 0 {
		filter.YearFrom = currentYear - 2
	}
	if filter.YearTo == 0 {
		filter.YearTo = currentYear
	}
	if filter.YearFrom < 1900 || filter.YearTo > 2200 {
		return filter, errors.New("rentang tahun laporan tidak valid")
	}
	if filter.YearTo < filter.YearFrom {
		return filter, errors.New("tahun akhir tidak boleh sebelum tahun awal")
	}
	if filter.YearTo-filter.YearFrom+1 > 10 {
		return filter, errors.New("rentang laporan maksimal 10 tahun")
	}
	filter.Mode = strings.ToUpper(strings.TrimSpace(filter.Mode))
	if filter.Mode == "" {
		filter.Mode = "ACTUAL"
	}
	if filter.Mode != "ACTUAL" && filter.Mode != "PROJECTION" {
		return filter, errors.New("mode laporan tidak valid")
	}
	filter.AssetStatus = strings.ToUpper(strings.TrimSpace(filter.AssetStatus))
	if filter.AssetStatus == "" {
		filter.AssetStatus = "ALL"
	}
	validStatuses := map[string]bool{"ALL": true, "AVAILABLE": true, "IN_USE": true, "MAINTENANCE": true, "BROKEN": true, "DISPOSED": true}
	if !validStatuses[filter.AssetStatus] {
		return filter, errors.New("status aset tidak valid")
	}
	filter.Search = strings.TrimSpace(filter.Search)
	return filter, nil
}
