// === [PARTIE TUI - Frontend] ===
// Écran Produits : liste filtrée, création, édition et suppression de produits.
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

type productItem struct{ p client.Product }

func (pi productItem) FilterValue() string { return pi.p.Name + " " + pi.p.Reference }
func (pi productItem) Title() string {
	badge := styles.StockBadge(pi.p.IsLowStock)
	return fmt.Sprintf("%-35s  %s", pi.p.Name, badge)
}
func (pi productItem) Description() string {
	return fmt.Sprintf("Réf: %-15s | Qté: %-5d | Prix: %6.2f € | %s",
		pi.p.Reference, pi.p.Quantity, pi.p.Price, pi.p.SupplierName)
}

// ─── Formulaire produit ───────────────────────────────────────────────────────

const productFieldCount = 6

type productForm struct {
	inputs  [productFieldCount]textinput.Model
	focused int
	editID  int64 // 0 = création
	err     string
}

var productLabels = [productFieldCount]string{
	"Nom", "Référence", "Quantité", "Prix (€)", "Seuil d'alerte", "ID Fournisseur",
}

func newProductForm() productForm {
	placeholders := [productFieldCount]string{
		"Café en grains 1 kg", "CAF-1KG-001", "42", "12.50", "laisser vide = seuil global", "1",
	}
	var inputs [productFieldCount]textinput.Model
	for i := range inputs {
		t := textinput.New()
		t.Placeholder = placeholders[i]
		t.Width = 30
		inputs[i] = t
	}
	return productForm{inputs: inputs}
}

func (f *productForm) fillFromProduct(p client.Product) {
	f.inputs[0].SetValue(p.Name)
	f.inputs[1].SetValue(p.Reference)
	f.inputs[2].SetValue(strconv.Itoa(p.Quantity))
	f.inputs[3].SetValue(fmt.Sprintf("%.2f", p.Price))
	if p.AlertThreshold != nil {
		f.inputs[4].SetValue(strconv.Itoa(*p.AlertThreshold))
	} else {
		f.inputs[4].SetValue("")
	}
	f.inputs[5].SetValue(strconv.FormatInt(p.SupplierID, 10))
	f.editID = p.ID
}

func (f productForm) toRequest() (client.ProductRequest, error) {
	name := strings.TrimSpace(f.inputs[0].Value())
	ref := strings.TrimSpace(f.inputs[1].Value())
	qtyStr := strings.TrimSpace(f.inputs[2].Value())
	priceStr := strings.TrimSpace(f.inputs[3].Value())
	threshStr := strings.TrimSpace(f.inputs[4].Value())
	supplierStr := strings.TrimSpace(f.inputs[5].Value())

	if name == "" || ref == "" {
		return client.ProductRequest{}, fmt.Errorf("le nom et la référence sont obligatoires")
	}
	qty, err := strconv.Atoi(qtyStr)
	if err != nil || qty < 0 {
		return client.ProductRequest{}, fmt.Errorf("quantité invalide")
	}
	price, err := strconv.ParseFloat(priceStr, 64)
	if err != nil || price < 0 {
		return client.ProductRequest{}, fmt.Errorf("prix invalide")
	}
	supplierID, err := strconv.ParseInt(supplierStr, 10, 64)
	if err != nil || supplierID <= 0 {
		return client.ProductRequest{}, fmt.Errorf("ID fournisseur invalide")
	}

	req := client.ProductRequest{
		Name: name, Reference: ref, Quantity: qty, Price: price, SupplierID: supplierID,
	}
	if threshStr != "" {
		v, err := strconv.Atoi(threshStr)
		if err != nil || v < 0 {
			return client.ProductRequest{}, fmt.Errorf("seuil invalide")
		}
		req.AlertThreshold = &v
	}
	return req, nil
}

// ─── État de l'écran ──────────────────────────────────────────────────────────

type productsState struct {
	list          list.Model
	form          productForm
	formMode      FormMode
	inForm        bool
	confirmDelete int64 // ID du produit à supprimer (0 = pas de confirmation)
}

func newProductsState(w, h int) productsState {
	delegate := list.NewDefaultDelegate()
	delegate.Styles.SelectedTitle = delegate.Styles.SelectedTitle.
		Foreground(lipgloss.Color("#FFFFFF")).
		Background(styles.ColorSelected)
	delegate.Styles.SelectedDesc = delegate.Styles.SelectedDesc.
		Foreground(styles.ColorSubtext).
		Background(styles.ColorSelected)

	l := list.New([]list.Item{}, delegate, w, h)
	l.Title = "Produits"
	l.SetShowStatusBar(true)
	l.SetFilteringEnabled(true)
	l.Styles.Title = styles.Title

	return productsState{list: l, form: newProductForm()}
}

// ─── Commandes ────────────────────────────────────────────────────────────────

func (m Model) cmdLoadProducts() tea.Cmd {
	return func() tea.Msg {
		products, err := m.apiClient.GetProducts("")
		if err != nil {
			return ErrMsg{Err: err}
		}
		return ProductsLoadedMsg{Products: products}
	}
}

func (m Model) cmdSaveProduct() tea.Cmd {
	form := m.productsSt.form
	return func() tea.Msg {
		req, err := form.toRequest()
		if err != nil {
			return ErrMsg{Err: err}
		}
		var p *client.Product
		if form.editID == 0 {
			p, err = m.apiClient.CreateProduct(req)
		} else {
			p, err = m.apiClient.UpdateProduct(form.editID, req)
		}
		if err != nil {
			return ErrMsg{Err: err}
		}
		return ProductSavedMsg{Product: *p}
	}
}

func (m Model) cmdDeleteProduct(id int64) tea.Cmd {
	return func() tea.Msg {
		if err := m.apiClient.DeleteProduct(id); err != nil {
			return ErrMsg{Err: err}
		}
		return ProductDeletedMsg{ID: id}
	}
}

// ─── Update ───────────────────────────────────────────────────────────────────

func (m Model) updateProducts(msg tea.Msg) (Model, tea.Cmd) {
	var cmds []tea.Cmd

	switch msg := msg.(type) {
	case ProductsLoadedMsg:
		items := make([]list.Item, len(msg.Products))
		for i, p := range msg.Products {
			items[i] = productItem{p}
		}
		cmd := m.productsSt.list.SetItems(items)
		m.loading = false
		return m, cmd

	case ProductSavedMsg:
		m.productsSt.inForm = false
		m.mode = ModeProducts
		m.success = fmt.Sprintf("Produit « %s » sauvegardé.", msg.Product.Name)
		m.loading = true
		return m, m.cmdLoadProducts()

	case ProductDeletedMsg:
		m.productsSt.confirmDelete = 0
		m.success = "Produit supprimé."
		m.loading = true
		return m, m.cmdLoadProducts()

	case ErrMsg:
		if m.productsSt.inForm {
			m.productsSt.form.err = msg.Err.Error()
		} else {
			m.err = msg.Err.Error()
		}
		m.loading = false
		return m, nil
	}

	// ── Mode formulaire ──
	if m.productsSt.inForm {
		return m.updateProductForm(msg)
	}

	// ── Mode liste ──
	switch msg := msg.(type) {
	case tea.KeyMsg:
		// Confirmation de suppression active
		if m.productsSt.confirmDelete != 0 {
			switch msg.String() {
			case "o", "y":
				id := m.productsSt.confirmDelete
				m.productsSt.confirmDelete = 0
				m.loading = true
				return m, m.cmdDeleteProduct(id)
			default:
				m.productsSt.confirmDelete = 0
			}
			return m, nil
		}

		switch msg.String() {
		case "esc", "backspace":
			return m, m.navigateTo(ModeDashboard)
		case "n":
			m.productsSt.form = newProductForm()
			m.productsSt.form.editID = 0
			m.productsSt.formMode = FormCreate
			m.productsSt.inForm = true
			m.mode = ModeProductForm
			cmds = append(cmds, m.productsSt.form.inputs[0].Focus())
		case "e":
			if item, ok := m.productsSt.list.SelectedItem().(productItem); ok {
				f := newProductForm()
				f.fillFromProduct(item.p)
				m.productsSt.form = f
				m.productsSt.formMode = FormEdit
				m.productsSt.inForm = true
				m.mode = ModeProductForm
				cmds = append(cmds, m.productsSt.form.inputs[0].Focus())
			}
		case "d":
			if item, ok := m.productsSt.list.SelectedItem().(productItem); ok {
				m.productsSt.confirmDelete = item.p.ID
			}
		case "r":
			m.loading = true
			return m, m.cmdLoadProducts()
		}
	}

	var cmd tea.Cmd
	m.productsSt.list, cmd = m.productsSt.list.Update(msg)
	cmds = append(cmds, cmd)
	return m, tea.Batch(cmds...)
}

func (m Model) updateProductForm(msg tea.Msg) (Model, tea.Cmd) {
	var cmds []tea.Cmd

	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "esc":
			m.productsSt.inForm = false
			m.mode = ModeProducts
			return m, nil
		case "tab", "down":
			m.productsSt.form.inputs[m.productsSt.form.focused].Blur()
			m.productsSt.form.focused = (m.productsSt.form.focused + 1) % productFieldCount
			cmds = append(cmds, m.productsSt.form.inputs[m.productsSt.form.focused].Focus())
		case "shift+tab", "up":
			m.productsSt.form.inputs[m.productsSt.form.focused].Blur()
			m.productsSt.form.focused = (m.productsSt.form.focused - 1 + productFieldCount) % productFieldCount
			cmds = append(cmds, m.productsSt.form.inputs[m.productsSt.form.focused].Focus())
		case "enter":
			if m.productsSt.form.focused < productFieldCount-1 {
				m.productsSt.form.inputs[m.productsSt.form.focused].Blur()
				m.productsSt.form.focused++
				cmds = append(cmds, m.productsSt.form.inputs[m.productsSt.form.focused].Focus())
			} else {
				m.loading = true
				return m, m.cmdSaveProduct()
			}
		case "ctrl+s":
			m.loading = true
			return m, m.cmdSaveProduct()
		}
	}

	// Propage les événements à tous les inputs
	for i := range m.productsSt.form.inputs {
		var cmd tea.Cmd
		m.productsSt.form.inputs[i], cmd = m.productsSt.form.inputs[i].Update(msg)
		cmds = append(cmds, cmd)
	}
	return m, tea.Batch(cmds...)
}

// ─── View ─────────────────────────────────────────────────────────────────────

func (m Model) viewProducts() string {
	if m.productsSt.inForm {
		return m.viewProductForm()
	}

	var b strings.Builder
	b.WriteString("\n")

	if m.loading {
		b.WriteString(m.spinner.View() + " Chargement des produits…\n")
	} else {
		b.WriteString(m.productsSt.list.View())
		b.WriteString("\n")
	}

	if m.err != "" {
		b.WriteString(styles.ErrorStyle.Render("✗ "+m.err) + "\n")
	}
	if m.success != "" {
		b.WriteString(styles.SuccessStyle.Render("✓ "+m.success) + "\n")
	}

	if m.productsSt.confirmDelete != 0 {
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

func (m Model) viewProductForm() string {
	f := m.productsSt.form
	action := "Nouveau produit"
	if f.editID != 0 {
		action = "Modifier le produit"
	}

	var b strings.Builder
	b.WriteString("\n")
	b.WriteString(styles.Title.Render(action))
	b.WriteString("\n\n")

	for i, input := range f.inputs {
		if i == f.focused {
			b.WriteString(styles.FocusedLabel.Render(productLabels[i]+":") + "  ")
		} else {
			b.WriteString(styles.InputLabel.Render(productLabels[i]+":") + "  ")
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
