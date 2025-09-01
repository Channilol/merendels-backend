package main

import (
	"fmt"
	"log"
	"merendels-backend/config"
	"merendels-backend/models"
	"merendels-backend/services"
	"time"
)

func main() {
	log.Println("🧪 Test Auth Service")

	// Connessione database
	config.ConnectDatabase()
	defer config.DB.Close()

	// Crea service
	authService := services.NewAuthService()

	// Test 1: Registrazione nuovo utente
	log.Println("\n👤 Test 1: Register New User")
	
	// Usa timestamp per email univoca
	timestamp := fmt.Sprintf("%d", time.Now().Unix())
	testEmail := fmt.Sprintf("test-user-%s@example.com", timestamp)
	
	// Crea richiesta registrazione
	authRequest := &models.CreateAuthCredentialRequest{
		UserID:   0, // Sarà ignorato, viene auto-generato
		Password: "password123",
	}

	userRequest := &models.CreateUserRequest{
		Name:      "Luigi Verdi",
		Email:     testEmail,
		RoleID:    nil, // Nessun ruolo per ora
		ManagerID: nil, // Nessun manager per ora
	}

	registerResponse, err := authService.Register(authRequest, userRequest)
	if err != nil {
		log.Printf("❌ Errore registrazione: %v", err)
		return
	}

	log.Printf("✅ Utente registrato con successo!")
	log.Printf("   Token: %s", registerResponse.Token[:50]+"...")
	log.Printf("   User ID: %d", registerResponse.User.ID)
	log.Printf("   Name: %s", registerResponse.User.Name)
	log.Printf("   Email: %s", registerResponse.User.Email)

	// Test 2: Login con credenziali corrette
	log.Println("\n🔐 Test 2: Login Success")
	
	loginRequest := &models.LoginRequest{
		Email:    testEmail,  // ← Usa la stessa email del test 1
		Password: "password123",
	}

	loginResponse, err := authService.Login(loginRequest)
	if err != nil {
		log.Printf("❌ Errore login: %v", err)
		return
	}

	log.Printf("✅ Login effettuato con successo!")
	log.Printf("   Token: %s", loginResponse.Token[:50]+"...")
	log.Printf("   User ID: %d", loginResponse.User.ID)

	// Test 3: Login con password sbagliata
	log.Println("\n❌ Test 3: Login Wrong Password")
	
	wrongRequest := &models.LoginRequest{
		Email:    testEmail,  // ← Usa la stessa email del test 1
		Password: "password-sbagliata",
	}

	_, err = authService.Login(wrongRequest)
	if err != nil {
		log.Printf("✅ Login fallito correttamente: %v", err)
	} else {
		log.Printf("❌ Login doveva fallire!")
	}

	// Test 4: Login con email inesistente
	log.Println("\n❌ Test 4: Login Non-existent Email")
	
	nonExistentRequest := &models.LoginRequest{
		Email:    "nonexistent@test.com",
		Password: "password123",
	}

	_, err = authService.Login(nonExistentRequest)
	if err != nil {
		log.Printf("✅ Login fallito correttamente: %v", err)
	} else {
		log.Printf("❌ Login doveva fallire!")
	}

	// Test 5: Registrazione email duplicata
	log.Println("\n❌ Test 5: Duplicate Email Registration")
	
	duplicateUserRequest := &models.CreateUserRequest{
		Name:  "Mario Clone",
		Email: testEmail, // ← Stessa email per testare duplicato
	}

	_, err = authService.Register(authRequest, duplicateUserRequest)
	if err != nil {
		log.Printf("✅ Registrazione fallita correttamente: %v", err)
	} else {
		log.Printf("❌ Registrazione doveva fallire!")
	}

	// Test 6: Change Password
	log.Println("\n🔑 Test 6: Change Password")
	
	err = authService.ChangePassword(registerResponse.User.ID, "password123", "nuova-password")
	if err != nil {
		log.Printf("❌ Errore cambio password: %v", err)
	} else {
		log.Printf("✅ Password cambiata con successo!")
		
		// Test login con nuova password
		newLoginRequest := &models.LoginRequest{
			Email:    testEmail,  // ← Usa la stessa email
			Password: "nuova-password",
		}
		
		_, err = authService.Login(newLoginRequest)
		if err != nil {
			log.Printf("❌ Login con nuova password fallito: %v", err)
		} else {
			log.Printf("✅ Login con nuova password riuscito!")
		}
	}

	log.Println("\n🎉 Tutti i test auth completati!")
}