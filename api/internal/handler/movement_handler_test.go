package handler

import (
	"fmt"
	"net/http"
	"testing"

	"github.com/KoZeuh/GO-Project-ESGI/api/internal/dto"
	"github.com/KoZeuh/GO-Project-ESGI/api/internal/models"
)

func TestMovementHandler_CreateAndList(t *testing.T) {
	ts := newTestServer(t)
	token := ts.authToken(t)
	supplier := createTestSupplier(t, ts, token)

	rec := ts.do(t, http.MethodPost, "/api/v1/products", token, dto.ProductRequest{
		Name: "Farine", Reference: "FAR-1", Quantity: 10, Price: 1.5, SupplierID: supplier.ID,
	})
	product := decodeJSON[models.Product](t, rec)

	rec = ts.do(t, http.MethodPost, "/api/v1/movements", token, dto.MovementRequest{
		ProductID: product.ID, Type: "IN", Quantity: 20, Note: "Réassort",
	})
	if rec.Code != http.StatusCreated {
		t.Fatalf("POST /movements status = %d, corps = %s", rec.Code, rec.Body.String())
	}

	rec = ts.do(t, http.MethodPost, "/api/v1/movements", token, dto.MovementRequest{
		ProductID: product.ID, Type: "OUT", Quantity: 1000,
	})
	if rec.Code != http.StatusConflict {
		t.Fatalf("POST /movements sortie excessive status = %d, attendu 409, corps = %s", rec.Code, rec.Body.String())
	}

	rec = ts.do(t, http.MethodGet, "/api/v1/movements", token, nil)
	if rec.Code != http.StatusOK {
		t.Fatalf("GET /movements status = %d", rec.Code)
	}
	history := decodeJSON[[]models.Movement](t, rec)
	if len(history) != 1 {
		t.Fatalf("GET /movements = %d entrées, attendu 1 (la sortie refusée ne doit pas apparaître)", len(history))
	}

	rec = ts.do(t, http.MethodGet, fmt.Sprintf("/api/v1/products/%d", product.ID), token, nil)
	updatedProduct := decodeJSON[models.Product](t, rec)
	if updatedProduct.Quantity != 30 {
		t.Fatalf("quantité produit après entrée de 20 = %d, attendu 30", updatedProduct.Quantity)
	}
}
