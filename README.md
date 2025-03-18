# MITM Proxy Service

Un service de proxy Man-in-the-Middle (MITM) écrit en Go qui intercepte, enregistre et journalise le trafic HTTP/HTTPS. Ce service utilise la bibliothèque `go-mitmproxy` pour fournir une interface web permettant de surveiller le trafic en temps réel.

## Fonctionnalités

- Interception et journalisation des requêtes HTTP/HTTPS
- Interface web pour surveiller le trafic en temps réel
- Masquage des en-têtes sensibles (par exemple, authorization, password, token)
- Exclusion de routes spécifiques de la journalisation
- Envoi des journaux à un service de journalisation externe
- Conteneurisation avec Docker pour un déploiement facile

## Prérequis

- Go 1.20 ou supérieur
- Docker et Docker Compose (pour l'exécution conteneurisée)

## Configuration

Le service peut être configuré à l'aide de variables d'environnement :

| Variable | Description | Valeur par défaut |
|----------|-------------|-------------------|
| LISTEN_ADDR | Adresse d'écoute du proxy | :8080 |
| LOGGER_ENDPOINT | Point de terminaison du service de journalisation | http://logger-service/api/logs |
| EXCLUDED_ROUTES | Routes à exclure de la journalisation (séparées par des virgules) | health,metrics |
| MASK_HEADERS | En-têtes à masquer (séparés par des virgules) | authorization,password,token,api-key |
| WEB_INTERFACE | Activer l'interface web | true |
| WEB_PORT | Port de l'interface web | 8081 |
| MAX_RETRIES | Nombre maximum de tentatives pour envoyer les journaux | 3 |
| RETRY_DELAY | Délai entre les tentatives | 500ms |

## Exécution

### Avec Docker Compose

```bash
docker-compose up -d
```

### Sans Docker

```bash
go mod download
go run main.go
```

## Utilisation

### Configuration du client

Pour utiliser le proxy MITM, configurez vos clients pour utiliser l'adresse du proxy (par défaut : http://localhost:8080).

### Accès à l'interface web

L'interface web est accessible à l'adresse http://localhost:8081 (ou le port configuré).

### Structure des journaux

Les journaux capturés contiennent les informations suivantes :

- ID unique de la requête
- ID de corrélation (pour le suivi des requêtes liées)
- Nom du client
- Utilisateur
- Horodatage
- Méthode HTTP
- URL
- En-têtes HTTP (avec masquage des informations sensibles)
- Corps de la requête
- Code de retour HTTP
- Corps de la réponse
- Temps d'exécution
- Type de journal (info, error, critical)

## Licence

MIT