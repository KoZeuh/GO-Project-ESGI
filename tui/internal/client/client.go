// === [PARTIE TUI - Frontend] ===
// Package client fournit l'interface Client et son implémentation HTTP
// pour communiquer avec l'API REST du gestionnaire de stock,
// ainsi qu'un MockClient pour le développement hors-ligne.
package client

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strconv"
	"time"
)

// ─── Types miroirs des modèles API ────────────────────────────────────────────

type User struct {
	ID        int64     `json:"id"`
	Username  string    `json:"username"`
	CreatedAt time.Time `json:"created_at"`
}

type LoginResponse struct {
	Token     string    `json:"token"`
	ExpiresAt time.Time `json:"expires_at"`
	User      User      `json:"user"`
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

type ProductRequest struct {
	Name           string  `json:"name"`
	Reference      string  `json:"reference"`
	Quantity       int     `json:"quantity"`
	Price          float64 `json:"price"`
	AlertThreshold *int    `json:"alert_threshold,omitempty"`
	SupplierID     int64   `json:"supplier_id"`
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

type SupplierRequest struct {
	Name    string `json:"name"`
	Email   string `json:"email,omitempty"`
	Phone   string `json:"phone,omitempty"`
	Address string `json:"address,omitempty"`
}

type Movement struct {
	ID          int64     `json:"id"`
	ProductID   int64     `json:"product_id"`
	ProductName string    `json:"product_name,omitempty"`
	Type        string    `json:"type"`
	Quantity    int       `json:"quantity"`
	Note        string    `json:"note"`
	UserID      int64     `json:"user_id"`
	Username    string    `json:"username,omitempty"`
	CreatedAt   time.Time `json:"created_at"`
}

type MovementRequest struct {
	ProductID int64  `json:"product_id"`
	Type      string `json:"type"`
	Quantity  int    `json:"quantity"`
	Note      string `json:"note,omitempty"`
}

type AlertSettings struct {
	ID               int64     `json:"id"`
	DefaultThreshold int       `json:"default_threshold"`
	UpdatedAt        time.Time `json:"updated_at"`
}

type AlertsResponse struct {
	DefaultThreshold int       `json:"default_threshold"`
	LowStockProducts []Product `json:"low_stock_products"`
}

type AlertSettingsRequest struct {
	DefaultThreshold int `json:"default_threshold"`
}

type ExportResponse struct {
	ExportedAt time.Time  `json:"exported_at"`
	Products   []Product  `json:"products"`
	Suppliers  []Supplier `json:"suppliers"`
	Movements  []Movement `json:"movements"`
}

// ─── Interface Client ─────────────────────────────────────────────────────────

// Client est l'interface que toute implémentation doit satisfaire.
type Client interface {
	SetToken(token string)

	Login(username, password string) (*LoginResponse, error)
	Register(username, password string) (*User, error)

	GetProducts(search string) ([]Product, error)
	GetProduct(id int64) (*Product, error)
	CreateProduct(req ProductRequest) (*Product, error)
	UpdateProduct(id int64, req ProductRequest) (*Product, error)
	DeleteProduct(id int64) error

	GetSuppliers() ([]Supplier, error)
	GetSupplier(id int64) (*Supplier, error)
	CreateSupplier(req SupplierRequest) (*Supplier, error)
	UpdateSupplier(id int64, req SupplierRequest) (*Supplier, error)
	DeleteSupplier(id int64) error

	GetMovements(productID int64, movType, from, to string) ([]Movement, error)
	CreateMovement(req MovementRequest) (*Movement, error)

	GetAlerts() (*AlertsResponse, error)
	UpdateAlertSettings(req AlertSettingsRequest) (*AlertSettings, error)

	GetExport() (*ExportResponse, error)
}

// ─── HTTPClient ───────────────────────────────────────────────────────────────

// HTTPClient est l'implémentation réelle qui appelle l'API REST.
type HTTPClient struct {
	baseURL string
	token   string
	http    *http.Client
}

// NewHTTPClient crée un HTTPClient pointant vers baseURL.
func NewHTTPClient(baseURL string) *HTTPClient {
	return &HTTPClient{
		baseURL: baseURL,
		http:    &http.Client{Timeout: 15 * time.Second},
	}
}

func (c *HTTPClient) SetToken(token string) { c.token = token }

func (c *HTTPClient) do(method, path string, body, out interface{}) error {
	var bodyReader io.Reader
	if body != nil {
		data, err := json.Marshal(body)
		if err != nil {
			return err
		}
		bodyReader = bytes.NewReader(data)
	}

	req, err := http.NewRequest(method, c.baseURL+path, bodyReader)
	if err != nil {
		return err
	}
	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}
	if c.token != "" {
		req.Header.Set("Authorization", "Bearer "+c.token)
	}

	resp, err := c.http.Do(req)
	if err != nil {
		return fmt.Errorf("connexion à l'API impossible: %w", err)
	}
	defer resp.Body.Close()

	data, err := io.ReadAll(resp.Body)
	if err != nil {
		return err
	}

	if resp.StatusCode >= 400 {
		var e struct {
			Error string `json:"error"`
		}
		if json.Unmarshal(data, &e) == nil && e.Error != "" {
			return fmt.Errorf("%s", e.Error)
		}
		return fmt.Errorf("erreur HTTP %d", resp.StatusCode)
	}

	if out != nil && resp.StatusCode != http.StatusNoContent && len(data) > 0 {
		return json.Unmarshal(data, out)
	}
	return nil
}

func (c *HTTPClient) Login(username, password string) (*LoginResponse, error) {
	var out LoginResponse
	err := c.do("POST", "/api/v1/auth/login", map[string]string{
		"username": username, "password": password,
	}, &out)
	if err != nil {
		return nil, err
	}
	return &out, nil
}

func (c *HTTPClient) Register(username, password string) (*User, error) {
	var out User
	err := c.do("POST", "/api/v1/auth/register", map[string]string{
		"username": username, "password": password,
	}, &out)
	if err != nil {
		return nil, err
	}
	return &out, nil
}

func (c *HTTPClient) GetProducts(search string) ([]Product, error) {
	path := "/api/v1/products"
	if search != "" {
		path += "?search=" + url.QueryEscape(search)
	}
	var out []Product
	return out, c.do("GET", path, nil, &out)
}

func (c *HTTPClient) GetProduct(id int64) (*Product, error) {
	var out Product
	return &out, c.do("GET", fmt.Sprintf("/api/v1/products/%d", id), nil, &out)
}

func (c *HTTPClient) CreateProduct(req ProductRequest) (*Product, error) {
	var out Product
	return &out, c.do("POST", "/api/v1/products", req, &out)
}

func (c *HTTPClient) UpdateProduct(id int64, req ProductRequest) (*Product, error) {
	var out Product
	return &out, c.do("PUT", fmt.Sprintf("/api/v1/products/%d", id), req, &out)
}

func (c *HTTPClient) DeleteProduct(id int64) error {
	return c.do("DELETE", fmt.Sprintf("/api/v1/products/%d", id), nil, nil)
}

func (c *HTTPClient) GetSuppliers() ([]Supplier, error) {
	var out []Supplier
	return out, c.do("GET", "/api/v1/suppliers", nil, &out)
}

func (c *HTTPClient) GetSupplier(id int64) (*Supplier, error) {
	var out Supplier
	return &out, c.do("GET", fmt.Sprintf("/api/v1/suppliers/%d", id), nil, &out)
}

func (c *HTTPClient) CreateSupplier(req SupplierRequest) (*Supplier, error) {
	var out Supplier
	return &out, c.do("POST", "/api/v1/suppliers", req, &out)
}

func (c *HTTPClient) UpdateSupplier(id int64, req SupplierRequest) (*Supplier, error) {
	var out Supplier
	return &out, c.do("PUT", fmt.Sprintf("/api/v1/suppliers/%d", id), req, &out)
}

func (c *HTTPClient) DeleteSupplier(id int64) error {
	return c.do("DELETE", fmt.Sprintf("/api/v1/suppliers/%d", id), nil, nil)
}

func (c *HTTPClient) GetMovements(productID int64, movType, from, to string) ([]Movement, error) {
	params := url.Values{}
	if productID > 0 {
		params.Set("product_id", strconv.FormatInt(productID, 10))
	}
	if movType != "" {
		params.Set("type", movType)
	}
	if from != "" {
		params.Set("from", from)
	}
	if to != "" {
		params.Set("to", to)
	}
	path := "/api/v1/movements"
	if len(params) > 0 {
		path += "?" + params.Encode()
	}
	var out []Movement
	return out, c.do("GET", path, nil, &out)
}

func (c *HTTPClient) CreateMovement(req MovementRequest) (*Movement, error) {
	var out Movement
	return &out, c.do("POST", "/api/v1/movements", req, &out)
}

func (c *HTTPClient) GetAlerts() (*AlertsResponse, error) {
	var out AlertsResponse
	return &out, c.do("GET", "/api/v1/alerts", nil, &out)
}

func (c *HTTPClient) UpdateAlertSettings(req AlertSettingsRequest) (*AlertSettings, error) {
	var out AlertSettings
	return &out, c.do("PUT", "/api/v1/alerts/settings", req, &out)
}

func (c *HTTPClient) GetExport() (*ExportResponse, error) {
	var out ExportResponse
	return &out, c.do("GET", "/api/v1/export", nil, &out)
}

// ─── MockClient ───────────────────────────────────────────────────────────────

// MockClient retourne des données statiques pour le développement hors-ligne.
// Utilisez --mock au démarrage pour l'activer.
type MockClient struct{ token string }

func NewMockClient() *MockClient { return &MockClient{} }
func (m *MockClient) SetToken(t string) { m.token = t }

var (
	mockThreshold = 5
	mockProducts  = []Product{
		{ID: 1, Name: "Café en grains 1 kg", Reference: "CAF-1KG-001", Quantity: 3, Price: 12.50, AlertThreshold: &mockThreshold, SupplierID: 1, SupplierName: "Dupont SA", IsLowStock: true},
		{ID: 2, Name: "Thé vert 500 g", Reference: "THE-500G-002", Quantity: 42, Price: 8.00, SupplierID: 1, SupplierName: "Dupont SA", IsLowStock: false},
		{ID: 3, Name: "Sucre blanc 1 kg", Reference: "SUC-1KG-003", Quantity: 15, Price: 1.50, SupplierID: 2, SupplierName: "BioFournisseur", IsLowStock: false},
	}
	mockSuppliers = []Supplier{
		{ID: 1, Name: "Dupont SA", Email: "contact@dupont.fr", Phone: "0102030405", Address: "1 rue de la Paix, Paris"},
		{ID: 2, Name: "BioFournisseur", Email: "bio@fourni.fr", Phone: "0607080910"},
	}
	mockMovements = []Movement{
		{ID: 1, ProductID: 1, ProductName: "Café en grains 1 kg", Type: "IN", Quantity: 10, Note: "Livraison initiale", UserID: 1, Username: "admin"},
		{ID: 2, ProductID: 1, ProductName: "Café en grains 1 kg", Type: "OUT", Quantity: 7, Note: "Vente", UserID: 1, Username: "admin"},
	}
)

func (m *MockClient) Login(username, password string) (*LoginResponse, error) {
	return &LoginResponse{Token: "mock-token", User: User{ID: 1, Username: username}}, nil
}
func (m *MockClient) Register(username, password string) (*User, error) {
	return &User{ID: 2, Username: username}, nil
}
func (m *MockClient) GetProducts(_ string) ([]Product, error) { return mockProducts, nil }
func (m *MockClient) GetProduct(id int64) (*Product, error) {
	for _, p := range mockProducts {
		if p.ID == id {
			cp := p
			return &cp, nil
		}
	}
	return nil, fmt.Errorf("produit introuvable")
}
func (m *MockClient) CreateProduct(req ProductRequest) (*Product, error) {
	return &Product{ID: 99, Name: req.Name, Reference: req.Reference, Quantity: req.Quantity, Price: req.Price}, nil
}
func (m *MockClient) UpdateProduct(id int64, req ProductRequest) (*Product, error) {
	return &Product{ID: id, Name: req.Name, Reference: req.Reference, Quantity: req.Quantity, Price: req.Price}, nil
}
func (m *MockClient) DeleteProduct(_ int64) error { return nil }

func (m *MockClient) GetSuppliers() ([]Supplier, error) { return mockSuppliers, nil }
func (m *MockClient) GetSupplier(id int64) (*Supplier, error) {
	for _, s := range mockSuppliers {
		if s.ID == id {
			cp := s
			return &cp, nil
		}
	}
	return nil, fmt.Errorf("fournisseur introuvable")
}
func (m *MockClient) CreateSupplier(req SupplierRequest) (*Supplier, error) {
	return &Supplier{ID: 99, Name: req.Name, Email: req.Email, Phone: req.Phone, Address: req.Address}, nil
}
func (m *MockClient) UpdateSupplier(id int64, req SupplierRequest) (*Supplier, error) {
	return &Supplier{ID: id, Name: req.Name, Email: req.Email, Phone: req.Phone, Address: req.Address}, nil
}
func (m *MockClient) DeleteSupplier(_ int64) error { return nil }

func (m *MockClient) GetMovements(_ int64, _, _, _ string) ([]Movement, error) {
	return mockMovements, nil
}
func (m *MockClient) CreateMovement(req MovementRequest) (*Movement, error) {
	return &Movement{ID: 99, ProductID: req.ProductID, Type: req.Type, Quantity: req.Quantity, Note: req.Note}, nil
}
func (m *MockClient) GetAlerts() (*AlertsResponse, error) {
	return &AlertsResponse{DefaultThreshold: 5, LowStockProducts: []Product{mockProducts[0]}}, nil
}
func (m *MockClient) UpdateAlertSettings(req AlertSettingsRequest) (*AlertSettings, error) {
	return &AlertSettings{ID: 1, DefaultThreshold: req.DefaultThreshold}, nil
}
func (m *MockClient) GetExport() (*ExportResponse, error) {
	return &ExportResponse{ExportedAt: time.Now(), Products: mockProducts, Suppliers: mockSuppliers, Movements: mockMovements}, nil
}
