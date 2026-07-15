# 📦 Gestionnaire de Stock

Application de gestion de stock pour petite boutique, développée en Go dans le cadre d'un projet d'école.

Le dépôt contient deux programmes indépendants qui communiquent via HTTP :

| Dossier | Rôle |
|---|---|
| [`api/`](api) | API REST (Gin + SQLite) : authentification, produits, fournisseurs, mouvements de stock, alertes, export |
| [`tui/`](tui) | Client en ligne de commande (Bubble Tea) qui consomme l'API pour piloter le stock depuis le terminal |

## Sommaire

- [Fonctionnalités](#fonctionnalités)
- [Stack technique](#stack-technique)
- [Arborescence du projet](#arborescence-du-projet)
- [Prérequis](#prérequis)
- [Installation](#installation)
- [Configuration de l'API](#configuration-de-lapi)
- [Lancer l'API](#lancer-lapi)
- [Lancer le client TUI](#lancer-le-client-tui)
- [Comptes de démonstration](#comptes-de-démonstration)
- [Documentation de l'API](#documentation-de-lapi)
- [Tests](#tests)
- [Build des binaires](#build-des-binaires)

## Fonctionnalités

- **Authentification** par JWT (inscription / connexion, mot de passe hashé avec bcrypt)
- **Produits** : création, consultation, mise à jour, suppression, recherche par nom/référence
- **Fournisseurs** : CRUD complet
- **Mouvements de stock** : entrées (`IN`) et sorties (`OUT`) avec note, historique filtrable par produit / type / période
- **Alertes de stock faible** : seuil par défaut global ou seuil spécifique par produit
- **Export JSON** d'un instantané complet du stock (produits, fournisseurs, mouvements), utilisé par le TUI comme sauvegarde locale
- **Client TUI** : tableau de bord, navigation clavier entre les écrans, mode `--mock` pour travailler hors-ligne sans API

## Stack technique

**API** (`api/`)
- [Go](https://go.dev) 1.26+
- [Gin](https://github.com/gin-gonic/gin) — routeur HTTP
- [modernc.org/sqlite](https://gitlab.com/cznic/sqlite) — driver SQLite 100% Go (pas de CGO requis)
- [golang-jwt/jwt](https://github.com/golang-jwt/jwt) + `golang.org/x/crypto/bcrypt` — authentification
- [godotenv](https://github.com/joho/godotenv) — chargement de la configuration depuis un fichier `.env`

**TUI** (`tui/`)
- Go 1.21+
- [Bubble Tea](https://github.com/charmbracelet/bubbletea), [Bubbles](https://github.com/charmbracelet/bubbles), [Lipgloss](https://github.com/charmbracelet/lipgloss) (écosystème [Charm](https://charm.sh))

## Arborescence du projet

```
GO-Project-ESGI/
├── api/                          # API REST
│   ├── cmd/api/main.go           # Point d'entrée : assemble repository -> service -> handler
│   ├── docs/openapi.yaml         # Spécification OpenAPI 3.0 de l'API
│   ├── internal/
│   │   ├── config/               # Chargement de la configuration (.env, variables d'env)
│   │   ├── database/             # Connexion SQLite, schéma, seed de démonstration
│   │   ├── dto/                  # Objets de requête/réponse HTTP
│   │   ├── handler/               # Contrôleurs Gin (routes, un fichier par ressource) + tests
│   │   ├── middleware/            # Middleware d'authentification JWT
│   │   ├── models/                # Entités métier (User, Product, Supplier, Movement...)
│   │   ├── repository/            # Accès aux données (SQL)
│   │   └── service/                # Logique métier + tests
│   ├── go.mod / go.sum
│
├── tui/                          # Client terminal
│   ├── cmd/tui/main.go            # Point d'entrée : flags --api / --mock, démarrage Bubble Tea
│   ├── internal/
│   │   ├── client/                # Interface Client HTTP + implémentation Mock (hors-ligne)
│   │   ├── model/                 # Écrans Bubble Tea (login, dashboard, products, suppliers, movements, alerts)
│   │   ├── store/                 # Cache local / état partagé, export vers fichier
│   │   └── styles/                # Styles Lipgloss partagés
│   ├── go.mod / go.sum
│
└── .gitignore
```

## Prérequis

- [Go](https://go.dev/dl/) 1.26 ou supérieur (l'API et le TUI sont deux modules Go séparés, chacun avec son propre `go.mod`)
- Aucune base de données externe à installer : l'API embarque SQLite (driver pur Go, sans CGO)

## Installation

```bash
git clone https://github.com/KoZeuh/GO-Project-ESGI.git
cd GO-Project-ESGI

# Dépendances de l'API
cd api && go mod download

# Dépendances du TUI
cd ../tui && go mod download
```

## Configuration de l'API

L'API se configure via des variables d'environnement, chargées automatiquement depuis un fichier `api/.env` s'il existe (sinon des valeurs par défaut s'appliquent).

Créez `api/.env` à partir de cet exemple :

```dotenv
# api/.env
PORT=8080
DB_PATH=./data/stock.db
JWT_SECRET=change-moi-en-production
JWT_EXPIRATION_HOURS=24
ENV=development
```

| Variable | Défaut | Description |
|---|---|---|
| `PORT` | `8080` | Port d'écoute du serveur HTTP |
| `DB_PATH` | `./data/stock.db` | Chemin du fichier SQLite (créé automatiquement, y compris le dossier parent) |
| `JWT_SECRET` | `dev-secret-change-me` | Clé de signature des tokens JWT — **à changer impérativement en production** |
| `JWT_EXPIRATION_HOURS` | `24` | Durée de validité d'un token, en heures |
| `ENV` | `development` | Environnement d'exécution (informatif) |

> Le fichier `api/data/` (base SQLite) et `api/.env` sont ignorés par Git (voir [`.gitignore`](.gitignore)).

## Lancer l'API

Depuis `api/` :

```bash
go run ./cmd/api
```

Au premier démarrage, le schéma SQL est appliqué et la base est peuplée avec des données de démonstration (fournisseurs, produits, mouvements) si elle est vide — voir [`internal/database/seed.go`](api/internal/database/seed.go).

Le serveur écoute par défaut sur `http://localhost:8080`. Vérification rapide :

```bash
curl http://localhost:8080/health
# {"status":"ok"}
```

## Lancer le client TUI

Depuis `tui/`, une fois l'API démarrée :

```bash
go run ./cmd/tui --api http://localhost:8080
```

Options disponibles :

| Flag | Défaut | Description |
|---|---|---|
| `--api` | `http://localhost:8080` | URL de base de l'API REST à consommer |
| `--mock` | `false` | Active un client simulé avec des données statiques, sans appel réseau (utile hors-ligne) |

```bash
# Exemple : mode hors-ligne, sans API
go run ./cmd/tui --mock
```

### Navigation dans le TUI

Une fois connecté, le tableau de bord donne accès aux différents écrans au clavier :

| Touche | Action |
|---|---|
| `p` | Produits |
| `s` | Fournisseurs |
| `m` | Mouvements |
| `a` | Alertes |
| `e` | Exporter le stock en JSON (`stock_export.json`) |
| `r` | Rafraîchir le tableau de bord |
| `q` / `Ctrl+C` | Quitter |

Chaque écran affiche sa propre barre de navigation en bas de l'écran (créer/éditer/supprimer, retour, etc.).

## Comptes de démonstration

Le seed de démonstration crée deux comptes utilisables directement dans le TUI ou via l'API :

| Utilisateur | Mot de passe |
|---|---|
| `admin` | `admin123` |
| `employee` | `employee123` |

Un nouveau compte peut aussi être créé via l'écran d'inscription du TUI (`Ctrl+R` sur l'écran de connexion) ou via `POST /api/v1/auth/register`.

## Documentation de l'API

La spécification complète (routes, schémas, codes d'erreur) est disponible au format OpenAPI 3.0 dans [`api/docs/openapi.yaml`](api/docs/openapi.yaml). Elle peut être visualisée avec [Swagger Editor](https://editor.swagger.io) (copier/coller le contenu du fichier) ou tout autre outil compatible OpenAPI 3.0.

Résumé des routes exposées (préfixe `/api/v1`, sauf `/health`) :

| Méthode | Route | Auth requise | Description |
|---|---|---|---|
| GET | `/health` | non | Vérifie que le serveur est démarré |
| POST | `/auth/register` | non | Crée un utilisateur |
| POST | `/auth/login` | non | Authentifie un utilisateur, retourne un JWT |
| GET | `/products?search=` | oui | Liste les produits (recherche optionnelle) |
| GET | `/products/:id` | oui | Détail d'un produit |
| POST | `/products` | oui | Crée un produit |
| PUT | `/products/:id` | oui | Met à jour un produit |
| DELETE | `/products/:id` | oui | Supprime un produit |
| GET | `/suppliers` | oui | Liste les fournisseurs |
| GET | `/suppliers/:id` | oui | Détail d'un fournisseur |
| POST | `/suppliers` | oui | Crée un fournisseur |
| PUT | `/suppliers/:id` | oui | Met à jour un fournisseur |
| DELETE | `/suppliers/:id` | oui | Supprime un fournisseur |
| GET | `/movements?product_id=&type=&from=&to=` | oui | Historique des mouvements (filtrable) |
| POST | `/movements` | oui | Enregistre un mouvement (`IN` ou `OUT`) |
| GET | `/alerts` | oui | Seuil par défaut + produits en stock faible |
| PUT | `/alerts/settings` | oui | Met à jour le seuil par défaut |
| GET | `/export` | oui | Export JSON complet (produits, fournisseurs, mouvements) |

Les routes protégées attendent un en-tête `Authorization: Bearer <token>`, où `<token>` est celui retourné par `/auth/login`.

Exemple de session complète en `curl` :

```bash
# 1. Connexion
TOKEN=$(curl -s -X POST http://localhost:8080/api/v1/auth/login \
  -H "Content-Type: application/json" \
  -d '{"username":"admin","password":"admin123"}' | jq -r .token)

# 2. Liste des produits
curl -s http://localhost:8080/api/v1/products \
  -H "Authorization: Bearer $TOKEN" | jq

# 3. Enregistrer une sortie de stock
curl -s -X POST http://localhost:8080/api/v1/movements \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{"product_id":1,"type":"OUT","quantity":2,"note":"Vente comptoir"}'
```

## Tests

Les tests (handlers et services) se trouvent dans `api/internal/handler` et `api/internal/service`.

```bash
cd api
go test ./...
```

## Build des binaires

```bash
# API
cd api
go build -o bin/api ./cmd/api
./bin/api

# TUI
cd ../tui
go build -o bin/tui ./cmd/tui
./bin/tui --api http://localhost:8080
```


## Installer l'application via HomeBrew

```
brew tap KoZeuh/tap
brew install stock-tui
```