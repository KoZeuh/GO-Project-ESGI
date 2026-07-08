// === [PARTIE TUI - Frontend] ===
// Écran Fournisseurs : liste, création, édition et suppression de fournisseurs.
package model

import (
	"fmt"
	"strings"

	"github.com/KoZeuh/GO-Project-ESGI/tui/internal/client"
	"github.com/KoZeuh/GO-Project-ESGI/tui/internal/styles"
	"github.com/charmbracelet/bubbles/list"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// ─── Item de liste ────────────────────────────────────────────────────────────

type supplierItem struct{ s client.Supplier }

func (si supplierItem) FilterValue() string { return si.s.Name + " " + si.s.Email }
func (si supplierItem) Title() string       { return si.s.Name }
func (si supplierItem) Description() string {
	parts := []string{}
	if si.s.Email != "" {
		parts = append(parts, "✉ "+si.s.Email)
	}
	if si.s.Phone != "" {
		parts = append(parts, "☎ "+si.s.Phone)
	}
	if si.s.Address != "" {
		parts = append(parts, "📍 "+si.s.Address)
	}
	return strings.Join(parts, "  |  ")
}

// ─── Formulaire fournisseur ───────────────────────────────────────────────────

const supplierFieldCount = 4

type supplierForm struct {
	inputs  [supplierFieldCount]textinput.Model
	focused int
	editID  int64
	err     string
}

var supplierLabels = [supplierFieldCount]string{"Nom", "Email", "Téléphone", "Adresse"}

func newSupplierForm() supplierForm {
	placeholders := [supplierFieldCount]string{
		"Dupont SA", "contact@dupont.fr", "0102030405", "12 rue de la Paix, Paris",
	}
	var inputs [supplierFieldCount]textinput.Model
	for i := range inputs {
		t := textinput.New()
		t.Placeholder = placeholders[i]
		t.Width = 35
		inputs[i] = t
	}
	return supplierForm{inputs: inputs}
}

func (f *supplierForm) fillFromSupplier(s client.Supplier) {
	f.inputs[0].SetValue(s.Name)
	f.inputs[1].SetValue(s.Email)
	f.inputs[2].SetValue(s.Phone)
	f.inputs[3].SetValue(s.Address)
	f.editID = s.ID
}

func (f supplierForm) toRequest() (client.SupplierRequest, error) {
	name := strings.TrimSpace(f.inputs[0].Value())
	if name == "" {
		return client.SupplierRequest{}, fmt.Errorf("le nom est obligatoire")
	}
	return client.SupplierRequest{
		Name:    name,
		Email:   strings.TrimSpace(f.inputs[1].Value()),
		Phone:   strings.TrimSpace(f.inputs[2].Value()),
		Address: strings.TrimSpace(f.inputs[3].Value()),
	}, nil
}

// ─── État de l'écran ──────────────────────────────────────────────────────────

type suppliersState struct {
	list          list.Model
	form          supplierForm
	formMode      FormMode
	inForm        bool
	confirmDelete int64
}

func newSuppliersState(w, h int) suppliersState {
	delegate := list.NewDefaultDelegate()
	delegate.Styles.SelectedTitle = delegate.Styles.SelectedTitle.
		Foreground(lipgloss.Color("#FFFFFF")).
		Background(styles.ColorSelected)
	delegate.Styles.SelectedDesc = delegate.Styles.SelectedDesc.
		Foreground(styles.ColorSubtext).
		Background(styles.ColorSelected)

	l := list.New([]list.Item{}, delegate, w, h)
	l.Title = "Fournisseurs"
	l.SetShowStatusBar(true)
	l.SetFilteringEnabled(true)
	l.Styles.Title = styles.Title

	return suppliersState{list: l, form: newSupplierForm()}
}

// ─── Commandes ────────────────────────────────────────────────────────────────

func (m Model) cmdLoadSuppliers() tea.Cmd {
	return func() tea.Msg {
		suppliers, err := m.apiClient.GetSuppliers()
		if err != nil {
			return ErrMsg{Err: err}
		}
		return SuppliersLoadedMsg{Suppliers: suppliers}
	}
}

func (m Model) cmdSaveSupplier() tea.Cmd {
	form := m.suppliersSt.form
	return func() tea.Msg {
		req, err := form.toRequest()
		if err != nil {
			return ErrMsg{Err: err}
		}
		var s *client.Supplier
		if form.editID == 0 {
			s, err = m.apiClient.CreateSupplier(req)
		} else {
			s, err = m.apiClient.UpdateSupplier(form.editID, req)
		}
		if err != nil {
			return ErrMsg{Err: err}
		}
		return SupplierSavedMsg{Supplier: *s}
	}
}

func (m Model) cmdDeleteSupplier(id int64) tea.Cmd {
	return func() tea.Msg {
		if err := m.apiClient.DeleteSupplier(id); err != nil {
			return ErrMsg{Err: err}
		}
		return SupplierDeletedMsg{ID: id}
	}
}

// ─── Update ───────────────────────────────────────────────────────────────────

func (m Model) updateSuppliers(msg tea.Msg) (Model, tea.Cmd) {
	var cmds []tea.Cmd

	switch msg := msg.(type) {
	case SuppliersLoadedMsg:
		items := make([]list.Item, len(msg.Suppliers))
		for i, s := range msg.Suppliers {
			items[i] = supplierItem{s}
		}
		cmd := m.suppliersSt.list.SetItems(items)
		m.loading = false
		return m, cmd

	case SupplierSavedMsg:
		m.suppliersSt.inForm = false
		m.mode = ModeSuppliers
		m.success = fmt.Sprintf("Fournisseur « %s » sauvegardé.", msg.Supplier.Name)
		m.loading = true
		return m, m.cmdLoadSuppliers()

	case SupplierDeletedMsg:
		m.suppliersSt.confirmDelete = 0
		m.success = "Fournisseur supprimé."
		m.loading = true
		return m, m.cmdLoadSuppliers()

	case ErrMsg:
		if m.suppliersSt.inForm {
			m.suppliersSt.form.err = msg.Err.Error()
		} else {
			m.err = msg.Err.Error()
		}
		m.loading = false
		return m, nil
	}

	if m.suppliersSt.inForm {
		return m.updateSupplierForm(msg)
	}

	switch msg := msg.(type) {
	case tea.KeyMsg:
		if m.suppliersSt.confirmDelete != 0 {
			switch msg.String() {
			case "o", "y":
				id := m.suppliersSt.confirmDelete
				m.suppliersSt.confirmDelete = 0
				m.loading = true
				return m, m.cmdDeleteSupplier(id)
			default:
				m.suppliersSt.confirmDelete = 0
			}
			return m, nil
		}

		switch msg.String() {
		case "esc", "backspace":
			return m, m.navigateTo(ModeDashboard)
		case "n":
			m.suppliersSt.form = newSupplierForm()
			m.suppliersSt.form.editID = 0
			m.suppliersSt.formMode = FormCreate
			m.suppliersSt.inForm = true
			m.mode = ModeSupplierForm
			cmds = append(cmds, m.suppliersSt.form.inputs[0].Focus())
		case "e":
			if item, ok := m.suppliersSt.list.SelectedItem().(supplierItem); ok {
				f := newSupplierForm()
				f.fillFromSupplier(item.s)
				m.suppliersSt.form = f
				m.suppliersSt.formMode = FormEdit
				m.suppliersSt.inForm = true
				m.mode = ModeSupplierForm
				cmds = append(cmds, m.suppliersSt.form.inputs[0].Focus())
			}
		case "d":
			if item, ok := m.suppliersSt.list.SelectedItem().(supplierItem); ok {
				m.suppliersSt.confirmDelete = item.s.ID
			}
		case "r":
			m.loading = true
			return m, m.cmdLoadSuppliers()
		}
	}

	var cmd tea.Cmd
	m.suppliersSt.list, cmd = m.suppliersSt.list.Update(msg)
	cmds = append(cmds, cmd)
	return m, tea.Batch(cmds...)
}

func (m Model) updateSupplierForm(msg tea.Msg) (Model, tea.Cmd) {
	var cmds []tea.Cmd

	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "esc":
			m.suppliersSt.inForm = false
			m.mode = ModeSuppliers
			return m, nil
		case "tab", "down":
			m.suppliersSt.form.inputs[m.suppliersSt.form.focused].Blur()
			m.suppliersSt.form.focused = (m.suppliersSt.form.focused + 1) % supplierFieldCount
			cmds = append(cmds, m.suppliersSt.form.inputs[m.suppliersSt.form.focused].Focus())
		case "shift+tab", "up":
			m.suppliersSt.form.inputs[m.suppliersSt.form.focused].Blur()
			m.suppliersSt.form.focused = (m.suppliersSt.form.focused - 1 + supplierFieldCount) % supplierFieldCount
			cmds = append(cmds, m.suppliersSt.form.inputs[m.suppliersSt.form.focused].Focus())
		case "enter":
			if m.suppliersSt.form.focused < supplierFieldCount-1 {
				m.suppliersSt.form.inputs[m.suppliersSt.form.focused].Blur()
				m.suppliersSt.form.focused++
				cmds = append(cmds, m.suppliersSt.form.inputs[m.suppliersSt.form.focused].Focus())
			} else {
				m.loading = true
				return m, m.cmdSaveSupplier()
			}
		case "ctrl+s":
			m.loading = true
			return m, m.cmdSaveSupplier()
		}
	}

	for i := range m.suppliersSt.form.inputs {
		var cmd tea.Cmd
		m.suppliersSt.form.inputs[i], cmd = m.suppliersSt.form.inputs[i].Update(msg)
		cmds = append(cmds, cmd)
	}
	return m, tea.Batch(cmds...)
}

// ─── View ─────────────────────────────────────────────────────────────────────

func (m Model) viewSuppliers() string {
	if m.suppliersSt.inForm {
		return m.viewSupplierForm()
	}

	var b strings.Builder
	b.WriteString("\n")

	if m.loading {
		b.WriteString(m.spinner.View() + " Chargement des fournisseurs…\n")
	} else {
		b.WriteString(m.suppliersSt.list.View())
		b.WriteString("\n")
	}

	if m.err != "" {
		b.WriteString(styles.ErrorStyle.Render("✗ "+m.err) + "\n")
	}
	if m.success != "" {
		b.WriteString(styles.SuccessStyle.Render("✓ "+m.success) + "\n")
	}

	if m.suppliersSt.confirmDelete != 0 {
		b.WriteString(styles.WarningStyle.Render("⚠ Confirmer la suppression ? ") +
			styles.Key("o") + " oui  " + styles.Key("n") + " non\n")
	} else {
		b.WriteString(styles.NavBar.Render(
			styles.Key("n") + " nouveau  " +
				styles.Key("e") + " modifier  " +
				styles.Key("d") + " supprimer  " +
				styles.Key("r") + " rafraîchir  " +
				styles.Key("Esc") + " retour",
		))
	}
	return b.String()
}

func (m Model) viewSupplierForm() string {
	f := m.suppliersSt.form
	action := "Nouveau fournisseur"
	if f.editID != 0 {
		action = "Modifier le fournisseur"
	}

	var b strings.Builder
	b.WriteString("\n")
	b.WriteString(styles.Title.Render(action))
	b.WriteString("\n\n")

	for i, input := range f.inputs {
		if i == f.focused {
			b.WriteString(styles.FocusedLabel.Render(supplierLabels[i]+":") + "  ")
		} else {
			b.WriteString(styles.InputLabel.Render(supplierLabels[i]+":") + "  ")
		}
		b.WriteString(input.View())
		b.WriteString("\n")
	}

	b.WriteString("\n")
	if f.err != "" {
		b.WriteString(styles.ErrorStyle.Render("✗ "+f.err) + "\n")
	}
	if m.loading {
		b.WriteString(m.spinner.View() + " Sauvegarde…\n")
	}

	b.WriteString(styles.NavBar.Render(
		styles.Key("Tab") + " champ suivant  " +
			styles.Key("Entrée") + " suivant/valider  " +
			styles.Key("Ctrl+S") + " sauvegarder  " +
			styles.Key("Esc") + " annuler",
	))
	return b.String()
}
