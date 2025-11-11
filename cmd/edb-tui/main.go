// Sauvegarde du fichier main.go - version du 11/11/2025
// Copie générée automatiquement

package main

import (
	"encoding/csv"
	"fmt"
	"net/http"
	"os"
	"sort"
	"strings"
	"time"

	"github.com/mattn/go-runewidth"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss/v2"
	"github.com/skip2/go-qrcode"
)

const csvURL = "https://raw.githubusercontent.com/adriens/edb-noumea-data/main/data/resume.csv"
const detailsURL = "https://raw.githubusercontent.com/adriens/edb-noumea-data/main/data/details.csv"

// Model for Bubbletea
// You can extend this with more fields for navigation, filtering, etc.
type Model struct {
	sortEcoli         bool // tri des détails selon E. coli
	sortEnte          bool // tri des détails selon Enté.
	data              [][]string
	details           [][]string
	err               error
	lastRefresh       time.Time
	nextRefresh       time.Time
	logs              []string // last actions
	showAbout         bool     // about screen toggle
	autoRefresh       bool     // pour indiquer si le refresh auto est actif
	width             int      // terminal width
	height            int      // terminal height
	selectedDetailRow int      // ligne sélectionnée dans le tableau des détails
	showLegendPopup   bool     // affiche la popup de légende
	showStatsPopup    bool     // affiche la popup de stats
}

func initialModel() Model {
	now := time.Now()
	return Model{logs: []string{}, showAbout: false, width: 80, height: 24, autoRefresh: true, lastRefresh: now, nextRefresh: now.Add(time.Hour), selectedDetailRow: 1, showLegendPopup: false, showStatsPopup: false, sortEcoli: false, sortEnte: false}
}

// Charge les deux CSV en parallèle
func fetchAllData() tea.Cmd {
	return func() tea.Msg {
		data, err := fetchCSVData(csvURL)
		if err != nil {
			return err
		}
		details, err := fetchCSVData(detailsURL)
		if err != nil {
			return err
		}
		return struct {
			data      [][]string
			details   [][]string
			fetchedAt time.Time
		}{data, details, time.Now()}
	}
}

func fetchCSVData(url string) ([][]string, error) {
	resp, err := http.Get(url)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	reader := csv.NewReader(resp.Body)
	return reader.ReadAll()
}

func (m Model) Init() tea.Cmd {
	m.nextRefresh = time.Now().Add(time.Hour)
	return tea.Batch(fetchAllData(), autoRefreshCmd())
}

// Commande Bubbletea pour le refresh auto toutes les heures
func autoRefreshCmd() tea.Cmd {
	return tea.Tick(time.Hour, func(t time.Time) tea.Msg {
		return "auto-refresh"
	})
}

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		if msg.String() == "e" {
			m.sortEcoli = !m.sortEcoli
			m.sortEnte = false // désactive le tri Enté. si activé
			return m, nil
		}
		if msg.String() == "n" {
			m.sortEnte = !m.sortEnte
			m.sortEcoli = false // désactive le tri E. coli si activé
			return m, nil
		}
		if msg.String() == "n" {
			m.sortEnte = !m.sortEnte
			m.sortEcoli = false // désactive le tri E. coli si activé
			m = m.addLog("Tri des détails par Enté.: " + map[bool]string{true: "activé", false: "désactivé"}[m.sortEnte])
			return m, nil
		}
		// ...existing code...
		if m.showAbout {
			m.showAbout = false
			return m, nil
		}
		if m.showLegendPopup {
			m.showLegendPopup = false
			return m, nil
		}
		if m.showStatsPopup {
			m.showStatsPopup = false
			return m, nil
		}
		switch msg.String() {
		case "ctrl+c", "q":
			m = m.addLog("Application quittée")
			return m, tea.Quit
		case "r":
			m.lastRefresh = time.Now()
			m.nextRefresh = m.lastRefresh.Add(time.Hour)
			m = m.addLog(fmt.Sprintf("Rafraîchissement manuel demandé. Dernier : %s. Prochain : %s.", m.lastRefresh.Format("15:04:05"), m.nextRefresh.Format("15:04:05")))
			return m, fetchAllData()
		case "a":
			m.showAbout = true
			return m, nil
		case "l":
			m.showLegendPopup = true
			return m, nil
		case "s":
			m.showStatsPopup = true
			return m, nil
		case "up":
			if len(m.details) > 1 && m.selectedDetailRow > 1 {
				m.selectedDetailRow--
			}
			return m, nil
		case "down":
			if len(m.details) > 1 && m.selectedDetailRow < len(m.details)-1 {
				m.selectedDetailRow++
			}
			return m, nil
		}
	case string:
		if msg == "auto-refresh" && m.autoRefresh {
			m.lastRefresh = time.Now()
			m.nextRefresh = m.lastRefresh.Add(time.Hour)
			m = m.addLog(fmt.Sprintf("Rafraîchissement automatique déclenché. Dernier : %s. Prochain : %s.", m.lastRefresh.Format("15:04:05"), m.nextRefresh.Format("15:04:05")))
			return m, tea.Batch(fetchAllData(), autoRefreshCmd())
		}
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		return m, nil
	case struct {
		data      [][]string
		details   [][]string
		fetchedAt time.Time
	}:
		m.data = msg.data
		m.details = msg.details
		m.lastRefresh = msg.fetchedAt
		m = m.addLog("Données rafraîchies depuis GitHub")
		return m, nil
	case [][]string:
		m.data = msg
		return m, nil
	case error:
		m.err = msg
		m = m.addLog(fmt.Sprintf("Erreur: %v", msg))
		return m, nil
	}
	return m, nil

}

// Ajoute une entrée au log, conserve les 3 dernières
func (m Model) addLog(entry string) Model {
	logs := append(m.logs, fmt.Sprintf("[%s] %s", time.Now().Format("15:04:05"), entry))
	if len(logs) > 3 {
		logs = logs[len(logs)-3:]
	}
	m.logs = logs
	return m
}

func (m Model) View() string {
	// Affichage popup stats
	if m.showStatsPopup {
		// Récupère les scores
		// Génère les histogrammes
		// Récupère les scores
		var ecoliScores []int
		var enteScores []int
		if len(m.details) > 1 {
			// Cherche les index
			ecoliIdx := -1
			enteIdx := -1
			for idx, name := range m.details[0] {
				if name == "E. coli" || name == "e_coli_npp_100ml" || name == "ec_npp_100ml" {
					ecoliIdx = idx
				}
				if name == "Enté." || name == "enterocoques_npp_100ml" || name == "ent_npp_100ml" {
					enteIdx = idx
				}
			}
			for i := 1; i < len(m.details); i++ {
				row := m.details[i]
				if ecoliIdx != -1 && ecoliIdx < len(row) {
					n := 0
					fmt.Sscanf(row[ecoliIdx], "%d", &n)
					ecoliScores = append(ecoliScores, n)
				}
				if enteIdx != -1 && enteIdx < len(row) {
					n := 0
					fmt.Sscanf(row[enteIdx], "%d", &n)
					enteScores = append(enteScores, n)
				}
			}
		}
		// Histogramme E. coli : bleu (≤500), jaune (≤1000), rouge (>1000)
		ecoliHisto := func(scores []int, seuilMax int, width int) string {
			bins := make([]int, 10)
			for _, v := range scores {
				idx := v * 10 / seuilMax
				if idx > 9 {
					idx = 9
				}
				if idx < 0 {
					idx = 0
				}
				bins[idx]++
			}
			maxBin := 1
			for _, b := range bins {
				if b > maxBin {
					maxBin = b
				}
			}
			lines := []string{}
			for i, b := range bins {
				barLen := int(float64(b) / float64(maxBin) * float64(width))
				if barLen < 1 && b > 0 {
					barLen = 1
				}
				label := fmt.Sprintf("%3d-%-3d", i*seuilMax/10, (i+1)*seuilMax/10)
				// Couleur selon la tranche
				var color string
				upper := (i + 1) * seuilMax / 10
				if upper <= 500 {
					color = "12" // bleu
				} else if upper <= 1000 {
					color = "3" // jaune
				} else {
					color = "1" // rouge
				}
				bar := lipgloss.NewStyle().Foreground(lipgloss.Color(color)).Render(strings.Repeat("█", barLen))
				lines = append(lines, fmt.Sprintf("%s | %s (%d)", label, bar, b))
			}
			return strings.Join(lines, "\n")
		}(ecoliScores, 1000, 24)

		// Histogramme Enté. : bleu (≤200), jaune (≤400), rouge (>400)
		enteHisto := func(scores []int, seuilMax int, width int) string {
			bins := make([]int, 10)
			for _, v := range scores {
				idx := v * 10 / seuilMax
				if idx > 9 {
					idx = 9
				}
				if idx < 0 {
					idx = 0
				}
				bins[idx]++
			}
			maxBin := 1
			for _, b := range bins {
				if b > maxBin {
					maxBin = b
				}
			}
			lines := []string{}
			for i, b := range bins {
				barLen := int(float64(b) / float64(maxBin) * float64(width))
				if barLen < 1 && b > 0 {
					barLen = 1
				}
				label := fmt.Sprintf("%3d-%-3d", i*seuilMax/10, (i+1)*seuilMax/10)
				var color string
				upper := (i + 1) * seuilMax / 10
				if upper <= 200 {
					color = "12" // bleu
				} else if upper <= 400 {
					color = "3" // jaune
				} else {
					color = "1" // rouge
				}
				bar := lipgloss.NewStyle().Foreground(lipgloss.Color(color)).Render(strings.Repeat("█", barLen))
				lines = append(lines, fmt.Sprintf("%s | %s (%d)", label, bar, b))
			}
			return strings.Join(lines, "\n")
		}(enteScores, 400, 24)
		statsText := "Histogramme E. coli (≤1000) :\n" + ecoliHisto + "\n\nHistogramme Enté. (≤400) :\n" + enteHisto + "\n\nAppuyez sur une touche pour fermer."
		statsPopup := lipgloss.NewStyle().Border(lipgloss.DoubleBorder()).BorderForeground(lipgloss.Color("14")).Padding(2, 4).Align(lipgloss.Left).Width(m.width / 2).Height(m.height / 2).Render(statsText)
		return lipgloss.Place(m.width, m.height, lipgloss.Center, lipgloss.Center, statsPopup)
	}
	if m.showAbout {
		aboutText := "\nDéveloppé par Adrien S.\nGitHub : https://github.com/adriens/edb-noumea-go\n\nScannez le QR code pour accéder au projet :\n"
		// Génère le QR code ASCII avec go-qrcode
		qrText := ""
		qr, err := qrcode.New("https://github.com/adriens/edb-noumea-go", qrcode.Medium)
		if err != nil {
			qrText = "[QR code non disponible]"
		} else {
			// ToString() renders the QR code as ASCII. You can use ToString(false) for a smaller version, ToString(true) for a larger one.
			qrText = qr.ToString(false)
		}
		aboutText += qrText + "\n\nAppuyez sur n'importe quelle touche pour revenir."
		aboutBox := lipgloss.NewStyle().Border(lipgloss.DoubleBorder()).BorderForeground(lipgloss.Color("10")).Padding(1, 4).Align(lipgloss.Center).Render(aboutText)
		centered := lipgloss.Place(m.width, m.height, lipgloss.Center, lipgloss.Center, aboutBox)
		return centered
	}

	if m.err != nil {
		return fmt.Sprintf("Erreur: %v\n", m.err)
	}
	if len(m.data) == 0 {
		return "Chargement des données..."
	}

	// Styles
	headerStyle := lipgloss.NewStyle().Bold(true).Padding(0, 1)
	cellStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("15")).Padding(0, 1)
	borderStyle := lipgloss.NewStyle().Border(lipgloss.NormalBorder()).BorderForeground(lipgloss.Color("8")).Margin(1, 2)
	// detailsBorderStyle removed (unused)

	// Zone d'information sur la date/heure de récupération des données
	var fetchInfo string
	if !m.lastRefresh.IsZero() {
		fetchInfo = lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("10")).Background(lipgloss.Color("8")).Padding(0, 1).Render(
			"Données récupérées depuis GitHub le "+m.lastRefresh.Format("02/01/2006 à 15:04:05")+" (source : github.com/adriens/edb-noumea-data)") + "\n\n"
	} else {
		fetchInfo = lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("10")).Background(lipgloss.Color("8")).Padding(0, 1).Render(
			"Données non encore récupérées.") + "\n\n"
	}

	// --- Tableau principal ---
	colWidths := make([]int, len(m.data[0]))
	for _, row := range m.data {
		for j, cell := range row {
			l := runewidth.StringWidth(cell)
			if l > colWidths[j] {
				colWidths[j] = l
			}
		}
	}
	var rows []string
	for rowIdx, row := range m.data {
		var cells []string
		for j, cell := range row {
			var content string
			cellWidth := runewidth.StringWidth(cell)
			pad := colWidths[j] - cellWidth
			if pad < 0 {
				pad = 0
			}
			if rowIdx == 0 {
				if cell == "plage" {
					cell = "Plage"
					cellWidth = runewidth.StringWidth(cell)
					pad = colWidths[j] - cellWidth
					if pad < 0 {
						pad = 0
					}
				}
				if cell == "etat_sanitaire" {
					cell = "Status"
					cellWidth = runewidth.StringWidth(cell)
					pad = colWidths[j] - cellWidth
					if pad < 0 {
						pad = 0
					}
				}
				content = cell + strings.Repeat(" ", pad)
				cells = append(cells, headerStyle.Render(content))
			} else {
				if m.data[0][j] == "plage" {
					content = cell + strings.Repeat(" ", pad)
					cells = append(cells, cellStyle.Render(content))
				} else if m.data[0][j] == "etat_sanitaire" && cell == "Baignade autorisée" {
					content = cell + strings.Repeat(" ", pad)
					greenStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("10")).Bold(true).Padding(0, 1)
					cells = append(cells, greenStyle.Render(content))
				} else {
					content = cell + strings.Repeat(" ", pad)
					cells = append(cells, cellStyle.Render(content))
				}
			}
		}
		rows = append(rows, strings.Join(cells, " │ "))
	}
	table := strings.Join(rows, "\n")
	table = borderStyle.Render(table)

	// --- Tableau détails ---
	detailsTable := ""
	var detailBox string
	if len(m.details) > 0 {
		// monoHeaderStyle and monoCellStyle removed (unused)

		// Trouve les index des colonnes à traiter
		removeIdx := -1
		dateIdx := -1
		heureIdx := -1
		descIdx := -1
		idPointIdx := -1
		for idx, col := range m.details[0] {
			if col == "point_de_prelevement" {
				removeIdx = idx
			}
			if col == "date" {
				dateIdx = idx
			}
			if col == "heure" {
				heureIdx = idx
			}
			if col == "desc_point_prelevement" {
				descIdx = idx
			}
			if col == "id_point_prelevement" {
				idPointIdx = idx
			}
		}

		// Filtre, fusionne et insère la colonne desc_point_prelevement après site
		filtered := make([][]string, len(m.details))
		for i, row := range m.details {
			filteredRow := make([]string, 0, len(row)-1)
			for j, cell := range row {
				if j == removeIdx || j == idPointIdx {
					continue
				}
				if j == dateIdx {
					// Fusionne date et heure
					dateVal := cell
					heureVal := ""
					if heureIdx != -1 && heureIdx < len(row) {
						heureVal = row[heureIdx]
					}
					if i == 0 {
						filteredRow = append(filteredRow, "Date")
					} else {
						filteredRow = append(filteredRow, dateVal+" "+heureVal)
					}
					continue
				}
				if j == heureIdx {
					continue // déjà fusionné
				}
				// Ajoute la colonne desc_point_prelevement juste après site, avec renommage
				if j == 0 && descIdx != -1 && descIdx < len(row) {
					if i == 0 {
						filteredRow = append(filteredRow, "Site")
						filteredRow = append(filteredRow, "Point de prélèvement")
					} else {
						// Supprime 'PLAGE DE ' au début de la colonne 'Site'
						siteVal := cell
						plagePrefix := "PLAGE DE "
						if strings.HasPrefix(strings.ToUpper(siteVal), plagePrefix) {
							siteVal = siteVal[len(plagePrefix):]
						}
						filteredRow = append(filteredRow, siteVal)
						// Met la première lettre en majuscule pour 'Point de prélèvement'
						descVal := row[descIdx]
						if len(descVal) > 0 {
							descVal = strings.ToUpper(descVal[:1]) + descVal[1:]
						}
						filteredRow = append(filteredRow, descVal)
					}
					continue
				}
				if j == descIdx {
					continue // déjà inséré
				}
				// Renomme l'en-tête 'enterocoques_npp_100ml' ou 'ent_npp_100ml' en 'Enté.'
				if i == 0 && (cell == "enterocoques_npp_100ml" || cell == "ent_npp_100ml") {
					filteredRow = append(filteredRow, "Enté.")
					continue
				}
				// Renomme l'en-tête 'ec_npp_100ml' en 'E. coli'
				if i == 0 && (cell == "e_coli_npp_100ml" || cell == "ec_npp_100ml") {
					filteredRow = append(filteredRow, "E. coli")
					continue
				}
				filteredRow = append(filteredRow, cell)
			}
			filtered[i] = filteredRow
		}

		// Tri selon E. coli si option activée
		if m.sortEcoli && len(filtered) > 1 {
			ecoliIdx := -1
			for idx, col := range filtered[0] {
				if col == "E. coli" {
					ecoliIdx = idx
					break
				}
			}
			if ecoliIdx != -1 {
				dataRows := filtered[1:]
				sort.Slice(dataRows, func(i, j int) bool {
					ni, nj := 0, 0
					fmt.Sscanf(dataRows[i][ecoliIdx], "%d", &ni)
					fmt.Sscanf(dataRows[j][ecoliIdx], "%d", &nj)
					return ni > nj
				})
				filtered = append([][]string{filtered[0]}, dataRows...)
			}
		}
		// Tri selon Enté. si option activée
		if m.sortEnte && len(filtered) > 1 {
			enteIdx := -1
			for idx, col := range filtered[0] {
				if col == "Enté." {
					enteIdx = idx
					break
				}
			}
			if enteIdx != -1 {
				dataRows := filtered[1:]
				sort.Slice(dataRows, func(i, j int) bool {
					ni, nj := 0, 0
					fmt.Sscanf(dataRows[i][enteIdx], "%d", &ni)
					fmt.Sscanf(dataRows[j][enteIdx], "%d", &nj)
					return ni > nj
				})
				filtered = append([][]string{filtered[0]}, dataRows...)
			}
		}
		// Step 1: Ensure all rows have the same number of columns
		numCols := len(filtered[0])
		for i := range filtered {
			if len(filtered[i]) < numCols {
				for k := len(filtered[i]); k < numCols; k++ {
					filtered[i] = append(filtered[i], "")
				}
			}
		}
		// Step 2: Calculate column widths using raw (unstyled) text
		dColWidths := make([]int, numCols)
		for _, row := range filtered {
			for j := 0; j < numCols; j++ {
				l := runewidth.StringWidth(row[j])
				if l > dColWidths[j] {
					dColWidths[j] = l
				}
			}
		}
		// Step 3: Render table with lipgloss styling and borders
		var dRows []string
		for rowIdx, row := range filtered {
			var dCells []string
			for j := 0; j < numCols; j++ {
				cell := row[j]
				pad := dColWidths[j] - runewidth.StringWidth(cell)
				if pad < 0 {
					pad = 0
				}
				content := cell + strings.Repeat(" ", pad)
				colName := filtered[0][j]
				// Style for header and cells
				var style lipgloss.Style
				if rowIdx == 0 {
					style = lipgloss.NewStyle().Bold(true).Padding(0, 1)
				} else if colName == "Site" && rowIdx > 0 {
					style = lipgloss.NewStyle().Foreground(lipgloss.Color("15")).Padding(0, 1)
					if rowIdx == m.selectedDetailRow {
						style = style.Background(lipgloss.Color("7")).Foreground(lipgloss.Color("0")).Bold(true).Underline(true)
					}
				} else if colName == "E. coli" {
					n := 0
					fmt.Sscanf(cell, "%d", &n)
					switch {
					case n <= 500:
						style = lipgloss.NewStyle().Foreground(lipgloss.Color("12")).Bold(true).Padding(0, 1)
					case n <= 1000:
						style = lipgloss.NewStyle().Foreground(lipgloss.Color("3")).Bold(true).Padding(0, 1)
					default:
						style = lipgloss.NewStyle().Foreground(lipgloss.Color("1")).Bold(true).Padding(0, 1)
					}
				} else if colName == "Enté." {
					n := 0
					fmt.Sscanf(cell, "%d", &n)
					switch {
					case n <= 200:
						style = lipgloss.NewStyle().Foreground(lipgloss.Color("12")).Bold(true).Padding(0, 1)
					case n <= 400:
						style = lipgloss.NewStyle().Foreground(lipgloss.Color("3")).Bold(true).Padding(0, 1)
					default:
						style = lipgloss.NewStyle().Foreground(lipgloss.Color("1")).Bold(true).Padding(0, 1)
					}
				} else if colName == "Point de prélèvement" && rowIdx > 0 {
					style = lipgloss.NewStyle().Foreground(lipgloss.Color("12")).Bold(true).Padding(0, 1)
					if rowIdx == m.selectedDetailRow {
						style = style.Background(lipgloss.Color("7")).Underline(true)
					}
				} else if colName == "Date" && rowIdx == m.selectedDetailRow {
					style = lipgloss.NewStyle().Foreground(lipgloss.Color("0")).Background(lipgloss.Color("7")).Bold(true).Underline(true).Padding(0, 1)
				} else {
					style = lipgloss.NewStyle().Foreground(lipgloss.Color("15")).Padding(0, 1)
					if rowIdx == m.selectedDetailRow {
						style = style.Background(lipgloss.Color("7")).Underline(true)
					}
				}
				dCells = append(dCells, style.Render(content))
			}
			dRows = append(dRows, "│ "+strings.Join(dCells, " │ ")+" │")
		}
		detailsTable = lipgloss.NewStyle().Border(lipgloss.DoubleBorder()).BorderForeground(lipgloss.Color("8")).Margin(0, 0).Render(strings.Join(dRows, "\n"))

		// Génère la box de détail pour la ligne sélectionnée
		if m.selectedDetailRow > 0 && m.selectedDetailRow < len(filtered) {
			detailRow := filtered[m.selectedDetailRow]
			var detailLines []string
			// Les index ne sont plus utilisés
			for i, val := range detailRow {
				label := filtered[0][i]
				line := lipgloss.NewStyle().Bold(true).Render(label) + " : " + val
				// Ajoute la barre colorée pour E. coli et Enté.
				if label == "E. coli" || label == "Enté." {
					n := 0
					fmt.Sscanf(val, "%d", &n)
					boxWidth := m.width / 2
					padding := 6
					maxBarLen := boxWidth - padding
					if maxBarLen < 8 {
						maxBarLen = 8
					}
					barLen := 0
					color := "12"
					var seuilMax int
					if label == "E. coli" {
						seuilMax = 1000
						switch {
						case n <= 500:
							color = "12"
						case n <= 1000:
							color = "3"
						default:
							color = "1"
						}
					} else if label == "Enté." {
						seuilMax = 400
						switch {
						case n <= 200:
							color = "12"
						case n <= 400:
							color = "3"
						default:
							color = "1"
						}
					}
					if n > seuilMax {
						barLen = maxBarLen
					} else if n < 0 {
						barLen = 0
					} else {
						barLen = int(float64(n) / float64(seuilMax) * float64(maxBarLen))
						if barLen < 1 {
							barLen = 1
						}
					}
					bar := lipgloss.NewStyle().Foreground(lipgloss.Color(color)).Bold(true).Align(lipgloss.Left).Width(maxBarLen).Render(strings.Repeat("━", barLen))
					// Affiche le score sur une ligne, la barre centrée juste en dessous
					detailLines = append(detailLines, line)
					detailLines = append(detailLines, bar)
					continue
				}
				detailLines = append(detailLines, line)
			}
			detailBox = lipgloss.NewStyle().Border(lipgloss.NormalBorder()).BorderForeground(lipgloss.Color("14")).Padding(1, 2).Margin(1, 0).Width(m.width / 2).Render(strings.Join(detailLines, "\n"))
		}
	}

	legendText := lipgloss.NewStyle().Bold(true).Render("E. coli") + " : Nombre de bactéries Escherichia coli pour 100ml d'eau (NPP = Nombre le Plus Probable)\n"
	legendText += lipgloss.NewStyle().Bold(true).Render("Enté.") + " : Nombre d'entérocoques pour 100ml d'eau (NPP = Nombre le Plus Probable)\n"
	legendText += "\nSeuils européens (Directive 2006/7/CE) :\n"
	legendText += "- " + lipgloss.NewStyle().Bold(true).Render("E. coli") + " : ≤ 500 (" + lipgloss.NewStyle().Foreground(lipgloss.Color("12")).Bold(true).Render("excellent") + "), "
	legendText += "≤ 1000 (" + lipgloss.NewStyle().Foreground(lipgloss.Color("3")).Bold(true).Render("passable") + "), "
	legendText += "> 1000 (" + lipgloss.NewStyle().Foreground(lipgloss.Color("1")).Bold(true).Render("baignade interdite") + ")\n"
	legendText += "- " + lipgloss.NewStyle().Bold(true).Render("Enté.") + " : ≤ 200 (" + lipgloss.NewStyle().Foreground(lipgloss.Color("12")).Bold(true).Render("excellent") + "), "
	legendText += "≤ 400 (" + lipgloss.NewStyle().Foreground(lipgloss.Color("3")).Bold(true).Render("passable") + "), "
	legendText += "> 400 (" + lipgloss.NewStyle().Foreground(lipgloss.Color("1")).Bold(true).Render("baignade interdite") + ")\n"

	// La légende n'est plus affichée dans la vue principale, uniquement en popup

	// Affichage popup légende
	if m.showLegendPopup {
		legendPopup := lipgloss.NewStyle().Border(lipgloss.DoubleBorder()).BorderForeground(lipgloss.Color("14")).Padding(2, 6).Align(lipgloss.Left).Width(m.width / 2).Height(m.height / 2).Render(legendText + "\n\nAppuyez sur une touche pour fermer.")
		return lipgloss.Place(m.width, m.height, lipgloss.Center, lipgloss.Center, legendPopup)
	}

	// Log section (affichée en dehors de la box principale)
	logInfo := fmt.Sprintf("Dernier refresh : %s | Prochain : %s", m.lastRefresh.Format("02/01/2006 15:04:05"), m.nextRefresh.Format("02/01/2006 15:04:05"))
	logBox := lipgloss.NewStyle().Border(lipgloss.NormalBorder()).BorderForeground(lipgloss.Color("8")).Padding(0, 2).Margin(0, 0).Width(m.width - 2).Render(logInfo + "\n" + strings.Join(m.logs, "\n"))

	// Phrase de présentation en haut de l'écran
	intro := "edb-noumea-tui : le premier tui Glamour en Go pour consulter la qualité des eaux de baignade à Nouméa"
	introStyle := lipgloss.NewStyle().Bold(true).Italic(true).Foreground(lipgloss.Color("11")).Background(lipgloss.Color("0")).Padding(0, 1)
	renderedIntro := introStyle.Render(intro)
	introWidth := runewidth.StringWidth(renderedIntro)
	totalWidth := m.width - 2
	padLeftIntro := (totalWidth - introWidth) / 2
	padRightIntro := totalWidth - introWidth - padLeftIntro
	centeredIntro := strings.Repeat(" ", padLeftIntro) + renderedIntro + strings.Repeat(" ", padRightIntro)

	// Titre centré façon btop
	appTitle := "Eaux de baignade - Nouméa"
	titleStyle := lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("14")).Background(lipgloss.Color("0")).Padding(0, 1)
	renderedTitle := titleStyle.Render(appTitle)
	titleWidth := runewidth.StringWidth(renderedTitle)
	padLeft := (totalWidth - titleWidth) / 2
	padRight := totalWidth - titleWidth - padLeft
	centeredTitle := strings.Repeat(" ", padLeft) + renderedTitle + strings.Repeat(" ", padRight)

	// Encapsule tout le contenu dans une box façon btop (sans la zone de log)
	mainContent := centeredIntro + "\n" + centeredTitle + "\n" + fetchInfo + table + "\n\n" + detailsTable + "\n" + detailBox + "\n[q] Quitter  [r] Rafraîchir  [a] À propos  [l] Légende  [s] Stats  [e] Trier E. coli  [n] Trier Enté.  [↑/↓] Sélection détail"
	outerBox := lipgloss.NewStyle().Border(lipgloss.DoubleBorder()).BorderForeground(lipgloss.Color("13")).Padding(1, 2).Margin(0, 0).Width(m.width - 2).Height(m.height - 4).Align(lipgloss.Center).Render(mainContent)

	// Affiche la box principale puis la zone de log en bas
	return outerBox + "\n" + logBox
}

func main() {
	// Enable full screen mode like 'top' using AltScreen
	p := tea.NewProgram(initialModel(), tea.WithAltScreen())
	if err := p.Start(); err != nil {
		fmt.Fprintf(os.Stderr, "Erreur: %v\n", err)
		os.Exit(1)
	}
}
