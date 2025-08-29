package main

import (
	"log"
	"merendels-backend/config"
	"merendels-backend/repositories"
)

func main() {
	// Connessione database
	config.ConnectDatabase()
	defer config.DB.Close()

	// Crea repository
	repo := repositories.NewUserRoleRepository()

	// Test DELETE
	id := 1
	status, err := repo.Delete(id)
	if err != nil {
		log.Printf("❌ Errore DELETE: %v", err)
		return
	}

	if status {
		log.Printf("Role eliminato con successo %v", status)
	}

}