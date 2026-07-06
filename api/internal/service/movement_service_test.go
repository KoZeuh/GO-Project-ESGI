package service

import (
	"errors"
	"testing"

	"github.com/KoZeuh/GO-Project-ESGI/api/internal/dto"
	"github.com/KoZeuh/GO-Project-ESGI/api/internal/models"
	"github.com/KoZeuh/GO-Project-ESGI/api/internal/repository"
)

func TestMovementService_CreateInIncreasesQuantity(t *testing.T) {
	db := newTestDB(t)
	supplierID := seedSupplier(t, db)
	userID := seedUser(t, db)

	productRepo := repository.NewProductRepository(db)
	movementRepo := repository.NewMovementRepository(db)
	products := newTestProductService(t, db)
	movements := NewMovementService(db, productRepo, movementRepo)

	product, err := products.Create(dto.ProductRequest{
		Name: "Farine", Reference: "FAR-1", Quantity: 10, Price: 1.2, SupplierID: supplierID,
	})
	if err != nil {
		t.Fatalf("Create() produit error = %v", err)
	}

	movement, err := movements.Create(dto.MovementRequest{
		ProductID: product.ID, Type: string(models.MovementIn), Quantity: 15, Note: "Réassort",
	}, userID)
	if err != nil {
		t.Fatalf("Create() mouvement error = %v", err)
	}
	if movement.Quantity != 15 || movement.Type != models.MovementIn {
		t.Fatalf("Create() mouvement inattendu = %+v", movement)
	}

	updated, err := products.Get(product.ID)
	if err != nil {
		t.Fatalf("Get() produit error = %v", err)
	}
	if updated.Quantity != 25 {
		t.Fatalf("quantité produit après entrée = %d, attendu 25", updated.Quantity)
	}
}

func TestMovementService_CreateOutRejectedWhenInsufficientStock(t *testing.T) {
	db := newTestDB(t)
	supplierID := seedSupplier(t, db)
	userID := seedUser(t, db)

	productRepo := repository.NewProductRepository(db)
	movementRepo := repository.NewMovementRepository(db)
	products := newTestProductService(t, db)
	movements := NewMovementService(db, productRepo, movementRepo)

	product, err := products.Create(dto.ProductRequest{
		Name: "Riz", Reference: "RIZ-1", Quantity: 5, Price: 3, SupplierID: supplierID,
	})
	if err != nil {
		t.Fatalf("Create() produit error = %v", err)
	}

	_, err = movements.Create(dto.MovementRequest{
		ProductID: product.ID, Type: string(models.MovementOut), Quantity: 10,
	}, userID)
	if !errors.Is(err, repository.ErrInsufficient) {
		t.Fatalf("Create() sortie excessive = %v, attendu ErrInsufficient", err)
	}

	// La transaction doit avoir été annulée : la quantité ne doit pas avoir bougé et aucun mouvement ne doit avoir été enregistré.
	unchanged, err := products.Get(product.ID)
	if err != nil {
		t.Fatalf("Get() produit error = %v", err)
	}
	if unchanged.Quantity != 5 {
		t.Fatalf("quantité produit après échec = %d, attendu 5 (inchangée)", unchanged.Quantity)
	}

	history, err := movements.List(repository.MovementFilter{ProductID: product.ID})
	if err != nil {
		t.Fatalf("List() mouvements error = %v", err)
	}
	if len(history) != 0 {
		t.Fatalf("List() mouvements = %d, attendu 0 (rollback)", len(history))
	}
}
