package service

import (
	"errors"
	"fmt"

	"github.com/KoZeuh/GO-Project-ESGI/api/internal/dto"
	"github.com/KoZeuh/GO-Project-ESGI/api/internal/models"
	"github.com/KoZeuh/GO-Project-ESGI/api/internal/repository"
)

// ProductService porte la logique métier liée aux produits : validation du fournisseur associé et calcul du statut d'alerte de stock faible.
type ProductService struct {
	products  *repository.ProductRepository
	suppliers *repository.SupplierRepository
	alerts    *AlertService
}

func NewProductService(products *repository.ProductRepository, suppliers *repository.SupplierRepository, alerts *AlertService) *ProductService {
	return &ProductService{products: products, suppliers: suppliers, alerts: alerts}
}

func (s *ProductService) List(search string) ([]models.Product, error) {
	products, err := s.products.FindAll(search)
	if err != nil {
		return nil, err
	}
	return s.alerts.AnnotateLowStock(products)
}

func (s *ProductService) Get(id int64) (*models.Product, error) {
	p, err := s.products.FindByID(id)
	if err != nil {
		return nil, err
	}
	annotated, err := s.alerts.AnnotateLowStock([]models.Product{*p})
	if err != nil {
		return nil, err
	}
	return &annotated[0], nil
}

func (s *ProductService) Create(req dto.ProductRequest) (*models.Product, error) {
	if err := s.ensureSupplierExists(req.SupplierID); err != nil {
		return nil, err
	}

	product := &models.Product{
		Name:           req.Name,
		Reference:      req.Reference,
		Quantity:       req.Quantity,
		Price:          req.Price,
		AlertThreshold: req.AlertThreshold,
		SupplierID:     req.SupplierID,
	}
	id, err := s.products.Create(product)
	if err != nil {
		return nil, err
	}
	return s.Get(id)
}

func (s *ProductService) Update(id int64, req dto.ProductRequest) (*models.Product, error) {
	if err := s.ensureSupplierExists(req.SupplierID); err != nil {
		return nil, err
	}

	product := &models.Product{
		ID:             id,
		Name:           req.Name,
		Reference:      req.Reference,
		Quantity:       req.Quantity,
		Price:          req.Price,
		AlertThreshold: req.AlertThreshold,
		SupplierID:     req.SupplierID,
	}
	if err := s.products.Update(product); err != nil {
		return nil, err
	}
	return s.Get(id)
}

func (s *ProductService) Delete(id int64) error {
	return s.products.Delete(id)
}

func (s *ProductService) ensureSupplierExists(supplierID int64) error {
	_, err := s.suppliers.FindByID(supplierID)
	if errors.Is(err, repository.ErrNotFound) {
		return ErrSupplierNotFound
	}
	if err != nil {
		return fmt.Errorf("vérification fournisseur : %w", err)
	}
	return nil
}
