package handler

import (
	"net/http"
	"testing"

	"github.com/KoZeuh/GO-Project-ESGI/api/internal/dto"
)

func TestAuthHandler_RegisterAndLogin(t *testing.T) {
	ts := newTestServer(t)

	rec := ts.do(t, http.MethodPost, "/api/v1/auth/register", "", dto.RegisterRequest{
		Username: "eve", Password: "password123",
	})
	if rec.Code != http.StatusCreated {
		t.Fatalf("POST /auth/register status = %d, corps = %s", rec.Code, rec.Body.String())
	}
	created := decodeJSON[dto.UserResponse](t, rec)
	if created.Username != "eve" {
		t.Fatalf("Register() username = %q, attendu \"eve\"", created.Username)
	}

	// Le nom d'utilisateur est déjà pris : conflit attendu.
	rec = ts.do(t, http.MethodPost, "/api/v1/auth/register", "", dto.RegisterRequest{
		Username: "eve", Password: "autreMotDePasse",
	})
	if rec.Code != http.StatusConflict {
		t.Fatalf("POST /auth/register (doublon) status = %d, attendu 409", rec.Code)
	}

	rec = ts.do(t, http.MethodPost, "/api/v1/auth/login", "", dto.LoginRequest{
		Username: "eve", Password: "password123",
	})
	if rec.Code != http.StatusOK {
		t.Fatalf("POST /auth/login status = %d, corps = %s", rec.Code, rec.Body.String())
	}
	login := decodeJSON[dto.LoginResponse](t, rec)
	if login.Token == "" || login.User.Username != "eve" {
		t.Fatalf("Login() réponse inattendue = %+v", login)
	}

	rec = ts.do(t, http.MethodPost, "/api/v1/auth/login", "", dto.LoginRequest{
		Username: "eve", Password: "mauvais",
	})
	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("POST /auth/login (mauvais mdp) status = %d, attendu 401", rec.Code)
	}
}

func TestAuthMiddleware_RejectsMissingOrInvalidToken(t *testing.T) {
	ts := newTestServer(t)

	rec := ts.do(t, http.MethodGet, "/api/v1/products", "", nil)
	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("GET /products sans token status = %d, attendu 401", rec.Code)
	}

	rec = ts.do(t, http.MethodGet, "/api/v1/products", "token-invalide", nil)
	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("GET /products avec token invalide status = %d, attendu 401", rec.Code)
	}
}
