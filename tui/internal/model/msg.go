// === [PARTIE TUI - Frontend] ===
// Messages Bubble Tea partagés entre les écrans du TUI.
package model

import "github.com/KoZeuh/GO-Project-ESGI/tui/internal/client"

// ─── Navigation ───────────────────────────────────────────────────────────────

// Mode représente l'écran actif.
type Mode int

const (
	ModeLogin Mode = iota
	ModeDashboard
	ModeProducts
	ModeProductForm
	ModeSuppliers
	ModeSupplierForm
	ModeMovements
	ModeRestock
	ModeAlerts
)

// FormMode distingue la création de l'édition.
type FormMode int

const (
	FormCreate FormMode = iota
	FormEdit
)

// ─── Messages ─────────────────────────────────────────────────────────────────

// ErrMsg transporte une erreur vers le modèle racine.
type ErrMsg struct{ Err error }

func (e ErrMsg) Error() string { return e.Err.Error() }

// Authentification
type LoginSuccessMsg struct {
	Token    string
	Username string
}

// Produits
type ProductsLoadedMsg struct{ Products []client.Product }
type ProductSavedMsg struct{ Product client.Product }
type ProductDeletedMsg struct{ ID int64 }

// Fournisseurs
type SuppliersLoadedMsg struct{ Suppliers []client.Supplier }
type SupplierSavedMsg struct{ Supplier client.Supplier }
type SupplierDeletedMsg struct{ ID int64 }

// Mouvements
type MovementsLoadedMsg struct{ Movements []client.Movement }
type MovementCreatedMsg struct{ Movement client.Movement }

// Alertes
type AlertsLoadedMsg struct{ Response client.AlertsResponse }
type AlertSettingsSavedMsg struct{ Settings client.AlertSettings }

// Export
type ExportDoneMsg struct{ Path string }
