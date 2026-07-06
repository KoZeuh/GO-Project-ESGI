package service

import (
	"testing"

	"github.com/KoZeuh/GO-Project-ESGI/api/internal/models"
)

func intPtr(v int) *int { return &v }

func TestEffectiveThreshold(t *testing.T) {
	tests := []struct {
		name             string
		product          models.Product
		defaultThreshold int
		want             int
	}{
		{"seuil propre défini", models.Product{AlertThreshold: intPtr(3)}, 10, 3},
		{"pas de seuil propre : seuil par défaut", models.Product{AlertThreshold: nil}, 10, 10},
		{"seuil propre à zéro reste respecté", models.Product{AlertThreshold: intPtr(0)}, 10, 0},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := EffectiveThreshold(tt.product, tt.defaultThreshold)
			if got != tt.want {
				t.Errorf("EffectiveThreshold() = %d, want %d", got, tt.want)
			}
		})
	}
}

func TestIsLowStock(t *testing.T) {
	tests := []struct {
		name             string
		product          models.Product
		defaultThreshold int
		want             bool
	}{
		{"quantité au-dessus du seuil", models.Product{Quantity: 20, AlertThreshold: intPtr(5)}, 10, false},
		{"quantité égale au seuil déclenche l'alerte", models.Product{Quantity: 5, AlertThreshold: intPtr(5)}, 10, true},
		{"quantité en dessous du seuil déclenche l'alerte", models.Product{Quantity: 2, AlertThreshold: intPtr(5)}, 10, true},
		{"utilise le seuil par défaut si absent", models.Product{Quantity: 8, AlertThreshold: nil}, 10, true},
		{"quantité nulle est toujours en alerte", models.Product{Quantity: 0, AlertThreshold: intPtr(0)}, 10, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := IsLowStock(tt.product, tt.defaultThreshold)
			if got != tt.want {
				t.Errorf("IsLowStock() = %v, want %v", got, tt.want)
			}
		})
	}
}
