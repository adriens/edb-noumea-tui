[![Build with Taskfile.dev](https://img.shields.io/badge/build%20with-Taskfile.dev-blue?logo=task)](https://taskfile.dev/)
[![Live CSV Data](https://img.shields.io/badge/GitHub-Live%20CSV%20Data-black?logo=github)](https://github.com/adriens/edb-noumea-data)

![Demo TUI](./assets/edb-tui-demo.gif)


# ❔ A propos

TUI pour consulter la qualité les données des eaux de baignade de Nouméa...
sans quitter son terminal... parce'que c'est cool.

## Prérequis

- Go 1.23 ou supérieur (recommandé Go 1.24)
- Avoir [`task`](https://taskfile.dev/) installé
- Accès à internet (pour télécharger le CSV)


## Compilation

Dans le dossier du TUI :
```sh
task
```

## Exécution


```sh
./edb
```


## Dépendances principales

- [Bubbletea](https://github.com/charmbracelet/bubbletea) (TUI)
- [Lipgloss](https://github.com/charmbracelet/lipgloss) (styles)

## Source des données

- [`github.com/adriens/edb-noumea-data`](https://github.com/adriens/edb-noumea-data)
