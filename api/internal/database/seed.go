package database

import (
	"database/sql"
	"fmt"
	"log"

	"golang.org/x/crypto/bcrypt"
)

// Seed insère des données de démonstration si la base est vide.
// N'exécute rien si des utilisateurs existent déjà (idempotent).
func Seed(db *sql.DB) error {
	var count int
	if err := db.QueryRow("SELECT COUNT(*) FROM users").Scan(&count); err != nil {
		return fmt.Errorf("vérification seed : %w", err)
	}
	if count > 0 {
		return nil // données déjà présentes
	}

	log.Println("Base vide détectée — insertion des données de démonstration…")

	tx, err := db.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	// ── Utilisateurs ────────────────────────────────────────────────────────
	adminHash, err := bcrypt.GenerateFromPassword([]byte("admin123"), bcrypt.DefaultCost)
	if err != nil {
		return err
	}
	empHash, err := bcrypt.GenerateFromPassword([]byte("employee123"), bcrypt.DefaultCost)
	if err != nil {
		return err
	}

	_, err = tx.Exec(`
		INSERT INTO users (username, password_hash) VALUES
		('admin',    ?),
		('employee', ?)
	`, string(adminHash), string(empHash))
	if err != nil {
		return fmt.Errorf("seed users : %w", err)
	}

	// ── Fournisseurs ─────────────────────────────────────────────────────────
	_, err = tx.Exec(`
		INSERT INTO suppliers (name, email, phone, address) VALUES
		('Dupont SA',       'contact@dupont.fr',   '0102030405', '1 rue de la Paix, 75001 Paris'),
		('BioFournisseur',  'bio@biofourni.fr',    '0607080910', '12 allée des Champs, 69000 Lyon'),
		('Metro Cash',      'pro@metro.fr',        '0145678912', '8 avenue Gambetta, 13000 Marseille')
	`)
	if err != nil {
		return fmt.Errorf("seed suppliers : %w", err)
	}

	// ── Produits ─────────────────────────────────────────────────────────────
	threshold5 := 5
	threshold10 := 10
	_ = threshold5
	_ = threshold10

	_, err = tx.Exec(`
		INSERT INTO products (name, reference, quantity, price, alert_threshold, supplier_id) VALUES
		('Café en grains 1 kg',   'CAF-1KG-001',  3,  12.50, 5,    1),
		('Thé vert 500 g',        'THE-500G-002', 42,  8.00, NULL, 1),
		('Sucre blanc 1 kg',      'SUC-1KG-003',  15,  1.50, NULL, 2),
		('Farine T45 1 kg',       'FAR-T45-004',   8,  0.95, 10,   2),
		('Huile d''olive 75 cl',  'HUI-OLI-005',  20,  6.80, NULL, 3),
		('Sel fin 1 kg',          'SEL-FIN-006',  50,  0.60, NULL, 3),
		('Poivre noir 100 g',     'POI-NOI-007',   2,  3.20, 5,    1),
		('Levure chimique 10 g',  'LEV-CHI-008',  30,  0.80, NULL, 2)
	`)
	if err != nil {
		return fmt.Errorf("seed products : %w", err)
	}

	// ── Mouvements ───────────────────────────────────────────────────────────
	_, err = tx.Exec(`
		INSERT INTO movements (product_id, type, quantity, note, user_id) VALUES
		(1, 'IN',  20, 'Livraison initiale',       1),
		(1, 'OUT', 17, 'Vente semaine 1',           1),
		(2, 'IN',  50, 'Livraison initiale',        1),
		(2, 'OUT',  8, 'Vente',                     2),
		(3, 'IN',  20, 'Livraison initiale',        1),
		(3, 'OUT',  5, 'Vente',                     2),
		(4, 'IN',  20, 'Livraison initiale',        1),
		(4, 'OUT', 12, 'Vente semaine 1',           1),
		(5, 'IN',  25, 'Livraison initiale',        1),
		(5, 'OUT',  5, 'Vente',                     2),
		(7, 'IN',  10, 'Livraison initiale',        1),
		(7, 'OUT',  8, 'Vente — stock quasi épuisé',1)
	`)
	if err != nil {
		return fmt.Errorf("seed movements : %w", err)
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("commit seed : %w", err)
	}

	log.Println("Données de démonstration insérées avec succès.")
	log.Println("  Comptes disponibles : admin / admin123  |  employee / employee123")
	return nil
}
