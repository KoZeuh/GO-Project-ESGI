// === [PARTIE TUI - Frontend] ===
// Écran Dashboard : vue d'ensemble du stock, alertes actives, navigation principale.
package model

import (
	"fmt"
	"strings"

	"github.com/KoZeuh/GO-Project-ESGI/tui/internal/client"
	"github.com/KoZeuh/GO-Project-ESGI/tui/internal/styles"
	tea "github.com/charmbracelet/bubbletea"
)

// dashboardState contient les données affichées dans le dashboard.
type dashboardState struct {
	alerts   client.AlertsResponse
	products []client.Product
	loaded   bool
}

// ─── Commandes ────────────────────────────────────────────────────────────────

func (m Model) cmdLoadDashboard() tea.Cmd {
	return func() tea.Msg {
		alerts, err := m.apiClient.GetAlerts()
		if err != nil {
			return ErrMsg{Err: err}
		}
		products, err := m.apiClient.GetProducts("")
		if err != nil {
			return ErrMsg{Err: err}
		}
		return dashboardLoadedMsg{alerts: *alerts, products: products}
	}
}

type dashboardLoadedMsg struct {
	alerts   client.AlertsResponse
	products []client.Product
}

func (m Model) cmdExport(path string) tea.Cmd {
	return func() tea.Msg {
		resp, err := m.apiClient.GetExport()
		if err != nil {
			return ErrMsg{Err: fmt.Errorf("export impossible : %w", err)}
		}
		m.store.ImportFromExport(resp)
		if err := m.store.SaveToFile(path); err != nil {
			return ErrMsg{Err: fmt.Errorf("sauvegarde impossible : %w", err)}
		}
		return ExportDoneMsg{Path: path}
	}
}

// ─── Update ───────────────────────────────────────────────────────────────────

func (m Model) updateDashboard(msg tea.Msg) (Model, tea.Cmd) {
	switch msg := msg.(type) {
	case dashboardLoadedMsg:
		m.dashSt.alerts = msg.alerts
		m.dashSt.products = msg.products
		m.dashSt.loaded = true
		m.loading = false
		m.store.Products = msg.products
		return m, nil

	case ExportDoneMsg:
		m.loading = false
		m.success = fmt.Sprintf("Export sauvegardé : %s", msg.Path)
		return m, nil

	case tea.KeyMsg:
		switch msg.String() {
		case "p":
			return m, m.navigateTo(ModeProducts)
		case "s":
			return m, m.navigateTo(ModeSuppliers)
		case "m":
			return m, m.navigateTo(ModeMovements)
		case "a":
			return m, m.navigateTo(ModeAlerts)
		case "e":
			m.loading = true
			return m, m.cmdExport("stock_export.json")
		case "r":
			m.dashSt.loaded = false
			m.loading = true
			return m, m.cmdLoadDashboard()
		case "q", "ctrl+c":
			return m, tea.Quit
		}
	}
	return m, nil
}

// ─── View ─────────────────────────────────────────────────────────────────────

func (m Model) viewDashboard() string {
	var b strings.Builder
	s := m.dashSt

	// En-tête
	b.WriteString("\n")
	b.WriteString(styles.AppTitle.Render(" 📦 Gestionnaire de Stock "))
	b.WriteString("  ")
	b.WriteString(styles.MutedStyle.Render(fmt.Sprintf("connecté : %s", m.username)))
	b.WriteString("\n\n")

	if m.loading || !s.loaded {
		b.WriteString(m.spinner.View() + " Chargement du tableau de bord…\n")
		b.WriteString(viewNavDashboard())
		return b.String()
	}

	// ─── Statistiques ────
	totalProducts := len(s.products)
	lowCount := 0
	for _, p := range s.products {
		if p.IsLowStock {
			lowCount++
		}
	}

	b.WriteString(styles.Title.Render("Tableau de bord"))
	b.WriteString("\n")

	statsRow := fmt.Sprintf(
		"%s  %s  %s",
		styles.Panel.Render(fmt.Sprintf(" Produits\n %s", styles.BoldStyle.Render(fmt.Sprintf("%d", totalProducts)))),
		renderAlertStat(lowCount),
		styles.Panel.Render(fmt.Sprintf(" Seuil par défaut\n %s", styles.BoldStyle.Render(fmt.Sprintf("%d", s.alerts.DefaultThreshold)))),
	)
	b.WriteString(statsRow)
	b.WriteString("\n\n")

	// ─── Alertes de stock faible ────
	if len(s.alerts.LowStockProducts) == 0 {
		b.WriteString(styles.SuccessStyle.Render("✓ Aucun produit en alerte de stock faible"))
		b.WriteString("\n")
	} else {
		b.WriteString(styles.ErrorStyle.Render(fmt.Sprintf("⚠ %d produit(s) en alerte de stock faible :", len(s.alerts.LowStockProducts))))
		b.WriteString("\n")
		for _, p := range s.alerts.LowStockProducts {
			threshold := s.alerts.DefaultThreshold
			if p.AlertThreshold != nil {
				threshold = *p.AlertThreshold
			}
			line := fmt.Sprintf("  • %-30s  qté: %-4d  seuil: %d",
				p.Name, p.Quantity, threshold)
			b.WriteString(styles.LowStockBadge.Render(line))
			b.WriteString("\n")
		}
	}

	// ─── Messages globaux ────
	b.WriteString("\n")
	if m.err != "" {
		b.WriteString(styles.ErrorStyle.Render("✗ "+m.err) + "\n")
	}
	if m.success != "" {
		b.WriteString(styles.SuccessStyle.Render("✓ "+m.success) + "\n")
	}

	b.WriteString(viewNavDashboard())
	return b.String()
}

func renderAlertStat(lowCount int) string {
	content := fmt.Sprintf(" Alertes\n %s", styles.BoldStyle.Render(fmt.Sprintf("%d", lowCount)))
	if lowCount > 0 {
		return styles.FocusedPanel.Render(styles.ErrorStyle.Render(content))
	}
	return styles.Panel.Render(content)
}

func viewNavDashboard() string {
	return "\n" + styles.NavBar.Render(
		styles.Key("p")+" produits  "+
			styles.Key("s")+" fournisseurs  "+
			styles.Key("m")+" mouvements  "+
			styles.Key("a")+" alertes  "+
			styles.Key("e")+" exporter JSON  "+
			styles.Key("r")+" rafraîchir  "+
			styles.Key("q")+" quitter",
	)
}
