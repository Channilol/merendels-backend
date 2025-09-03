# Usa un'immagine base per Go
FROM golang:1.25-alpine AS builder

# Imposta la directory di lavoro
WORKDIR /app

# Copia i file go.mod e go.sum
COPY go.mod go.sum ./

# Scarica le dipendenze
RUN go mod download

# Copia il resto del codice sorgente
COPY . .

# Compila l'applicazione Go
RUN go build -o main .

# Usa Alpine per immagine finale più leggera
FROM alpine:latest

# Installa certificati CA
RUN apk --no-cache add ca-certificates

WORKDIR /root/

# Copia il binario compilato
COPY --from=builder /app/main .

# Espone la porta
EXPOSE 8080

# Comando per avviare l'applicazione
CMD ["./main"]