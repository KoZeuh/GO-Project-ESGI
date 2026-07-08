// === [PARTIE TUI - Frontend] ===
// Écran Mouvements : historique des entrées/sorties + formulaire de nouveau mouvement.
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

type movementItem struct{ mv client.Movement }

func (mi movementItem) FilterValue() string {
	return mi.mv.ProductName + " " + mi.mv.Type + " " + mi.mv.Username
}
func (mi movementItem) Title() string {
	badge := styles.MovTypeBadge(mi.mv.Type)
	return fmt.Sprintf("%s  %-30s  qté: %d", badge, mi.mv.ProductName, mi.mv.Quantity)
}
func (mi movementItem) Description() string {
	note := mi.mv.Note
	if note == "" {
		note = "—"
	}
	return fmt.Sprintf("Par: %-12s | Note: %-25s | %s",
		mi.mv.Username, note, mi.mv.CreatedAt.Format("02/01/2006 15:04"))
}

// ─── Formulaire mouvement (réapprovisionnement / sortie) ─────────────────────

const movFieldCount = 4

type movementForm struct {
	inputs  [movFieldCount]textinput.Model
	focused int
	err     string
}

var movLabels = [movFieldCount]string{"ID Produit", "Type (IN/OUT)", "Quantité", "Note"}

func newMovementForm() movementForm {
	placeholders := [movFieldCount]string{"5", "IN", "10", "Réassort hebdomadaire"}
	var inputs [movFieldCount]textinput.Model
	for i := range inputs {
		t := textinput.New()
		t.Placeholder = placeholders[i]
		t.Width = 30
		inputs[i] = t
	}
	// Valeur par défaut pour le type
	inputs[1].SetValue("IN")
	return movementForm{inputs: inputs}
}

func (f movementForm) toRequest() (client.MovementRequest, error) {
	pidStr := strings.TrimSpace(f.inputs[0].Value())
	typ := strings.ToUpper(strings.TrimSpace(f.inputs[1].Value()))
	qtyStr := strings.TrimSpace(f.inputs[2].Value())
	note := strings.TrimSpace(f.inputs[3].Value())

	pid, err := strconv.ParseInt(pidStr, 10, 64)
	if err != nil || pid <= 0 {
		return client.MovementRequest{}, fmt.Errorf("ID produit invalide")
	}
	if typ != "IN" && typ != "OUT" {
		return client.MovementRequest{}, fmt.Errorf("type doit être IN ou OUT")
	}
	qty, err := strconv.Atoi(qtyStr)
	if err != nil || qty <= 0 {
		return client.MovementRequest{}, fmt.Errorf("quantité invalide (doit être > 0)")
	}
	return client.MovementRequest{ProductID: pid, Type: typ, Quantity: qty, Note: note}, nil
}

// ─── Filtre ───────────────────────────────────────────────────────────────────

type movFilter struct {
	typeFilter string // "", "IN", "OUT"
}

func (f movFilter) label() string {
	switch f.typeFilter {
	case "IN":
		return styles.INBadge.Render("▲ IN seulement")
	case "OUT":
		return styles.OUTBadge.Render("▼ OUT seulement")
	default:
		return styles.MutedStyle.Render("tous les types")
	}
}

// ─── État de l'écran ──────────────────────────────────────────────────────────

type movementsState struct {
	list   list.Model
	form   movementForm
	inForm bool
	filter movFilter
}

func newMovementsState(w, h int) movementsState {
	delegate := list.NewDefaultDelegate()
	delegate.Styles.SelectedTitle = delegate.Styles.SelectedTitle.
		Foreground(lipgloss.Color("#FFFFFF")).
		Background(styles.ColorSelected)
	delegate.Styles.SelectedDesc = delegate.Styles.SelectedDesc.
		Foreground(styles.ColorSubtext).
		Background(styles.ColorSelected)

	l := list.New([]list.Item{}, delegate, w, h)
	l.Title = "Historique des mouvements"
	l.SetShowStatusBar(true)
	l.SetFilteringEnabled(true)
	l.Styles.Title = styles.Title

	return movementsState{list: l, form: newMovementForm()}
}

// ─── Commandes ────────────────────────────────────────────────────────────────

func (m Model) cmdLoadMovements() tea.Cmd {
	f := m.movementsSt.filter
	return func() tea.Msg {
		movs, err := m.apiClient.GetMovements(0, f.typeFilter, "", "")
		if err != nil {
			return ErrMsg{Err: err}
		}
		return MovementsLoadedMsg{Movements: movs}
	}
}

func (m Model) cmdCreateMovement() tea.Cmd {
	form := m.movementsSt.form
	return func() tea.Msg {
		req, err := form.toRequest()
		if err != nil {
			return ErrMsg{Err: err}
		}
		mv, err := m.apiClient.CreateMovement(req)
		if err != nil {
			return ErrMsg{Err: err}
		}
		return MovementCreatedMsg{Movement: *mv}
	}
}

// ─── Update ───────────────────────────────────────────────────────────────────

func (m Model) updateMovements(msg tea.Msg) (Model, tea.Cmd) {
	var cmds []tea.Cmd

	switch msg := msg.(type) {
	case MovementsLoadedMsg:
		items := make([]list.Item, len(msg.Movements))
		for i, mv := range msg.Movements {
			items[i] = movementItem{mv}
		}
		cmd := m.movementsSt.list.SetItems(items)
		m.loading = false
		return m, cmd

	case MovementCreatedMsg:
		m.movementsSt.inForm = false
		m.mode = ModeMovements
		badge := styles.MovTypeBadge(msg.Movement.Type)
		m.success = fmt.Sprintf("Mouvement %s créé (qté: %d).", badge, msg.Movement.Quantity)
		m.loading = true
		return m, m.cmdLoadMovements()

	case ErrMsg:
		if m.movementsSt.inForm {
			m.movementsSt.form.err = msg.Err.Error()
		} else {
			m.err = msg.Err.Error()
		}
		m.loading = false
		return m, nil
	}

	if m.movementsSt.inForm {
		return m.updateMovementForm(msg)
	}

	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "esc", "backspace":
			return m, m.navigateTo(ModeDashboard)
		case "n":
			m.movementsSt.form = newMovementForm()
			m.movementsSt.inForm = true
			m.mode = ModeRestock
			cmds = append(cmds, m.movementsSt.form.inputs[0].Focus())
		case "f":
			// Cycle sur le filtre de type : "" -> "IN" -> "OUT" -> ""
			switch m.movementsSt.filter.typeFilter {
			case "":
				m.movementsSt.filter.typeFilter = "IN"
			case "IN":
				m.movementsSt.filter.typeFilter = "OUT"
			default:
				m.movementsSt.filter.typeFilter = ""
			}
			m.loading = true
			return m, m.cmdLoadMovements()
		case "r":
			m.loading = true
			return m, m.cmdLoadMovements()
		}
	}

	var cmd tea.Cmd
	m.movementsSt.list, cmd = m.movementsSt.list.Update(msg)
	cmds = append(cmds, cmd)
	return m, tea.Batch(cmds...)
}

func (m Model) updateMovementForm(msg tea.Msg) (Model, tea.Cmd) {
	var cmds []tea.Cmd

	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "esc":
			m.movementsSt.inForm = false
			m.mode = ModeMovements
			return m, nil
		case "tab", "down":
			m.movementsSt.form.inputs[m.movementsSt.form.focused].Blur()
			m.movementsSt.form.focused = (m.movementsSt.form.focused + 1) % movFieldCount
			cmds = append(cmds, m.movementsSt.form.inputs[m.movementsSt.form.focused].Focus())
		case "shift+tab", "up":
			m.movementsSt.form.inputs[m.movementsSt.form.focused].Blur()
			m.movementsSt.form.focused = (m.movementsSt.form.focused - 1 + movFieldCount) % movFieldCount
			cmds = append(cmds, m.movementsSt.form.inputs[m.movementsSt.form.focused].Focus())
		case "enter":
			if m.movementsSt.form.focused < movFieldCount-1 {
				m.movementsSt.form.inputs[m.movementsSt.form.focused].Blur()
				m.movementsSt.form.focused++
				cmds = append(cmds, m.movementsSt.form.inputs[m.movementsSt.form.focused].Focus())
			} else {
				m.loading = true
				return m, m.cmdCreateMovement()
			}
		case "ctrl+s":
			m.loading = true
			return m, m.cmdCreateMovement()
		}
	}

	for i := range m.movementsSt.form.inputs {
		var cmd tea.Cmd
		m.movementsSt.form.inputs[i], cmd = m.movementsSt.form.inputs[i].Update(msg)
		cmds = append(cmds, cmd)
	}
	return m, tea.Batch(cmds...)
}

// ─── View ─────────────────────────────────────────────────────────────────────

func (m Model) viewMovements() string {
	if m.movementsSt.inForm {
		return m.viewMovementForm()
	}

	var b strings.Builder
	b.WriteString("\n")
	b.WriteString(styles.Subtitle.Render("Filtre : "+m.movementsSt.filter.label()) + "\n\n")

	if m.loading {
		b.WriteString(m.spinner.View() + " Chargement des mouvements…\n")
	} else {
		b.WriteString(m.movementsSt.list.View())
		b.WriteString("\n")
	}

	if m.err != "" {
		b.WriteString(styles.ErrorStyle.Render("✗ "+m.err) + "\n")
	}
	if m.success != "" {
		b.WriteString(styles.SuccessStyle.Render("✓ "+m.success) + "\n")
	}

	b.WriteString(styles.NavBar.Render(
		styles.Key("n") + " nouveau mvt  " +
			styles.Key("f") + " filtrer type  " +
			styles.Key("r") + " rafraîchir  " +
			styles.Key("Esc") + " retour",
	))
	return b.String()
}

func (m Model) viewMovementForm() string {
	f := m.movementsSt.form

	var b strings.Builder
	b.WriteString("\n")
	b.WriteString(styles.Title.Render("Nouveau mouvement de stock"))
	b.WriteString("\n\n")

	for i, input := range f.inputs {
		if i == f.focused {
			b.WriteString(styles.FocusedLabel.Render(movLabels[i]+":") + "  ")
		} else {
			b.WriteString(styles.InputLabel.Render(movLabels[i]+":") + "  ")
		}
		b.WriteString(input.View())
		b.WriteString("\n")
	}

	b.WriteString("\n")
	b.WriteString(styles.MutedStyle.Render(
		"Type IN = entrée de stock (réapprovisionnement)\n" +
			"Type OUT = sortie de stock (vente, perte…)",
	))
	b.WriteString("\n\n")

	if f.err != "" {
		b.WriteString(styles.ErrorStyle.Render("✗ "+f.err) + "\n")
	}
	if m.loading {
		b.WriteString(m.spinner.View() + " Enregistrement…\n")
	}

	b.WriteString(styles.NavBar.Render(
		styles.Key("Tab") + " champ suivant  " +
			styles.Key("Entrée") + " suivant/valider  " +
			styles.Key("Ctrl+S") + " sauvegarder  " +
			styles.Key("Esc") + " annuler",
	))
	return b.String()
}
