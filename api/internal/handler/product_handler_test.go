package handler

import (
	"fmt"
	"net/http"
	"testing"

	"github.com/KoZeuh/GO-Project-ESGI/api/internal/dto"
	"github.com/KoZeuh/GO-Project-ESGI/api/internal/models"
)

func createTestSupplier(t *testing.T, ts *testServer, token string) models.Supplier {
	t.Helper()
	rec := ts.do(t, http.MethodPost, "/api/v1/suppliers", token, dto.SupplierRequest{Name: "Fournisseur A"})
	if rec.Code != http.StatusCreated {
		t.Fatalf("POST /suppliers status = %d, corps = %s", rec.Code, rec.Body.String())
	}
	return decodeJSON[models.Supplier](t, rec)
}

func TestProductHandler_CRUD(t *testing.T) {
	ts := newTestServer(t)
	token := ts.authToken(t)
	supplier := createTestSupplier(t, ts, token)

	rec := ts.do(t, http.MethodGet, "/api/v1/products", token, nil)
	if rec.Code != http.StatusOK {
		t.Fatalf("GET /products status = %d", rec.Code)
	}
	if list := decodeJSON[[]models.Product](t, rec); len(list) != 0 {
		t.Fatalf("GET /products initial = %d éléments, attendu 0", len(list))
	}

	rec = ts.do(t, http.MethodPost, "/api/v1/products", token, dto.ProductRequest{
		Name: "Café", Reference: "CAF-1", Quantity: 50, Price: 9.9, SupplierID: supplier.ID,
	})
	if rec.Code != http.StatusCreated {
		t.Fatalf("POST /products status = %d, corps = %s", rec.Code, rec.Body.String())
	}
	product := decodeJSON[models.Product](t, rec)

	rec = ts.do(t, http.MethodGet, fmt.Sprintf("/api/v1/products/%d", product.ID), token, nil)
	if rec.Code != http.StatusOK {
		t.Fatalf("GET /products/:id status = %d", rec.Code)
	}

	rec = ts.do(t, http.MethodPut, fmt.Sprintf("/api/v1/products/%d", product.ID), token, dto.ProductRequest{
		Name: "Café Arabica", Reference: "CAF-1", Quantity: 50, Price: 10.5, SupplierID: supplier.ID,
	})
	if rec.Code != http.StatusOK {
		t.Fatalf("PUT /products/:id status = %d, corps = %s", rec.Code, rec.Body.String())
	}
	updated := decodeJSON[models.Product](t, rec)
	if updated.Name != "Café Arabica" {
		t.Fatalf("PUT /products/:id nom = %q, attendu \"Café Arabica\"", updated.Name)
	}

	rec = ts.do(t, http.MethodDelete, fmt.Sprintf("/api/v1/products/%d", product.ID), token, nil)
	if rec.Code != http.StatusNoContent {
		t.Fatalf("DELETE /products/:id status = %d", rec.Code)
	}

	rec = ts.do(t, http.MethodGet, fmt.Sprintf("/api/v1/products/%d", product.ID), token, nil)
	if rec.Code != http.StatusNotFound {
		t.Fatalf("GET /products/:id après suppression status = %d, attendu 404", rec.Code)
	}
}

func TestProductHandler_CreateWithUnknownSupplierReturns400(t *testing.T) {
	ts := newTestServer(t)
	token := ts.authToken(t)

	rec := ts.do(t, http.MethodPost, "/api/v1/products", token, dto.ProductRequest{
		Name: "Café", Reference: "CAF-1", Quantity: 10, Price: 5, SupplierID: 999,
	})
	if rec.Code != http.StatusBadRequest {
		t.Fatalf("POST /products avec fournisseur inconnu status = %d, attendu 400", rec.Code)
	}
}

func TestAlertsHandler_LowStockAndSettings(t *testing.T) {
	ts := newTestServer(t)
	token := ts.authToken(t)
	supplier := createTestSupplier(t, ts, token)

	ts.do(t, http.MethodPost, "/api/v1/products", token, dto.ProductRequest{
		Name: "Lait", Reference: "LAI-1", Quantity: 2, Price: 1.1, SupplierID: supplier.ID,
	})

	rec := ts.do(t, http.MethodGet, "/api/v1/alerts", token, nil)
	if rec.Code != http.StatusOK {
		t.Fatalf("GET /alerts status = %d", rec.Code)
	}
	alerts := decodeJSON[dto.AlertsResponse](t, rec)
	if len(alerts.LowStockProducts) != 1 {
		t.Fatalf("GET /alerts produits en alerte = %d, attendu 1", len(alerts.LowStockProducts))
	}

	rec = ts.do(t, http.MethodPut, "/api/v1/alerts/settings", token, dto.AlertSettingsRequest{DefaultThreshold: 1})
	if rec.Code != http.StatusOK {
		t.Fatalf("PUT /alerts/settings status = %d, corps = %s", rec.Code, rec.Body.String())
	}

	rec = ts.do(t, http.MethodGet, "/api/v1/alerts", token, nil)
	alerts = decodeJSON[dto.AlertsResponse](t, rec)
	if len(alerts.LowStockProducts) != 0 {
		t.Fatalf("GET /alerts après abaissement du seuil = %d produits, attendu 0", len(alerts.LowStockProducts))
	}
}
