// === [PARTIE TUI - Frontend] ===
// Package styles définit les styles Lip Gloss partagés entre tous les écrans.
package styles

import "github.com/charmbracelet/lipgloss"

// ─── Palette de couleurs ──────────────────────────────────────────────────────

var (
	ColorPrimary  = lipgloss.Color("#7C3AED")
	ColorSuccess  = lipgloss.Color("#10B981")
	ColorWarning  = lipgloss.Color("#F59E0B")
	ColorDanger   = lipgloss.Color("#EF4444")
	ColorMuted    = lipgloss.Color("#6B7280")
	ColorText     = lipgloss.Color("#F9FAFB")
	ColorSubtext  = lipgloss.Color("#9CA3AF")
	ColorBorder   = lipgloss.Color("#374151")
	ColorSelected = lipgloss.Color("#4C1D95")
)

// ─── Styles communs ───────────────────────────────────────────────────────────

var (
	AppTitle = lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("#FFFFFF")).
			Background(ColorPrimary).
			Padding(0, 2)

	Title = lipgloss.NewStyle().
		Bold(true).
		Foreground(ColorPrimary).
		MarginBottom(1)

	Subtitle = lipgloss.NewStyle().
			Foreground(ColorSubtext)

	Panel = lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(ColorBorder).
		Padding(0, 2)

	FocusedPanel = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(ColorPrimary).
			Padding(0, 2)

	ErrorStyle = lipgloss.NewStyle().
			Foreground(ColorDanger).
			Bold(true)

	SuccessStyle = lipgloss.NewStyle().
			Foreground(ColorSuccess).
			Bold(true)

	WarningStyle = lipgloss.NewStyle().
			Foreground(ColorWarning).
			Bold(true)

	MutedStyle = lipgloss.NewStyle().
			Foreground(ColorMuted)

	BoldStyle = lipgloss.NewStyle().Bold(true)

	NavBar = lipgloss.NewStyle().
		Foreground(ColorMuted).
		MarginTop(1)

	NavKey = lipgloss.NewStyle().
		Foreground(ColorPrimary).
		Bold(true)

	InputLabel = lipgloss.NewStyle().
			Foreground(ColorSubtext).
			Width(22)

	FocusedLabel = lipgloss.NewStyle().
			Foreground(ColorPrimary).
			Bold(true).
			Width(22)

	SelectedItem = lipgloss.NewStyle().
			Foreground(ColorText).
			Background(ColorSelected).
			Bold(true).
			Padding(0, 1)

	NormalItem = lipgloss.NewStyle().
			Foreground(ColorText).
			Padding(0, 1)

	LowStockBadge = lipgloss.NewStyle().
			Foreground(ColorDanger).
			Bold(true)

	OkBadge = lipgloss.NewStyle().
		Foreground(ColorSuccess)

	INBadge = lipgloss.NewStyle().
		Foreground(ColorSuccess).
		Bold(true)

	OUTBadge = lipgloss.NewStyle().
		Foreground(ColorDanger).
		Bold(true)
)

// ─── Fonctions utilitaires ───────────────────────────────────────────────────

// Key retourne un raccourci clavier stylisé : [k].
func Key(k string) string {
	return NavKey.Render("["+k+"]")
}

// StockBadge retourne un badge coloré selon l'état de stock.
func StockBadge(lowStock bool) string {
	if lowStock {
		return LowStockBadge.Render("⚠ Stock faible")
	}
	return OkBadge.Render("✓ OK")
}

// MovTypeBadge retourne un badge IN/OUT coloré.
func MovTypeBadge(t string) string {
	if t == "IN" {
		return INBadge.Render("▲ IN ")
	}
	return OUTBadge.Render("▼ OUT")
}
