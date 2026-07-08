// === [PARTIE TUI - Frontend] ===
// Package store gère le cache local en mémoire et la persistance JSON sur disque.
package store

import (
	"encoding/json"
	"os"
	"time"

	"github.com/KoZeuh/GO-Project-ESGI/tui/internal/client"
)

// Store est le cache local de l'application TUI.
// Il peut être persisté sur disque (export JSON local) ou rechargé depuis l'API.
type Store struct {
	Products   []client.Product   `json:"products"`
	Suppliers  []client.Supplier  `json:"suppliers"`
	Movements  []client.Movement  `json:"movements"`
	ExportedAt time.Time          `json:"exported_at"`
	Loaded     bool               `json:"-"`
}

// New retourne un Store vide.
func New() *Store {
	return &Store{}
}

// ImportFromExport remplace le cache local par les données de l'export API.
func (s *Store) ImportFromExport(resp *client.ExportResponse) {
	s.Products = resp.Products
	s.Suppliers = resp.Suppliers
	s.Movements = resp.Movements
	s.ExportedAt = resp.ExportedAt
	s.Loaded = true
}

// SaveToFile sérialise le cache local vers un fichier JSON.
func (s *Store) SaveToFile(path string) error {
	f, err := os.Create(path)
	if err != nil {
		return err
	}
	defer f.Close()
	enc := json.NewEncoder(f)
	enc.SetIndent("", "  ")
	return enc.Encode(s)
}

// LoadFromFile charge le cache local depuis un fichier JSON existant.
func (s *Store) LoadFromFile(path string) error {
	f, err := os.Open(path)
	if err != nil {
		return err
	}
	defer f.Close()
	if err := json.NewDecoder(f).Decode(s); err != nil {
		return err
	}
	s.Loaded = true
	return nil
}

// LowStockCount retourne le nombre de produits en alerte de stock faible.
func (s *Store) LowStockCount() int {
	n := 0
	for _, p := range s.Products {
		if p.IsLowStock {
			n++
		}
	}
	return n
}
