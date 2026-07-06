package dto

import (
	"time"

	"github.com/KoZeuh/GO-Project-ESGI/api/internal/models"
)

// --- Erreurs ---

// ErrorResponse est le format uniforme de toute réponse d'erreur de l'API
type ErrorResponse struct {
	Error string `json:"error"`
}

// --- Authentification ---

type RegisterRequest struct {
	Username string `json:"username" binding:"required,min=3,max=50"`
	Password string `json:"password" binding:"required,min=6"`
}

type LoginRequest struct {
	Username string `json:"username" binding:"required"`
	Password string `json:"password" binding:"required"`
}

type UserResponse struct {
	ID        int64     `json:"id"`
	Username  string    `json:"username"`
	Role      string    `json:"role"`
	CreatedAt time.Time `json:"created_at"`
}

type LoginResponse struct {
	Token     string       `json:"token"`
	ExpiresAt time.Time    `json:"expires_at"`
	User      UserResponse `json:"user"`
}

func NewUserResponse(u *models.User) UserResponse {
	return UserResponse{ID: u.ID, Username: u.Username, Role: u.Role, CreatedAt: u.CreatedAt}
}

// --- Produits ---

// ProductRequest est le payload d'entrée pour la création et la mise à jour
// d'un produit. AlertThreshold est un pointeur : nil signifie "utiliser le
// seuil par défaut global" (voir logique métier dans internal/service/alert_service.go).
type ProductRequest struct {
	Name           string  `json:"name" binding:"required"`
	Reference      string  `json:"reference" binding:"required"`
	Quantity       int     `json:"quantity" binding:"gte=0"`
	Price          float64 `json:"price" binding:"gte=0"`
	AlertThreshold *int    `json:"alert_threshold"`
	SupplierID     int64   `json:"supplier_id" binding:"required"`
}

// --- Fournisseurs ---

type SupplierRequest struct {
	Name    string `json:"name" binding:"required"`
	Email   string `json:"email" binding:"omitempty,email"`
	Phone   string `json:"phone"`
	Address string `json:"address"`
}

// --- Mouvements de stock ---

type MovementRequest struct {
	ProductID int64  `json:"product_id" binding:"required"`
	Type      string `json:"type" binding:"required,oneof=IN OUT"`
	Quantity  int    `json:"quantity" binding:"required,gt=0"`
	Note      string `json:"note"`
}

// --- Alertes ---

type AlertSettingsRequest struct {
	DefaultThreshold int `json:"default_threshold" binding:"gte=0"`
}

type AlertsResponse struct {
	DefaultThreshold int              `json:"default_threshold"`
	LowStockProducts []models.Product `json:"low_stock_products"`
}

// --- Export ---

type ExportResponse struct {
	ExportedAt time.Time         `json:"exported_at"`
	Products   []models.Product  `json:"products"`
	Suppliers  []models.Supplier `json:"suppliers"`
	Movements  []models.Movement `json:"movements"`
}
