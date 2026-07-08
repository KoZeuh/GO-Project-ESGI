// === [PARTIE TUI - Frontend] ===
// Écran Alertes : liste des produits en stock faible + paramètre du seuil global.
package model

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/KoZeuh/GO-Project-ESGI/tui/internal/client"
	"github.com/KoZeuh/GO-Project-ESGI/tui/internal/styles"
	"github.com/charmbracelet/bubbles/list"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// ─── Item de liste ────────────────────────────────────────────────────────────

type alertItem struct{ p client.Product }

func (ai alertItem) FilterValue() string { return ai.p.Name + " " + ai.p.Reference }
func (ai alertItem) Title() string {
	return fmt.Sprintf("%-35s  qté: %-5d", ai.p.Name, ai.p.Quantity)
}
func (ai alertItem) Description() string {
	effectiveThreshold := "seuil global"
	if ai.p.AlertThreshold != nil {
		effectiveThreshold = fmt.Sprintf("seuil produit: %d", *ai.p.AlertThreshold)
	}
	return fmt.Sprintf("Réf: %-15s | %s | Fournisseur: %s",
		ai.p.Reference, effectiveThreshold, ai.p.SupplierName)
}

// ─── État de l'écran ──────────────────────────────────────────────────────────

type alertsState struct {
	list             list.Model
	response         client.AlertsResponse
	thresholdInput   textinput.Model
	inEdit           bool
}

func newAlertsState(w, h int) alertsState {
	delegate := list.NewDefaultDelegate()
	delegate.Styles.SelectedTitle = delegate.Styles.SelectedTitle.
		Foreground(lipgloss.Color("#FFFFFF")).
		Background(styles.ColorSelected)
	delegate.Styles.SelectedDesc = delegate.Styles.SelectedDesc.
		Foreground(styles.ColorSubtext).
		Background(styles.ColorSelected)

	l := list.New([]list.Item{}, delegate, w, h)
	l.Title = "Alertes de stock faible"
	l.SetShowStatusBar(true)
	l.SetFilteringEnabled(false)
	l.Styles.Title = styles.Title

	t := textinput.New()
	t.Placeholder = "5"
	t.Width = 10
	t.CharLimit = 6

	return alertsState{list: l, thresholdInput: t}
}

// ─── Commandes ────────────────────────────────────────────────────────────────

func (m Model) cmdLoadAlerts() tea.Cmd {
	return func() tea.Msg {
		resp, err := m.apiClient.GetAlerts()
		if err != nil {
			return ErrMsg{Err: err}
		}
		return AlertsLoadedMsg{Response: *resp}
	}
}

func (m Model) cmdSaveAlertSettings() tea.Cmd {
	val := m.alertsSt.thresholdInput.Value()
	return func() tea.Msg {
		v, err := strconv.Atoi(strings.TrimSpace(val))
		if err != nil || v < 0 {
			return ErrMsg{Err: fmt.Errorf("seuil invalide (entier ≥ 0 attendu)")}
		}
		settings, err := m.apiClient.UpdateAlertSettings(client.AlertSettingsRequest{DefaultThreshold: v})
		if err != nil {
			return ErrMsg{Err: err}
		}
		return AlertSettingsSavedMsg{Settings: *settings}
	}
}

// ─── Update ───────────────────────────────────────────────────────────────────

func (m Model) updateAlerts(msg tea.Msg) (Model, tea.Cmd) {
	var cmds []tea.Cmd

	switch msg := msg.(type) {
	case AlertsLoadedMsg:
		m.alertsSt.response = msg.Response
		items := make([]list.Item, len(msg.Response.LowStockProducts))
		for i, p := range msg.Response.LowStockProducts {
			items[i] = alertItem{p}
		}
		cmd := m.alertsSt.list.SetItems(items)
		m.loading = false
		return m, cmd

	case AlertSettingsSavedMsg:
		m.alertsSt.inEdit = false
		m.alertsSt.response.DefaultThreshold = msg.Settings.DefaultThreshold
		m.success = fmt.Sprintf("Seuil global mis à jour : %d", msg.Settings.DefaultThreshold)
		m.loading = true
		return m, m.cmdLoadAlerts()

	case ErrMsg:
		m.err = msg.Err.Error()
		m.loading = false
		return m, nil
	}

	switch msg := msg.(type) {
	case tea.KeyMsg:
		if m.alertsSt.inEdit {
			switch msg.String() {
			case "esc":
				m.alertsSt.inEdit = false
				return m, nil
			case "enter", "ctrl+s":
				m.loading = true
				return m, m.cmdSaveAlertSettings()
			}
			var cmd tea.Cmd
			m.alertsSt.thresholdInput, cmd = m.alertsSt.thresholdInput.Update(msg)
			return m, cmd
		}

		switch msg.String() {
		case "esc", "backspace":
			return m, m.navigateTo(ModeDashboard)
		case "e":
			m.alertsSt.thresholdInput.SetValue(
				strconv.Itoa(m.alertsSt.response.DefaultThreshold),
			)
			m.alertsSt.inEdit = true
			cmds = append(cmds, m.alertsSt.thresholdInput.Focus())
		case "r":
			m.loading = true
			return m, m.cmdLoadAlerts()
		}
	}

	if !m.alertsSt.inEdit {
		var cmd tea.Cmd
		m.alertsSt.list, cmd = m.alertsSt.list.Update(msg)
		cmds = append(cmds, cmd)
	}
	return m, tea.Batch(cmds...)
}

// ─── View ─────────────────────────────────────────────────────────────────────

func (m Model) viewAlerts() string {
	var b strings.Builder
	s := m.alertsSt

	b.WriteString("\n")
	b.WriteString(styles.Title.Render("Alertes de stock faible"))
	b.WriteString("\n")

	// Seuil global
	thresholdLine := fmt.Sprintf("Seuil par défaut global : %s",
		styles.BoldStyle.Render(strconv.Itoa(s.response.DefaultThreshold)))
	b.WriteString(styles.Panel.Render(thresholdLine))
	b.WriteString("\n\n")

	// Édition du seuil
	if s.inEdit {
		b.WriteString(styles.FocusedLabel.Render("Nouveau seuil:") + "  ")
		b.WriteString(s.thresholdInput.View())
		b.WriteString("\n")
		b.WriteString(styles.MutedStyle.Render("Appuyez sur Entrée pour valider, Esc pour annuler"))
		b.WriteString("\n\n")
	}

	if m.loading {
		b.WriteString(m.spinner.View() + " Chargement…\n")
	} else if len(s.response.LowStockProducts) == 0 {
		b.WriteString(styles.SuccessStyle.Render("✓ Aucun produit en alerte de stock faible"))
		b.WriteString("\n")
	} else {
		b.WriteString(styles.ErrorStyle.Render(
			fmt.Sprintf("⚠ %d produit(s) sous le seuil :", len(s.response.LowStockProducts)),
		))
		b.WriteString("\n")
		b.WriteString(s.list.View())
		b.WriteString("\n")
	}

	if m.err != "" {
		b.WriteString(styles.ErrorStyle.Render("✗ "+m.err) + "\n")
	}
	if m.success != "" {
		b.WriteString(styles.SuccessStyle.Render("✓ "+m.success) + "\n")
	}

	if !s.inEdit {
		b.WriteString(styles.NavBar.Render(
			styles.Key("e") + " modifier seuil global  " +
				styles.Key("r") + " rafraîchir  " +
				styles.Key("Esc") + " retour",
		))
	}
	return b.String()
}
