package service

import (
	"database/sql"
	"errors"
	"testing"

	"github.com/KoZeuh/GO-Project-ESGI/api/internal/dto"
	"github.com/KoZeuh/GO-Project-ESGI/api/internal/repository"
)

func newTestProductService(t *testing.T, db *sql.DB) *ProductService {
	t.Helper()
	productRepo := repository.NewProductRepository(db)
	supplierRepo := repository.NewSupplierRepository(db)
	alertService := NewAlertService(repository.NewAlertRepository(db), productRepo)
	return NewProductService(productRepo, supplierRepo, alertService)
}

func TestProductService_CreateRejectsUnknownSupplier(t *testing.T) {
	db := newTestDB(t)
	products := newTestProductService(t, db)

	_, err := products.Create(dto.ProductRequest{
		Name: "Café", Reference: "CAF-1", Quantity: 10, Price: 5, SupplierID: 999,
	})
	if !errors.Is(err, ErrSupplierNotFound) {
		t.Fatalf("Create() avec fournisseur inconnu = %v, attendu ErrSupplierNotFound", err)
	}
}

func TestProductService_CreateAndLowStockFlag(t *testing.T) {
	db := newTestDB(t)
	supplierID := seedSupplier(t, db)
	products := newTestProductService(t, db)

	// Le seuil par défaut (5, voir schema.sql) s'applique car alert_threshold est nil.
	product, err := products.Create(dto.ProductRequest{
		Name: "Thé vert", Reference: "THE-1", Quantity: 3, Price: 4.5, SupplierID: supplierID,
	})
	if err != nil {
		t.Fatalf("Create() error = %v", err)
	}
	if !product.IsLowStock {
		t.Fatalf("Create() produit avec quantité 3 < seuil par défaut 5 devrait être IsLowStock=true")
	}
	if product.SupplierName == "" {
		t.Fatal("Create() devrait renvoyer le nom du fournisseur joint")
	}

	list, err := products.List("")
	if err != nil {
		t.Fatalf("List() error = %v", err)
	}
	if len(list) != 1 {
		t.Fatalf("List() a renvoyé %d produits, attendu 1", len(list))
	}
}

func TestProductService_UpdateAndDelete(t *testing.T) {
	db := newTestDB(t)
	supplierID := seedSupplier(t, db)
	products := newTestProductService(t, db)

	created, err := products.Create(dto.ProductRequest{
		Name: "Sucre", Reference: "SUC-1", Quantity: 20, Price: 2, SupplierID: supplierID,
	})
	if err != nil {
		t.Fatalf("Create() error = %v", err)
	}

	threshold := 15
	updated, err := products.Update(created.ID, dto.ProductRequest{
		Name: "Sucre roux", Reference: "SUC-1", Quantity: 20, Price: 2.5,
		AlertThreshold: &threshold, SupplierID: supplierID,
	})
	if err != nil {
		t.Fatalf("Update() error = %v", err)
	}
	if updated.Name != "Sucre roux" || updated.Price != 2.5 || updated.IsLowStock {
		t.Fatalf("Update() résultat inattendu (quantité 20 > seuil 15, ne devrait pas être en alerte) = %+v", updated)
	}

	if err := products.Delete(created.ID); err != nil {
		t.Fatalf("Delete() error = %v", err)
	}
	if _, err := products.Get(created.ID); !errors.Is(err, repository.ErrNotFound) {
		t.Fatalf("Get() après suppression = %v, attendu ErrNotFound", err)
	}
}
