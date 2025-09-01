package services

import (
	"crypto/rand"
	"encoding/hex"
	"errors"
	"fmt"
	"log"
	"merendels-backend/models"
	"merendels-backend/repositories"
	"merendels-backend/utils"
	"strings"

	"golang.org/x/crypto/bcrypt"
)

type AuthService struct {
	authRepository *repositories.AuthRepository
	userRepository *repositories.UserRoleRepository
	// userRepository per validazione role_id
}

// NewAuthService crea una nuova istanza del servizio
func NewAuthService() *AuthService {
	return &AuthService{
		authRepository: repositories.NewAuthRepository(),
		userRepository: repositories.NewUserRoleRepository(),
	}
}

// Login verifica le credenziali e restituisce il JWT Token
func (s *AuthService) Login(request *models.LoginRequest) (*models.LoginResponse, error) {
	// Business logic: validazione input
	if request.Email == "" || request.Password == "" {
		return nil, errors.New("email and password are required")
	}

	// Normalizza email
	email := strings.ToLower(strings.TrimSpace(request.Email))

	// Recupera dati utenti con una JOIN	
	loginData, err := s.authRepository.GetUserForLogin(email)
	if err != nil {
		return nil, fmt.Errorf("error retrieving user data: %w", err)
	}

	if loginData == nil {
		// Registra tentativo fallito per email inesistente
		log.Printf("Login attempt for non-existent email: %s", email)
		return nil, errors.New("invalid email or password")
	}
	
	// Check numero di tentativi falliti negli ultimi x minuti
	failedAttempts, _ := s.authRepository.GetRecentFailedAttempts(loginData.UserID, 15)
	if failedAttempts >= 5 {
		s.authRepository.RecordLoginAttempt(loginData.UserID, models.LoginFailure)
		return nil, errors.New("too many failed attempts, please try again later")
	}

	// BUSINESS LOGIC: Verifica password con bcrypt + salt
	// QUESTA È LA FIX: aggiungi il salt alla password prima del confronto
	passwordWithSalt := request.Password + loginData.Salt
	err = bcrypt.CompareHashAndPassword([]byte(loginData.PasswordHash), []byte(passwordWithSalt))
	if err != nil {
		// Password sbagliata, registro tentativo fallito
		s.authRepository.RecordLoginAttempt(loginData.UserID, models.LoginFailure)
		log.Printf("Failed login attempt for user with ID: %d", loginData.UserID)
		return nil, errors.New("invalid email or password")
	}

	// Login avvenuto con successo, registro tentativo riuscito
	s.authRepository.RecordLoginAttempt(loginData.UserID, models.LoginSuccess)

	// Genero JWT Token
	token, err := utils.GenerateToken(loginData.UserID, loginData.Email, loginData.RoleID, loginData.HierarchyLevel)
	if err != nil {
		return nil, fmt.Errorf("error generating token: %w", err)
	}

	response := &models.LoginResponse{
		Token: token,
		User: struct {
			ID int `json:"id"`
			Name string `json:"name"`
			Email string `json:"email"`
		}{
			ID: loginData.UserID,
			Name: loginData.Name,
			Email: loginData.Email,
		},
	}

	log.Printf("Successful login for user ID: %d (%s)", loginData.UserID, loginData.Email)
	return response, nil
}

// Register crea un nuovo utente con credenziali hashate
func (s *AuthService) Register(request *models.CreateAuthCredentialRequest, userDetails *models.CreateUserRequest) (*models.LoginResponse, error) {
	// Validazioni
	if request.Password == "" {
		return nil, errors.New("password cannot be empty")
	}

	if len(request.Password) < 6 {
		return nil, errors.New("password must be at least 6 characters")
	}

	if userDetails.Email == "" || userDetails.Name == "" {
		return nil, errors.New("name and email are required")
	}

	// Normalizzazione email
	email := strings.ToLower(strings.TrimSpace(userDetails.Email))

	// Business login: verifica se l'email é unica
	exists, err := s.authRepository.CheckEmailExists(email)
	if err != nil {
		return nil, fmt.Errorf("error checking email: %w", err)
	}
	if exists {
		return nil, errors.New("email already registered")
	}

	// Business logic: valida role_id se fornito
	if userDetails.RoleID != nil && *userDetails.RoleID > 0 {
		role, err := s.userRepository.GetByID(*userDetails.RoleID)
		if err != nil {
			return nil, fmt.Errorf("error validating role: %w", err)
		}
		if role == nil {
			return nil, errors.New("invalid role_id")
		}
	}

	// Hash della password con bcrypt
	passwordHash, salt, err := s.hashPassword(request.Password)
	if err != nil {
		return nil, fmt.Errorf("error hashing password: %w", err)
	}

	// Create user model
	user := &models.User{
		Name: userDetails.Name,
		Email: email,
		RoleID: userDetails.RoleID,
		ManagerID: userDetails.ManagerID,
	}

	// Salva utente + credenziali in transazione
	err = s.authRepository.CreateUserWithCredentials(user, passwordHash, salt)

	if err != nil {
		return nil, fmt.Errorf("error creating user: %w", err)
	}

	// Auto-login dopo registrazione
	loginRequest := &models.LoginRequest{
		Email: email,
		Password: request.Password,
	}

	return s.Login(loginRequest)
}

// ChangePassword cambia la password di un utente
func (s *AuthService) ChangePassword(userID int, currentPassword, newPassword string) error {
	// Validazioni
	if newPassword == "" {
		return errors.New("password must not be empty")
	}

	if len(newPassword) < 6 {
		return errors.New("password must have at least 6 characters")
	}

	// TODO: Da implementare verifica della password corrente

	// Hash della nuova password
	passwordHash, salt, err := s.hashPassword(newPassword)
	if err != nil {
		return fmt.Errorf("error hashing new password: %w", err)
	}

	// Aggiorna nel database
	_, err = s.authRepository.UpdatePassword(userID, passwordHash, salt)
	if err != nil {
		return fmt.Errorf("error updating password: %w", err)
	}

	log.Printf("Password changed for user ID %d;", userID)
	return nil
}

// hashPassword genera hash bcrypt + salt casuale
func (s *AuthService) hashPassword(password string) (string, string, error) {
	// Genera salt casuale (16 bytes = 32 char hex)
	saltBytes := make([]byte, 16)
	_, err := rand.Read(saltBytes)
	if err != nil {
		return "", "", err
	}
	salt := hex.EncodeToString(saltBytes)

	// Hash con bcrypt (cost 12, bilanciato tra sicurezza e performance)
	passwordWithSalt := password + salt
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(passwordWithSalt), 12)
	if err != nil {
		return "", "", err
	}

	return string(hashedPassword), salt, nil

}