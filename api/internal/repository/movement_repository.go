package repository

import (
	"database/sql"
	"errors"
	"fmt"

	"github.com/KoZeuh/GO-Project-ESGI/api/internal/models"
)

// MovementRepository encapsule l'accès à la table "movements".
type MovementRepository struct {
	db dbtx
}

func NewMovementRepository(db *sql.DB) *MovementRepository {
	return &MovementRepository{db: db}
}

// WithTx retourne une instance du repository exécutant ses requêtes dans la
// transaction fournie, pour composer des opérations atomiques entre repositories.
func (r *MovementRepository) WithTx(tx *sql.Tx) *MovementRepository {
	return &MovementRepository{db: tx}
}

// MovementFilter permet de filtrer l'historique des mouvements.
type MovementFilter struct {
	ProductID int64               // 0 = tous les produits
	Type      models.MovementType // "" = tous les types
	From      string              // format RFC3339, "" = pas de borne basse
	To        string              // format RFC3339, "" = pas de borne haute
}

const movementSelect = `
	SELECT m.id, m.product_id, p.name, m.type, m.quantity, m.note,
	       m.user_id, u.username, m.created_at
	FROM movements m
	JOIN products p ON p.id = m.product_id
	JOIN users u ON u.id = m.user_id
`

// Create insère un mouvement de stock (sans modifier la quantité produit :
// c'est la responsabilité du service, qui orchestre les deux opérations).
func (r *MovementRepository) Create(m *models.Movement) (int64, error) {
	res, err := r.db.Exec(
		`INSERT INTO movements (product_id, type, quantity, note, user_id) VALUES (?, ?, ?, ?, ?)`,
		m.ProductID, m.Type, m.Quantity, m.Note, m.UserID,
	)
	if err != nil {
		return 0, fmt.Errorf("création mouvement : %w", err)
	}
	return res.LastInsertId()
}

// FindAll retourne l'historique des mouvements filtré, du plus récent au plus ancien.
func (r *MovementRepository) FindAll(f MovementFilter) ([]models.Movement, error) {
	query := movementSelect + " WHERE 1 = 1"
	args := []any{}

	if f.ProductID != 0 {
		query += " AND m.product_id = ?"
		args = append(args, f.ProductID)
	}
	if f.Type != "" {
		query += " AND m.type = ?"
		args = append(args, f.Type)
	}
	if f.From != "" {
		query += " AND m.created_at >= ?"
		args = append(args, f.From)
	}
	if f.To != "" {
		query += " AND m.created_at <= ?"
		args = append(args, f.To)
	}
	query += " ORDER BY m.created_at DESC"

	rows, err := r.db.Query(query, args...)
	if err != nil {
		return nil, fmt.Errorf("liste mouvements : %w", err)
	}
	defer rows.Close()

	movements := []models.Movement{}
	for rows.Next() {
		var m models.Movement
		if err := rows.Scan(&m.ID, &m.ProductID, &m.ProductName, &m.Type, &m.Quantity,
			&m.Note, &m.UserID, &m.Username, &m.CreatedAt); err != nil {
			return nil, fmt.Errorf("lecture mouvement : %w", err)
		}
		movements = append(movements, m)
	}
	return movements, rows.Err()
}

func (r *MovementRepository) FindByID(id int64) (*models.Movement, error) {
	row := r.db.QueryRow(movementSelect+" WHERE m.id = ?", id)
	var m models.Movement
	err := row.Scan(&m.ID, &m.ProductID, &m.ProductName, &m.Type, &m.Quantity,
		&m.Note, &m.UserID, &m.Username, &m.CreatedAt)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, ErrNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("lecture mouvement : %w", err)
	}
	return &m, nil
}
