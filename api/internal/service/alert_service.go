package service

import (
	"fmt"

	"github.com/KoZeuh/GO-Project-ESGI/api/internal/models"
	"github.com/KoZeuh/GO-Project-ESGI/api/internal/repository"
)

// AlertService centralise la logique métier de calcul des seuils d'alerte :
// un produit est en alerte de stock faible dès que sa quantité descend à son seuil propre (alert_threshold) s'il en a un, sinon au seuil par défaut global configuré dans alert_settings.
type AlertService struct {
	alerts   *repository.AlertRepository
	products *repository.ProductRepository
}

func NewAlertService(alerts *repository.AlertRepository, products *repository.ProductRepository) *AlertService {
	return &AlertService{alerts: alerts, products: products}
}

func (s *AlertService) GetSettings() (*models.AlertSettings, error) {
	return s.alerts.GetSettings()
}

func (s *AlertService) UpdateSettings(defaultThreshold int) (*models.AlertSettings, error) {
	return s.alerts.UpdateSettings(defaultThreshold)
}

// EffectiveThreshold retourne le seuil qui s'applique réellement à un produit.
func EffectiveThreshold(p models.Product, defaultThreshold int) int {
	if p.AlertThreshold != nil {
		return *p.AlertThreshold
	}
	return defaultThreshold
}

// IsLowStock indique si un produit est en dessous (ou à) son seuil d'alerte effectif.
func IsLowStock(p models.Product, defaultThreshold int) bool {
	return p.Quantity <= EffectiveThreshold(p, defaultThreshold)
}

// annotate calcule et positionne le champ IsLowStock sur chaque produit.
func (s *AlertService) annotate(products []models.Product) ([]models.Product, error) {
	settings, err := s.alerts.GetSettings()
	if err != nil {
		return nil, fmt.Errorf("calcul des alertes : %w", err)
	}
	for i := range products {
		products[i].IsLowStock = IsLowStock(products[i], settings.DefaultThreshold)
	}
	return products, nil
}

// AnnotateLowStock est exposée pour être réutilisée par ProductService lors du listing général des produits (chaque produit connaît son propre statut).
func (s *AlertService) AnnotateLowStock(products []models.Product) ([]models.Product, error) {
	return s.annotate(products)
}

// GetLowStockProducts retourne le seuil par défaut ainsi que la liste des produits actuellement en alerte de stock faible.
func (s *AlertService) GetLowStockProducts() (int, []models.Product, error) {
	all, err := s.products.FindAll("")
	if err != nil {
		return 0, nil, fmt.Errorf("chargement produits : %w", err)
	}
	settings, err := s.alerts.GetSettings()
	if err != nil {
		return 0, nil, fmt.Errorf("chargement réglages d'alerte : %w", err)
	}

	lowStock := []models.Product{}
	for _, p := range all {
		p.IsLowStock = IsLowStock(p, settings.DefaultThreshold)
		if p.IsLowStock {
			lowStock = append(lowStock, p)
		}
	}
	return settings.DefaultThreshold, lowStock, nil
}
