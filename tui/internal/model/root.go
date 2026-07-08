// === [PARTIE TUI - Frontend] ===
// Modèle racine Bubble Tea : agrège tous les états d'écran et gère la navigation.
package model

import (
	"github.com/KoZeuh/GO-Project-ESGI/tui/internal/client"
	"github.com/KoZeuh/GO-Project-ESGI/tui/internal/store"
	"github.com/charmbracelet/bubbles/spinner"
	tea "github.com/charmbracelet/bubbletea"
)

// Model est le modèle principal de l'application TUI.
// Chaque champ xxxSt contient l'état d'un écran spécifique,
// défini dans le fichier correspondant (login.go, products.go, etc.).
type Model struct {
	mode Mode

	apiClient client.Client
	store     *store.Store

	// Authentification
	token    string
	username string

	// Dimensions du terminal
	width  int
	height int

	// Spinner partagé pour les états de chargement
	spinner spinner.Model
	loading bool

	// Messages globaux
	err     string
	success string

	// États des écrans (définis dans les fichiers respectifs)
	loginSt     loginState
	dashSt      dashboardState
	productsSt  productsState
	suppliersSt suppliersState
	movementsSt movementsState
	alertsSt    alertsState
}

// New construit le modèle initial avec le client et le store donnés.
func New(c client.Client, s *store.Store) Model {
	sp := spinner.New()
	sp.Spinner = spinner.Dot

	m := Model{
		mode:      ModeLogin,
		apiClient: c,
		store:     s,
		spinner:   sp,
		width:     100,
		height:    30,
	}
	m.loginSt = newLoginState()
	m.productsSt = newProductsState(m.listWidth(), m.listHeight())
	m.suppliersSt = newSuppliersState(m.listWidth(), m.listHeight())
	m.movementsSt = newMovementsState(m.listWidth(), m.listHeight())
	m.alertsSt = newAlertsState(m.listWidth(), m.listHeight())
	return m
}

// Init implémente tea.Model.
func (m Model) Init() tea.Cmd {
	return tea.Batch(m.spinner.Tick, m.loginSt.inputs[0].Focus())
}

// Update implémente tea.Model.
func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmds []tea.Cmd

	switch msg := msg.(type) {
	case tea.KeyMsg:
		if msg.String() == "ctrl+c" {
			return m, tea.Quit
		}
		// Efface les messages globaux à la prochaine frappe
		m.err = ""
		m.success = ""

	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		m.resizeLists()
		return m, nil

	case spinner.TickMsg:
		var cmd tea.Cmd
		m.spinner, cmd = m.spinner.Update(msg)
		cmds = append(cmds, cmd)
	}

	// Délègue au gestionnaire d'écran courant
	var cmd tea.Cmd
	switch m.mode {
	case ModeLogin:
		m, cmd = m.updateLogin(msg)
	case ModeDashboard:
		m, cmd = m.updateDashboard(msg)
	case ModeProducts, ModeProductForm:
		m, cmd = m.updateProducts(msg)
	case ModeSuppliers, ModeSupplierForm:
		m, cmd = m.updateSuppliers(msg)
	case ModeMovements, ModeRestock:
		m, cmd = m.updateMovements(msg)
	case ModeAlerts:
		m, cmd = m.updateAlerts(msg)
	}
	cmds = append(cmds, cmd)

	return m, tea.Batch(cmds...)
}

// View implémente tea.Model.
func (m Model) View() string {
	switch m.mode {
	case ModeLogin:
		return m.viewLogin()
	case ModeDashboard:
		return m.viewDashboard()
	case ModeProducts, ModeProductForm:
		return m.viewProducts()
	case ModeSuppliers, ModeSupplierForm:
		return m.viewSuppliers()
	case ModeMovements, ModeRestock:
		return m.viewMovements()
	case ModeAlerts:
		return m.viewAlerts()
	}
	return m.spinner.View() + " Chargement…"
}

// ─── Helpers internes ─────────────────────────────────────────────────────────

func (m Model) listWidth() int {
	w := m.width - 4
	if w < 40 {
		return 80
	}
	return w
}

func (m Model) listHeight() int {
	h := m.height - 10
	if h < 5 {
		return 20
	}
	return h
}

func (m *Model) resizeLists() {
	w := m.listWidth()
	h := m.listHeight()
	m.productsSt.list.SetWidth(w)
	m.productsSt.list.SetHeight(h)
	m.suppliersSt.list.SetWidth(w)
	m.suppliersSt.list.SetHeight(h)
	m.movementsSt.list.SetWidth(w)
	m.movementsSt.list.SetHeight(h)
	m.alertsSt.list.SetWidth(w)
	m.alertsSt.list.SetHeight(h)
}

// navigateTo change d'écran et charge les données si nécessaire.
func (m *Model) navigateTo(mode Mode) tea.Cmd {
	m.mode = mode
	m.err = ""
	m.success = ""
	switch mode {
	case ModeDashboard:
		m.loading = true
		return m.cmdLoadDashboard()
	case ModeProducts:
		m.loading = true
		return m.cmdLoadProducts()
	case ModeSuppliers:
		m.loading = true
		return m.cmdLoadSuppliers()
	case ModeMovements:
		m.loading = true
		return m.cmdLoadMovements()
	case ModeAlerts:
		m.loading = true
		return m.cmdLoadAlerts()
	}
	return nil
}
