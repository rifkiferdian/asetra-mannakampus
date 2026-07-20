package services

import (
	"errors"
	"gobase-app/models"
	"gobase-app/repositories"
	"strings"
	"time"
)

type AssetDisposalReportService struct {
	Repo *repositories.AssetDisposalReportRepository
}

func (s *AssetDisposalReportService) GetReport(filter models.AssetDisposalReportFilter) (models.AssetDisposalReportResult, error) {
	filter, err := normalizeAssetDisposalReportFilter(filter)
	if err != nil {
		return models.AssetDisposalReportResult{}, err
	}
	return s.Repo.GetReport(filter)
}

func (s *AssetDisposalReportService) GetExportRows(filter models.AssetDisposalReportFilter) ([]models.AssetDisposalReportRow, error) {
	filter, err := normalizeAssetDisposalReportFilter(filter)
	if err != nil {
		return nil, err
	}
	return s.Repo.GetExportRows(filter)
}

func normalizeAssetDisposalReportFilter(filter models.AssetDisposalReportFilter) (models.AssetDisposalReportFilter, error) {
	from, err := time.Parse("2006-01-02", strings.TrimSpace(filter.DateFrom))
	if err != nil {
		return filter, errors.New("tanggal awal laporan tidak valid")
	}
	until, err := time.Parse("2006-01-02", strings.TrimSpace(filter.DateTo))
	if err != nil {
		return filter, errors.New("tanggal akhir laporan tidak valid")
	}
	if until.Before(from) {
		return filter, errors.New("tanggal akhir tidak boleh sebelum tanggal awal")
	}
	filter.DateFrom, filter.DateTo = from.Format("2006-01-02"), until.Format("2006-01-02")
	filter.Status = strings.ToUpper(strings.TrimSpace(filter.Status))
	if filter.Status == "" {
		filter.Status = "POSTED"
	}
	if filter.Status != "ALL" && filter.Status != "POSTED" && filter.Status != "REVERSED" {
		return filter, errors.New("status laporan tidak valid")
	}
	filter.Result = strings.ToUpper(strings.TrimSpace(filter.Result))
	if filter.Result == "" {
		filter.Result = "ALL"
	}
	if filter.Result != "ALL" && filter.Result != "PROFIT" && filter.Result != "LOSS" && filter.Result != "BREAK_EVEN" {
		return filter, errors.New("hasil disposal tidak valid")
	}
	filter.Search = strings.TrimSpace(filter.Search)
	if filter.Page < 1 {
		filter.Page = 1
	}
	if filter.PerPage <= 0 || filter.PerPage > 100 {
		filter.PerPage = 50
	}
	return filter, nil
}
