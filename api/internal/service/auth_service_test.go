package service

import (
	"errors"
	"testing"
	"time"

	"github.com/KoZeuh/GO-Project-ESGI/api/internal/repository"
)

func newTestAuthService(t *testing.T) *AuthService {
	db := newTestDB(t)
	return NewAuthService(repository.NewUserRepository(db), "test-secret", time.Hour)
}

func TestAuthService_RegisterAndLogin(t *testing.T) {
	auth := newTestAuthService(t)

	user, err := auth.Register("bob", "secret123")
	if err != nil {
		t.Fatalf("Register() error = %v", err)
	}
	if user.Username != "bob" || user.Role != "employee" {
		t.Fatalf("Register() résultat inattendu = %+v", user)
	}

	token, expiresAt, loggedIn, err := auth.Login("bob", "secret123")
	if err != nil {
		t.Fatalf("Login() error = %v", err)
	}
	if token == "" {
		t.Fatal("Login() a renvoyé un token vide")
	}
	if !expiresAt.After(time.Now()) {
		t.Fatal("Login() a renvoyé une date d'expiration passée")
	}
	if loggedIn.ID != user.ID {
		t.Fatalf("Login() a renvoyé l'utilisateur %d, attendu %d", loggedIn.ID, user.ID)
	}

	claims, err := auth.ValidateToken(token)
	if err != nil {
		t.Fatalf("ValidateToken() error = %v", err)
	}
	if claims.UserID != user.ID || claims.Username != "bob" {
		t.Fatalf("ValidateToken() claims inattendus = %+v", claims)
	}
}

func TestAuthService_RegisterDuplicateUsername(t *testing.T) {
	auth := newTestAuthService(t)

	if _, err := auth.Register("carol", "secret123"); err != nil {
		t.Fatalf("premier Register() error = %v", err)
	}
	if _, err := auth.Register("carol", "autreMotDePasse"); !errors.Is(err, repository.ErrDuplicate) {
		t.Fatalf("Register() en doublon = %v, attendu ErrDuplicate", err)
	}
}

func TestAuthService_LoginWrongPassword(t *testing.T) {
	auth := newTestAuthService(t)

	if _, err := auth.Register("dave", "correctPassword"); err != nil {
		t.Fatalf("Register() error = %v", err)
	}
	if _, _, _, err := auth.Login("dave", "wrongPassword"); !errors.Is(err, ErrInvalidCredentials) {
		t.Fatalf("Login() avec mauvais mot de passe = %v, attendu ErrInvalidCredentials", err)
	}
	if _, _, _, err := auth.Login("inconnu", "peuImporte"); !errors.Is(err, ErrInvalidCredentials) {
		t.Fatalf("Login() avec utilisateur inconnu = %v, attendu ErrInvalidCredentials", err)
	}
}
