package service

import (
	"database/sql"
	"testing"

	"github.com/KoZeuh/GO-Project-ESGI/api/internal/database"
	"github.com/KoZeuh/GO-Project-ESGI/api/internal/models"
	"github.com/KoZeuh/GO-Project-ESGI/api/internal/repository"
)

// newTestDB ouvre une base SQLite en mémoire avec le schéma appliqué, isolée pour chaque test.
func newTestDB(t *testing.T) *sql.DB {
	t.Helper()
	db, err := database.Connect(":memory:")
	if err != nil {
		t.Fatalf("connexion base de test : %v", err)
	}
	t.Cleanup(func() { db.Close() })
	return db
}

// seedSupplier crée un fournisseur de test et retourne son identifiant.
func seedSupplier(t *testing.T, db *sql.DB) int64 {
	t.Helper()
	id, err := repository.NewSupplierRepository(db).Create(&models.Supplier{
		Name:  "Fournisseur Test",
		Email: "contact@fournisseur-test.fr",
	})
	if err != nil {
		t.Fatalf("création fournisseur de test : %v", err)
	}
	return id
}

// seedUser crée un utilisateur de test via le service d'auth et retourne son identifiant.
func seedUser(t *testing.T, db *sql.DB) int64 {
	t.Helper()
	auth := NewAuthService(repository.NewUserRepository(db), "test-secret", 0)
	user, err := auth.Register("alice", "password123")
	if err != nil {
		t.Fatalf("création utilisateur de test : %v", err)
	}
	return user.ID
}
