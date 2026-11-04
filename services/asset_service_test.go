package services

import (
	"gobase-app/models"
	"strings"
	"testing"
)

type fakeAssetRepository struct {
	AssetRepositoryPort
	createdAsset *models.AssetInput
	asset        *models.Asset
}

func (f *fakeAssetRepository) AssetCategoryBelongsToType(categoryID, assetTypeID int64) (bool, error) {
	return categoryID == 1 && assetTypeID == 1, nil
}

func (f *fakeAssetRepository) LocationBelongsToStore(locationID int64, storeID int) (bool, error) {
	return locationID == 10 && storeID == 2, nil
}

func (f *fakeAssetRepository) CreateAsset(input models.AssetInput) error {
	f.createdAsset = &input
	return nil
}

func (f *fakeAssetRepository) GetAssetByID(id int64, scopes ...models.AccessScope) (*models.Asset, error) {
	return f.asset, nil
}

func TestSaveAssetRejectsInvalidLifecycleInputBeforeRepository(t *testing.T) {
	tests := []struct {
		name  string
		input models.AssetInput
		want  string
	}{
		{
			name: "invalid acquisition date",
			input: validAssetInput(func(input *models.AssetInput) {
				input.AcquisitionDate = "03/11/2026"
			}),
			want: "tanggal perolehan tidak valid",
		},
		{
			name: "negative acquisition value",
			input: validAssetInput(func(input *models.AssetInput) {
				input.AcquisitionValue = models.MustMoney("-1")
			}),
			want: "nilai perolehan tidak valid",
		},
		{
			name: "disposed bypass",
			input: validAssetInput(func(input *models.AssetInput) {
				input.Status = models.AssetStatusDisposed
			}),
			want: "hanya dapat dibuat melalui proses disposal",
		},
	}

	service := &AssetService{}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			err := service.SaveAsset(test.input)
			if err == nil || !strings.Contains(err.Error(), test.want) {
				t.Fatalf("SaveAsset() error = %v, want containing %q", err, test.want)
			}
		})
	}
}

func TestAssetMovementLifecycle(t *testing.T) {
	if allowedAssetMovementType("DISPOSE") {
		t.Fatal("DISPOSE must go through the dedicated disposal workflow")
	}
	if got := nextAssetStatus(models.AssetMovementReturn, 0); got != models.AssetStatusAvailable {
		t.Fatalf("return status = %q, want %q", got, models.AssetStatusAvailable)
	}
	if got := nextAssetStatus(models.AssetMovementAssign, 12); got != models.AssetStatusInUse {
		t.Fatalf("assign status = %q, want %q", got, models.AssetStatusInUse)
	}
	if got := nextAssetStatus(models.AssetMovementMaintenance, 0); got != models.AssetStatusMaintenance {
		t.Fatalf("maintenance status = %q, want %q", got, models.AssetStatusMaintenance)
	}
}

func TestSaveAssetCreatesValidatedMoneyValueWithinStoreScope(t *testing.T) {
	repo := &fakeAssetRepository{}
	service := NewAssetService(repo)
	input := validAssetInput(func(input *models.AssetInput) {
		input.StoreID = 2
		input.LocationID = 10
		input.AcquisitionValue = models.MustMoney("1250000.75")
	})
	if err := service.SaveAsset(input, models.AccessScope{StoreIDs: []int{2}}); err != nil {
		t.Fatalf("SaveAsset() error = %v", err)
	}
	if repo.createdAsset == nil {
		t.Fatal("repository CreateAsset was not called")
	}
	if got := repo.createdAsset.AcquisitionValue.DecimalString(); got != "1250000.75" {
		t.Fatalf("stored acquisition value = %q", got)
	}
}

func TestSaveAssetRejectsStoreOutsideScope(t *testing.T) {
	repo := &fakeAssetRepository{}
	service := NewAssetService(repo)
	input := validAssetInput(func(input *models.AssetInput) {
		input.StoreID = 2
		input.LocationID = 10
	})
	err := service.SaveAsset(input, models.AccessScope{StoreIDs: []int{7}})
	if err == nil || !strings.Contains(err.Error(), "akses ke store") {
		t.Fatalf("SaveAsset() error = %v, want store access error", err)
	}
	if repo.createdAsset != nil {
		t.Fatal("repository CreateAsset must not be called")
	}
}

func TestComponentMovementLifecycle(t *testing.T) {
	if !allowedComponentMovementType(models.ComponentMovementDispose) {
		t.Fatal("component disposal must remain an explicit movement")
	}
	if got := nextComponentStatus(models.ComponentMovementInstall, 42); got != models.ComponentStatusInstalled {
		t.Fatalf("install status = %q, want %q", got, models.ComponentStatusInstalled)
	}
	if got := nextComponentStatus(models.ComponentMovementUninstall, 0); got != models.ComponentStatusInStorage {
		t.Fatalf("uninstall status = %q, want %q", got, models.ComponentStatusInStorage)
	}
	if got := nextComponentStatus(models.ComponentMovementDispose, 0); got != models.ComponentStatusDisposed {
		t.Fatalf("dispose status = %q, want %q", got, models.ComponentStatusDisposed)
	}
}

func validAssetInput(change func(*models.AssetInput)) models.AssetInput {
	input := models.AssetInput{
		AssetName:        "Laptop Kasir",
		AssetTypeID:      1,
		AssetCategoryID:  1,
		AcquisitionDate:  "2026-11-03",
		AcquisitionValue: models.MustMoney("1000000"),
		Status:           models.AssetStatusAvailable,
	}
	change(&input)
	return input
}
