package repository

import (
	"database/sql"
	"errors"
	"fmt"
	"strings"

	"github.com/KoZeuh/GO-Project-ESGI/api/internal/models"
)

// ProductRepository encapsule l'accès à la table "products".
// Le calcul du seuil d'alerte effectif (is_low_stock) est une décision
// métier et reste donc de la responsabilité de la couche service.
type ProductRepository struct {
	db dbtx
}

func NewProductRepository(db *sql.DB) *ProductRepository {
	return &ProductRepository{db: db}
}

// WithTx retourne une instance du repository exécutant ses requêtes dans la
// transaction fournie, pour composer des opérations atomiques entre repositories.
func (r *ProductRepository) WithTx(tx *sql.Tx) *ProductRepository {
	return &ProductRepository{db: tx}
}

const productSelect = `
	SELECT p.id, p.name, p.reference, p.quantity, p.price, p.alert_threshold,
	       p.supplier_id, s.name, p.created_at, p.updated_at
	FROM products p
	JOIN suppliers s ON s.id = p.supplier_id
`

func (r *ProductRepository) Create(p *models.Product) (int64, error) {
	res, err := r.db.Exec(
		`INSERT INTO products (name, reference, quantity, price, alert_threshold, supplier_id)
		 VALUES (?, ?, ?, ?, ?, ?)`,
		p.Name, p.Reference, p.Quantity, p.Price, p.AlertThreshold, p.SupplierID,
	)
	if err != nil {
		if strings.Contains(err.Error(), "UNIQUE") {
			return 0, ErrDuplicate
		}
		return 0, fmt.Errorf("création produit : %w", err)
	}
	return res.LastInsertId()
}

// FindAll retourne tous les produits, filtrés par nom/référence si search
// est non vide (recherche insensible à la casse, correspondance partielle).
func (r *ProductRepository) FindAll(search string) ([]models.Product, error) {
	query := productSelect
	args := []any{}
	if search != "" {
		query += ` WHERE p.name LIKE ? OR p.reference LIKE ?`
		like := "%" + search + "%"
		args = append(args, like, like)
	}
	query += ` ORDER BY p.name`

	rows, err := r.db.Query(query, args...)
	if err != nil {
		return nil, fmt.Errorf("liste produits : %w", err)
	}
	defer rows.Close()

	products := []models.Product{}
	for rows.Next() {
		p, err := scanProduct(rows)
		if err != nil {
			return nil, err
		}
		products = append(products, p)
	}
	return products, rows.Err()
}

func (r *ProductRepository) FindByID(id int64) (*models.Product, error) {
	row := r.db.QueryRow(productSelect+` WHERE p.id = ?`, id)
	p, err := scanProduct(row)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, ErrNotFound
	}
	if err != nil {
		return nil, err
	}
	return &p, nil
}

func (r *ProductRepository) Update(p *models.Product) error {
	res, err := r.db.Exec(
		`UPDATE products
		 SET name = ?, reference = ?, quantity = ?, price = ?, alert_threshold = ?, supplier_id = ?, updated_at = CURRENT_TIMESTAMP
		 WHERE id = ?`,
		p.Name, p.Reference, p.Quantity, p.Price, p.AlertThreshold, p.SupplierID, p.ID,
	)
	if err != nil {
		if strings.Contains(err.Error(), "UNIQUE") {
			return ErrDuplicate
		}
		return fmt.Errorf("mise à jour produit : %w", err)
	}
	return checkAffected(res)
}

func (r *ProductRepository) Delete(id int64) error {
	res, err := r.db.Exec(`DELETE FROM products WHERE id = ?`, id)
	if err != nil {
		return fmt.Errorf("suppression produit : %w", err)
	}
	return checkAffected(res)
}

// AdjustQuantity applique un delta (positif ou négatif) à la quantité d'un
// produit. La contrainte CHECK (quantity >= 0) du schéma empêche toute
// sortie de stock supérieure au stock disponible.
func (r *ProductRepository) AdjustQuantity(id int64, delta int) error {
	res, err := r.db.Exec(
		`UPDATE products SET quantity = quantity + ?, updated_at = CURRENT_TIMESTAMP WHERE id = ?`,
		delta, id,
	)
	if err != nil {
		if strings.Contains(err.Error(), "CHECK") {
			return ErrInsufficient
		}
		return fmt.Errorf("ajustement quantité produit : %w", err)
	}
	return checkAffected(res)
}

type scannable interface {
	Scan(dest ...any) error
}

func scanProduct(row scannable) (models.Product, error) {
	var p models.Product
	err := row.Scan(
		&p.ID, &p.Name, &p.Reference, &p.Quantity, &p.Price, &p.AlertThreshold,
		&p.SupplierID, &p.SupplierName, &p.CreatedAt, &p.UpdatedAt,
	)
	if err != nil {
		return models.Product{}, fmt.Errorf("lecture produit : %w", err)
	}
	return p, nil
}
