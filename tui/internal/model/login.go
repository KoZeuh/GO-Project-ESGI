// === [PARTIE TUI - Frontend] ===
// Écran de connexion : formulaire username / password, soumission via JWT.
package model

import (
	"fmt"
	"strings"

	"github.com/KoZeuh/GO-Project-ESGI/tui/internal/styles"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
)

// loginState contient tout l'état de l'écran de connexion.
type loginState struct {
	inputs    [2]textinput.Model // [0]=username, [1]=password
	focused   int
	err       string
	loading   bool
	isRegister bool // true = mode inscription
}

func newLoginState() loginState {
	username := textinput.New()
	username.Placeholder = "Nom d'utilisateur"
	username.CharLimit = 50
	username.Width = 30

	password := textinput.New()
	password.Placeholder = "Mot de passe"
	password.EchoMode = textinput.EchoPassword
	password.EchoCharacter = '•'
	password.CharLimit = 72
	password.Width = 30

	return loginState{inputs: [2]textinput.Model{username, password}}
}

// ─── Commandes ────────────────────────────────────────────────────────────────

func (m Model) cmdLogin() tea.Cmd {
	username := m.loginSt.inputs[0].Value()
	password := m.loginSt.inputs[1].Value()
	return func() tea.Msg {
		resp, err := m.apiClient.Login(username, password)
		if err != nil {
			return ErrMsg{Err: fmt.Errorf("connexion impossible : %w", err)}
		}
		return LoginSuccessMsg{Token: resp.Token, Username: resp.User.Username}
	}
}

func (m Model) cmdRegister() tea.Cmd {
	username := m.loginSt.inputs[0].Value()
	password := m.loginSt.inputs[1].Value()
	return func() tea.Msg {
		_, err := m.apiClient.Register(username, password)
		if err != nil {
			return ErrMsg{Err: fmt.Errorf("inscription impossible : %w", err)}
		}
		// Après inscription, on tente la connexion automatiquement
		resp, err := m.apiClient.Login(username, password)
		if err != nil {
			return ErrMsg{Err: err}
		}
		return LoginSuccessMsg{Token: resp.Token, Username: resp.User.Username}
	}
}

// ─── Update ───────────────────────────────────────────────────────────────────

func (m Model) updateLogin(msg tea.Msg) (Model, tea.Cmd) {
	var cmds []tea.Cmd

	switch msg := msg.(type) {
	case LoginSuccessMsg:
		m.token = msg.Token
		m.username = msg.Username
		m.apiClient.SetToken(msg.Token)
		m.loginSt.loading = false
		m.loginSt.err = ""
		// Réinitialise les champs du formulaire
		m.loginSt = newLoginState()
		cmd := m.navigateTo(ModeDashboard)
		return m, cmd

	case ErrMsg:
		m.loginSt.err = msg.Err.Error()
		m.loginSt.loading = false
		return m, nil

	case tea.KeyMsg:
		if m.loginSt.loading {
			return m, nil
		}
		switch msg.String() {
		case "tab", "down":
			m.loginSt.focused = (m.loginSt.focused + 1) % 2
			for i := range m.loginSt.inputs {
				if i == m.loginSt.focused {
					cmds = append(cmds, m.loginSt.inputs[i].Focus())
				} else {
					m.loginSt.inputs[i].Blur()
				}
			}
		case "shift+tab", "up":
			m.loginSt.focused = (m.loginSt.focused + 1) % 2
			for i := range m.loginSt.inputs {
				if i == m.loginSt.focused {
					cmds = append(cmds, m.loginSt.inputs[i].Focus())
				} else {
					m.loginSt.inputs[i].Blur()
				}
			}
		case "enter":
			if m.loginSt.focused < 1 {
				// Passe au champ suivant si on n'est pas au dernier
				m.loginSt.focused++
				for i := range m.loginSt.inputs {
					if i == m.loginSt.focused {
						cmds = append(cmds, m.loginSt.inputs[i].Focus())
					} else {
						m.loginSt.inputs[i].Blur()
					}
				}
			} else {
				// Soumet le formulaire
				if m.loginSt.inputs[0].Value() == "" || m.loginSt.inputs[1].Value() == "" {
					m.loginSt.err = "Veuillez remplir tous les champs"
					return m, nil
				}
				m.loginSt.loading = true
				m.loginSt.err = ""
				if m.loginSt.isRegister {
					return m, m.cmdRegister()
				}
				return m, m.cmdLogin()
			}
		case "ctrl+r":
			m.loginSt.isRegister = !m.loginSt.isRegister
			m.loginSt.err = ""
		}
	}

	// Propage les événements aux inputs
	for i := range m.loginSt.inputs {
		var cmd tea.Cmd
		m.loginSt.inputs[i], cmd = m.loginSt.inputs[i].Update(msg)
		cmds = append(cmds, cmd)
	}

	return m, tea.Batch(cmds...)
}

// ─── View ─────────────────────────────────────────────────────────────────────

func (m Model) viewLogin() string {
	s := m.loginSt

	var b strings.Builder

	// En-tête
	b.WriteString("\n")
	b.WriteString(styles.AppTitle.Render(" 📦 Gestionnaire de Stock "))
	b.WriteString("\n\n")

	action := "Connexion"
	if s.isRegister {
		action = "Inscription"
	}
	b.WriteString(styles.Title.Render(action))
	b.WriteString("\n")

	// Champs
	for i, input := range s.inputs {
		label := ""
		switch i {
		case 0:
			label = "Utilisateur"
		case 1:
			label = "Mot de passe"
		}
		if i == s.focused {
			b.WriteString(styles.FocusedLabel.Render(label+":") + "  ")
		} else {
			b.WriteString(styles.InputLabel.Render(label+":") + "  ")
		}
		b.WriteString(input.View())
		b.WriteString("\n")
	}

	b.WriteString("\n")

	// État de chargement / erreur
	if s.loading {
		loadingMsg := "Connexion en cours…"
		if s.isRegister {
			loadingMsg = "Inscription en cours…"
		}
		b.WriteString(m.spinner.View() + " " + loadingMsg + "\n")
	} else if s.err != "" {
		b.WriteString(styles.ErrorStyle.Render("✗ "+s.err) + "\n")
	}

	b.WriteString("\n")

	// Barre de navigation
	nav := styles.Key("Entrée") + " valider  " +
		styles.Key("Tab") + " champ suivant  " +
		styles.Key("Ctrl+R") + " " + toggleLabel(s.isRegister) + "  " +
		styles.Key("Ctrl+C") + " quitter"
	b.WriteString(styles.NavBar.Render(nav))

	return b.String()
}

func toggleLabel(isRegister bool) string {
	if isRegister {
		return "mode connexion"
	}
	return "mode inscription"
}
