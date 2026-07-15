// === [PARTIE TUI - Frontend] ===
// Point d'entrée du TUI : initialise le client HTTP (ou mock), le store local et
// démarre le programme Bubble Tea en mode altscreen.
package main

import (
	"flag"
	"fmt"
	"os"

	"github.com/KoZeuh/GO-Project-ESGI/tui/internal/client"
	"github.com/KoZeuh/GO-Project-ESGI/tui/internal/model"
	"github.com/KoZeuh/GO-Project-ESGI/tui/internal/store"
	tea "github.com/charmbracelet/bubbletea"
)

var version = "dev"

func main() {
	apiURL := flag.String("api", "http://localhost:8080", "URL de base de l'API REST")
	useMock := flag.Bool("mock", false, "Utiliser le client mock (développement hors-ligne)")
	showVersion := flag.Bool("version", false, "Affiche la version et quitte")
	flag.Parse()

	if *showVersion {
		fmt.Println(version)
		return
	}

	var c client.Client
	if *useMock {
		c = client.NewMockClient()
		fmt.Fprintln(os.Stderr, "⚠  Mode mock activé — aucun appel réseau réel")
	} else {
		c = client.NewHTTPClient(*apiURL)
	}

	s := store.New()
	m := model.New(c, s)

	p := tea.NewProgram(m, tea.WithAltScreen())
	if _, err := p.Run(); err != nil {
		fmt.Fprintf(os.Stderr, "Erreur TUI : %v\n", err)
		os.Exit(1)
	}
}
