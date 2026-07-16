package services

import (
	"errors"
	"gobase-app/models"
	"gobase-app/repositories"
	"strings"
)

type AssetDepreciationService struct {
	Repo *repositories.AssetDepreciationRepository
}

func (s *AssetDepreciationService) GetAssetDepreciationDetail(assetID int64) (models.AssetDepreciationDetail, error) {
	if assetID <= 0 {
		return models.AssetDepreciationDetail{}, errors.New("asset tidak valid")
	}
	return s.Repo.GetAssetDepreciationDetail(assetID)
}

func (s *AssetDepreciationService) GetPostedDepreciationByAssetID(assetID int64, limit int) ([]models.AssetDepreciationPosting, error) {
	if assetID <= 0 {
		return nil, errors.New("asset tidak valid")
	}
	return s.Repo.GetPostedDepreciationByAssetID(assetID, limit)
}

func (s *AssetDepreciationService) GetMonthlyDepreciation(filter models.MonthlyDepreciationFilter) (models.MonthlyDepreciationResult, error) {
	if filter.Year < 2000 || filter.Year > 2200 {
		return models.MonthlyDepreciationResult{}, errors.New("tahun depresiasi tidak valid")
	}
	if filter.Month < 1 || filter.Month > 12 {
		return models.MonthlyDepreciationResult{}, errors.New("bulan depresiasi tidak valid")
	}
	filter.Status = strings.ToUpper(strings.TrimSpace(filter.Status))
	if filter.Status == "" {
		filter.Status = "ALL"
	}
	if filter.Status != "ALL" && filter.Status != "DRAFT" && filter.Status != "POSTED" && filter.Status != "SKIPPED" {
		return models.MonthlyDepreciationResult{}, errors.New("status depresiasi tidak valid")
	}
	filter.Search = strings.TrimSpace(filter.Search)
	if filter.Page < 1 {
		filter.Page = 1
	}
	if filter.PerPage <= 0 || filter.PerPage > 100 {
		filter.PerPage = 50
	}
	return s.Repo.GetMonthlyDepreciation(filter)
}

func (s *AssetDepreciationService) GenerateSchedules() error {
	return s.Repo.GenerateAllSchedules()
}

func (s *AssetDepreciationService) PostSchedules(ids []int64) (int64, error) {
	unique := make([]int64, 0, len(ids))
	seen := make(map[int64]bool, len(ids))
	for _, id := range ids {
		if id <= 0 || seen[id] {
			continue
		}
		seen[id] = true
		unique = append(unique, id)
	}
	return s.Repo.PostSchedules(unique)
}
