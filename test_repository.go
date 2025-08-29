package main

import (
	"log"
	"merendels-backend/config"
	"merendels-backend/models"
	"merendels-backend/repositories"
)

func main() {
	// Connessione database
	config.ConnectDatabase()
	defer config.DB.Close()

	// Crea repository
	repo := repositories.NewUserRoleRepository()

	// Test 1: Crea un nuovo user role
	log.Println("🧪 Test 1: Create UserRole")
	userRole := &models.UserRole{
		Name:           "Manager",
		HierarchyLevel: 2,
	}

	createdUserRole ,err := repo.Create(userRole)
	if err != nil {
		log.Printf("❌ Errore Create: %v", err)
		return
	}
	log.Printf("✅ UserRole creato: %v", createdUserRole)

	// Test 2: Recupera per ID
	log.Println("\n🧪 Test 2: GetByID")
	found, err := repo.GetByID(userRole.ID)
	if err != nil {
		log.Printf("❌ Errore GetByID: %v", err)
		return
	}
	if found != nil {
		log.Printf("✅ UserRole trovato: %+v", *found)
	} else {
		log.Println("❌ UserRole non trovato")
	}

	// Test 3: GetAll
	log.Println("\n🧪 Test 3: GetAll")
	allRoles, err := repo.GetAll()
	if err != nil {
		log.Printf("❌ Errore GetAll: %v", err)
		return
	}
	log.Printf("✅ Trovati %d user roles:", len(allRoles))
	for _, role := range allRoles {
		log.Printf("  - ID: %d, Name: %s, Level: %d", role.ID, role.Name, role.HierarchyLevel)
	}

	log.Println("\n🎉 Tutti i test completati!")
}