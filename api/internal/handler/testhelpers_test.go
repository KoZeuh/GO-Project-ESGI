package handler

import (
	"bytes"
	"encoding/json"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"

	"github.com/KoZeuh/GO-Project-ESGI/api/internal/database"
	"github.com/KoZeuh/GO-Project-ESGI/api/internal/repository"
	"github.com/KoZeuh/GO-Project-ESGI/api/internal/service"
)

func init() {
	gin.SetMode(gin.TestMode)
}

// testServer assemble une instance complète de l'API (DB en mémoire incluse) pour des tests d'intégration des handlers via httptest.
type testServer struct {
	router *gin.Engine
	auth   *service.AuthService
}

func newTestServer(t *testing.T) *testServer {
	t.Helper()

	db, err := database.Connect(":memory:")
	if err != nil {
		t.Fatalf("connexion base de test : %v", err)
	}
	t.Cleanup(func() { db.Close() })

	userRepo := repository.NewUserRepository(db)
	supplierRepo := repository.NewSupplierRepository(db)
	productRepo := repository.NewProductRepository(db)
	movementRepo := repository.NewMovementRepository(db)
	alertRepo := repository.NewAlertRepository(db)

	authService := service.NewAuthService(userRepo, "test-secret", time.Hour)
	alertService := service.NewAlertService(alertRepo, productRepo)
	productService := service.NewProductService(productRepo, supplierRepo, alertService)
	supplierService := service.NewSupplierService(supplierRepo)
	movementService := service.NewMovementService(db, productRepo, movementRepo)

	handlers := Handlers{
		Auth:      NewAuthHandler(authService),
		Products:  NewProductHandler(productService),
		Suppliers: NewSupplierHandler(supplierService),
		Movements: NewMovementHandler(movementService),
		Alerts:    NewAlertHandler(alertService),
		Export:    NewExportHandler(productService, supplierService, movementService),
	}

	return &testServer{router: NewRouter(handlers, authService), auth: authService}
}

// authToken enregistre un utilisateur de test et retourne un JWT valide pour lui.
func (ts *testServer) authToken(t *testing.T) string {
	t.Helper()
	if _, err := ts.auth.Register("testuser", "password123"); err != nil {
		t.Fatalf("création utilisateur de test : %v", err)
	}
	token, _, _, err := ts.auth.Login("testuser", "password123")
	if err != nil {
		t.Fatalf("connexion utilisateur de test : %v", err)
	}
	return token
}

// do exécute une requête HTTP contre le routeur de test et renvoie la réponse.
func (ts *testServer) do(t *testing.T, method, path, token string, body any) *httptest.ResponseRecorder {
	t.Helper()

	var reader *bytes.Reader
	if body != nil {
		raw, err := json.Marshal(body)
		if err != nil {
			t.Fatalf("encodage du corps de requête : %v", err)
		}
		reader = bytes.NewReader(raw)
	} else {
		reader = bytes.NewReader(nil)
	}

	req := httptest.NewRequest(method, path, reader)
	req.Header.Set("Content-Type", "application/json")
	if token != "" {
		req.Header.Set("Authorization", "Bearer "+token)
	}

	rec := httptest.NewRecorder()
	ts.router.ServeHTTP(rec, req)
	return rec
}

func decodeJSON[T any](t *testing.T, rec *httptest.ResponseRecorder) T {
	t.Helper()
	var out T
	if err := json.Unmarshal(rec.Body.Bytes(), &out); err != nil {
		t.Fatalf("décodage JSON de la réponse (%s) : %v", rec.Body.String(), err)
	}
	return out
}
