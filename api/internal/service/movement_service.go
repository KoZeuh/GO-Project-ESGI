package service

import (
	"database/sql"
	"fmt"

	"github.com/KoZeuh/GO-Project-ESGI/api/internal/dto"
	"github.com/KoZeuh/GO-Project-ESGI/api/internal/models"
	"github.com/KoZeuh/GO-Project-ESGI/api/internal/repository"
)

// MovementService orchestre la création d'un mouvement de stock : il doit à la fois enregistrer l'historique et ajuster la quantité du produit
// concerné, de façon atomique (les deux opérations réussissent ou échouent ensemble).
type MovementService struct {
	db        *sql.DB
	products  *repository.ProductRepository
	movements *repository.MovementRepository
}

func NewMovementService(db *sql.DB, products *repository.ProductRepository, movements *repository.MovementRepository) *MovementService {
	return &MovementService{db: db, products: products, movements: movements}
}

func (s *MovementService) List(filter repository.MovementFilter) ([]models.Movement, error) {
	return s.movements.FindAll(filter)
}

// Create enregistre un mouvement (entrée ou sortie) et met à jour la quantité du produit dans une même transaction SQL. Une sortie supérieure
// au stock disponible est rejetée grâce à la contrainte CHECK (quantity >= 0).
func (s *MovementService) Create(req dto.MovementRequest, userID int64) (*models.Movement, error) {
	delta := req.Quantity
	if models.MovementType(req.Type) == models.MovementOut {
		delta = -req.Quantity
	}

	tx, err := s.db.Begin()
	if err != nil {
		return nil, fmt.Errorf("démarrage transaction : %w", err)
	}
	defer tx.Rollback()

	if err := s.products.WithTx(tx).AdjustQuantity(req.ProductID, delta); err != nil {
		return nil, err
	}

	movement := &models.Movement{
		ProductID: req.ProductID,
		Type:      models.MovementType(req.Type),
		Quantity:  req.Quantity,
		Note:      req.Note,
		UserID:    userID,
	}
	id, err := s.movements.WithTx(tx).Create(movement)
	if err != nil {
		return nil, fmt.Errorf("création mouvement : %w", err)
	}

	if err := tx.Commit(); err != nil {
		return nil, fmt.Errorf("validation transaction : %w", err)
	}

	return s.movements.FindByID(id)
}
