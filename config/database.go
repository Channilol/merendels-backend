package config

import (
	"database/sql"
	"fmt"
	"log"

	_ "github.com/lib/pq"
)

// DB -> maiuscola quindi pubblica, *sql.DB dichiarazione della variabile che contiente nil (puntatore vuoto)
var DB *sql.DB

// Funzione per aprire la connessione al DB PostgreSQL
func ConnectDatabase() {
	// Parametri di connessione
	host := "localhost"
	port := 5432
	user := "postgres"
	password := "1234"
	dbname := "merendels_db"

	// Stringa di connessione PostgreSQL
	psqlInfo := fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s sslmode=disable", host, port, user, password, dbname)

	db, err := sql.Open("postgres", psqlInfo)
	if err != nil {
		log.Fatal("Error opening database connection:", err)
	} 

	// Test della connessione
	err = db.Ping()
	if err != nil {
		log.Fatal("Error pinging database:", err)
	}

	// La connessione é stata stabilita e lo pinga correttamente, imposta la variabile glovale
	DB = db

	log.Printf("Database connected successfully!")
}