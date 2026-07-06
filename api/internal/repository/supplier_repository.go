package repository

import (
	"database/sql"
	"errors"
	"fmt"

	"github.com/KoZeuh/GO-Project-ESGI/api/internal/models"
)

// SupplierRepository encapsule l'accès à la table "suppliers".
type SupplierRepository struct {
	db *sql.DB
}

func NewSupplierRepository(db *sql.DB) *SupplierRepository {
	return &SupplierRepository{db: db}
}

func (r *SupplierRepository) Create(s *models.Supplier) (int64, error) {
	res, err := r.db.Exec(
		`INSERT INTO suppliers (name, email, phone, address) VALUES (?, ?, ?, ?)`,
		s.Name, s.Email, s.Phone, s.Address,
	)
	if err != nil {
		return 0, fmt.Errorf("création fournisseur : %w", err)
	}
	return res.LastInsertId()
}

func (r *SupplierRepository) FindAll() ([]models.Supplier, error) {
	rows, err := r.db.Query(
		`SELECT id, name, email, phone, address, created_at, updated_at FROM suppliers ORDER BY name`,
	)
	if err != nil {
		return nil, fmt.Errorf("liste fournisseurs : %w", err)
	}
	defer rows.Close()

	suppliers := []models.Supplier{}
	for rows.Next() {
		var s models.Supplier
		if err := rows.Scan(&s.ID, &s.Name, &s.Email, &s.Phone, &s.Address, &s.CreatedAt, &s.UpdatedAt); err != nil {
			return nil, fmt.Errorf("lecture fournisseur : %w", err)
		}
		suppliers = append(suppliers, s)
	}
	return suppliers, rows.Err()
}

func (r *SupplierRepository) FindByID(id int64) (*models.Supplier, error) {
	row := r.db.QueryRow(
		`SELECT id, name, email, phone, address, created_at, updated_at FROM suppliers WHERE id = ?`,
		id,
	)
	var s models.Supplier
	err := row.Scan(&s.ID, &s.Name, &s.Email, &s.Phone, &s.Address, &s.CreatedAt, &s.UpdatedAt)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, ErrNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("lecture fournisseur : %w", err)
	}
	return &s, nil
}

func (r *SupplierRepository) Update(s *models.Supplier) error {
	res, err := r.db.Exec(
		`UPDATE suppliers SET name = ?, email = ?, phone = ?, address = ?, updated_at = CURRENT_TIMESTAMP WHERE id = ?`,
		s.Name, s.Email, s.Phone, s.Address, s.ID,
	)
	if err != nil {
		return fmt.Errorf("mise à jour fournisseur : %w", err)
	}
	return checkAffected(res)
}

func (r *SupplierRepository) Delete(id int64) error {
	res, err := r.db.Exec(`DELETE FROM suppliers WHERE id = ?`, id)
	if err != nil {
		return fmt.Errorf("suppression fournisseur : %w", err)
	}
	return checkAffected(res)
}

func checkAffected(res sql.Result) error {
	n, err := res.RowsAffected()
	if err != nil {
		return fmt.Errorf("vérification de l'opération : %w", err)
	}
	if n == 0 {
		return ErrNotFound
	}
	return nil
}
