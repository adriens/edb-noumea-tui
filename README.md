
[
	![Build with Taskfile.dev](https://img.shields.io/badge/build%20with-Taskfile.dev-blue?logo=task)
](https://taskfile.dev/)


# ❔ A propos

TUI pour consulter la qualité les données des eaux de baignade de Nouméa...
sans quitter son terminal... parce'que c'est cool.

## Prérequis

- Go 1.21 ou supérieur (recommandé Go 1.24)
- Avoir [`task`](https://taskfile.dev/) installé
- Accès à internet (pour télécharger le CSV)


## Compilation

Dans le dossier du TUI :
```sh
task run
```

## Exécution


```sh
# Lance l'application
./edb
```


## Dépendances principales
- [Bubbletea](https://github.com/charmbracelet/bubbletea) (TUI)
- [Lipgloss](https://github.com/charmbracelet/lipgloss) (styles)

## Source des données
- [CSV public](https://github.com/adriens/edb-noumea-data/tree/main/data)
