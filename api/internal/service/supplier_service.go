package service

import (
	"github.com/KoZeuh/GO-Project-ESGI/api/internal/dto"
	"github.com/KoZeuh/GO-Project-ESGI/api/internal/models"
	"github.com/KoZeuh/GO-Project-ESGI/api/internal/repository"
)

// SupplierService porte les règles de gestion des fournisseurs. Le CRUD est simple : aucune règle métier supplémentaire n'est requise par le cahier des charges.
type SupplierService struct {
	suppliers *repository.SupplierRepository
}

func NewSupplierService(suppliers *repository.SupplierRepository) *SupplierService {
	return &SupplierService{suppliers: suppliers}
}

func (s *SupplierService) List() ([]models.Supplier, error) {
	return s.suppliers.FindAll()
}

func (s *SupplierService) Get(id int64) (*models.Supplier, error) {
	return s.suppliers.FindByID(id)
}

func (s *SupplierService) Create(req dto.SupplierRequest) (*models.Supplier, error) {
	supplier := &models.Supplier{Name: req.Name, Email: req.Email, Phone: req.Phone, Address: req.Address}
	id, err := s.suppliers.Create(supplier)
	if err != nil {
		return nil, err
	}
	return s.suppliers.FindByID(id)
}

func (s *SupplierService) Update(id int64, req dto.SupplierRequest) (*models.Supplier, error) {
	supplier := &models.Supplier{ID: id, Name: req.Name, Email: req.Email, Phone: req.Phone, Address: req.Address}
	if err := s.suppliers.Update(supplier); err != nil {
		return nil, err
	}
	return s.suppliers.FindByID(id)
}

func (s *SupplierService) Delete(id int64) error {
	return s.suppliers.Delete(id)
}
