# edb-noumea-go

TUI pour consulter les données des eaux de baignade de Nouméa (CSV public).

## Prérequis
- Go 1.21 ou supérieur (recommandé Go 1.24)
- Accès à internet pour télécharger le CSV

## Installation des dépendances

Dans le dossier racine du projet :
```sh
go mod tidy
```

## Compilation

Dans le dossier du TUI :
```sh
cd cmd/edb-tui
# Compile le binaire
go build -v
```

## Exécution

Toujours dans `cmd/edb-tui` :
```sh
# Lance l'application
./edb-tui
```

Ou directement sans compilation préalable :
```sh
go run main.go
```

## Fonctionnalités
- Téléchargement automatique du CSV des eaux de baignade
- Affichage stylisé dans le terminal
- Quitter avec `q` ou `Ctrl+C`

## Dépendances principales
- [Bubbletea](https://github.com/charmbracelet/bubbletea) (TUI)
- [Lipgloss](https://github.com/charmbracelet/lipgloss) (styles)

## Source des données
- [CSV public](https://github.com/adriens/edb-noumea-data/tree/main/data)

---

Pour toute amélioration ou bug, ouvrez une issue sur le dépôt.
