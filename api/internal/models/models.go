// Package models définit les structures de données (entités)

package models

import "time"

type User struct {
	ID           int64     `json:"id"`
	Username     string    `json:"username"`
	PasswordHash string    `json:"-"`
	Role         string    `json:"role"`
	CreatedAt    time.Time `json:"created_at"`
}

type Supplier struct {
	ID        int64     `json:"id"`
	Name      string    `json:"name"`
	Email     string    `json:"email"`
	Phone     string    `json:"phone"`
	Address   string    `json:"address"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

type Product struct {
	ID             int64     `json:"id"`
	Name           string    `json:"name"`
	Reference      string    `json:"reference"`
	Quantity       int       `json:"quantity"`
	Price          float64   `json:"price"`
	AlertThreshold *int      `json:"alert_threshold"`
	SupplierID     int64     `json:"supplier_id"`
	SupplierName   string    `json:"supplier_name,omitempty"`
	IsLowStock     bool      `json:"is_low_stock"`
	CreatedAt      time.Time `json:"created_at"`
	UpdatedAt      time.Time `json:"updated_at"`
}

type MovementType string

const (
	MovementIn  MovementType = "IN"
	MovementOut MovementType = "OUT"
)

type Movement struct {
	ID          int64        `json:"id"`
	ProductID   int64        `json:"product_id"`
	ProductName string       `json:"product_name,omitempty"`
	Type        MovementType `json:"type"`
	Quantity    int          `json:"quantity"`
	Note        string       `json:"note"`
	UserID      int64        `json:"user_id"`
	Username    string       `json:"username,omitempty"`
	CreatedAt   time.Time    `json:"created_at"`
}

type AlertSettings struct {
	ID               int64     `json:"id"`
	DefaultThreshold int       `json:"default_threshold"`
	UpdatedAt        time.Time `json:"updated_at"`
}
