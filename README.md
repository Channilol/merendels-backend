# Merendels Backend

Backend API in Go per la gestione di ferie, permessi e timbrature.

## Struttura del Progetto

```
merendels-backend/
├── config/          # Configurazione database
├── models/          # Strutture dati
├── repositories/    # Interazioni con database
├── services/       # Logica di business
├── handlers/       # Gestori richieste HTTP
├── routes/         # Definizione endpoint
├── utils/          # Utility varie
├── main.go         # Entry point
└── go.mod          # Dipendenze Go
```

## Setup e Avvio

### 1. Installa Go

Assicurati di avere Go installato (versione 1.19+)

### 2. Clona e Setup

```bash
git clone [your-repo]
cd merendels-backend
go mod tidy
```

### 3. Configura Database

Modifica `config/database.go` con i tuoi parametri PostgreSQL:

```go
host := "localhost"
port := "5432"
user := "il-tuo-username"
password := "la-tua-password"
dbname := "merendels_db"
```

### 4. Avvia il Server

```bash
go run main.go
```

Il server partirà su `http://localhost:8080`

## Endpoint Disponibili

### User Roles

- `GET /api/user-roles` - Lista tutti i ruoli
- `GET /api/user-roles/:id` - Singolo ruolo per ID
- `POST /api/user-roles` - Crea nuovo ruolo

### Esempio Richiesta POST

```json
{
  "name": "Manager",
  "hierarchy_level": 2
}
```

### Health Check

- `GET /health` - Verifica stato del server

## Testare con Postman/curl

```bash
# Health check
curl http://localhost:8080/health

# Crea user role
curl -X POST http://localhost:8080/api/user-roles \
  -H "Content-Type: application/json" \
  -d '{"name": "Manager", "hierarchy_level": 2}'

# Lista user roles
curl http://localhost:8080/api/user-roles
```
