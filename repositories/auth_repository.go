package repositories

import (
	"database/sql"
	"merendels-backend/config"
	"merendels-backend/models"
)

type AuthRepository struct{}

// Nuova istanza della repo
func NewAuthRepository() *AuthRepository {
	return &AuthRepository{}
}

// LoginData é la struttura per i dati necessari al login con le JOIN ottimizzate
type LoginData struct {
	UserID         int    `json:"user_id"`
	Name           string `json:"name"`
	Email          string `json:"email"`
	RoleID         *int   `json:"role_id"`
	ManagerID      *int   `json:"manager_id"`
	HierarchyLevel *int   `json:"hierarchy_level"`
	PasswordHash   string `json:"-"`
	Salt           string `json:"-"`
}

// GetUserForLogin recupera tutti i dati necessari per il login con una query JOIN
func (r *AuthRepository) GetUserForLogin(email string) (*LoginData, error) {
	query := `
		SELECT 
			u.id, 
			u.name, 
			u.email, 
			u.role_id, 
			u.manager_id,
			ur.hierarchy_level,
			ac.password_hash,
			ac.salt
		FROM users u 
		LEFT JOIN user_roles ur ON u.role_id = ur.id
		JOIN auth_credentials ac ON u.id = ac.user_id  
		WHERE u.email = $1`

	var loginData LoginData

	err := config.DB.QueryRow(query, email).Scan(
		&loginData.UserID,
		&loginData.Name,
		&loginData.Email,
		&loginData.RoleID,
		&loginData.ManagerID,
		&loginData.HierarchyLevel,
		&loginData.PasswordHash,
		&loginData.Salt,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil,nil // Case user non trovato, non é un errore
		}
		return nil, err
	}

	return &loginData, nil
}

// CreateUserWithCredentials crea un nuovo utente + credenziali in una transazione
func (r *AuthRepository) CreateUserWithCredentials(user *models.User, passwordHash, salt string) error {
	// Inizio transazione per consistenza
	tx, err := config.DB.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback() // Rollback automatico se non si committiamo

	// 1. Inserisco l'utente, uso QueryRow per tornare la row e con Scan assegno l'id creato dal database a user
	userQuery := 
		`INSERT INTO users (name, email, role_id, manager_id) VALUES ($1, $2, $3, $4)
		RETURNING id`
	err = tx.QueryRow(userQuery, user.Name, user.Email, user.RoleID, user.ManagerID).Scan(&user.ID)
	if err != nil {
		return err
	}
	
	// 2. Inserisco le credenziali
	credQuery := `
		INSERT INTO auth_credentials (user_id, password_hash, salt) 
		VALUES ($1, $2, $3)`
	_, err = tx.Exec(credQuery, user.ID, passwordHash, salt)
	if err != nil {
		return err
	}	

	// 3. Commit della transazione
	return tx.Commit()
}

// UpdatePassword aggirona la password di un utente
func (r *AuthRepository) UpdatePassword(userID int, newPasswordHash string, newSalt string) (bool, error) {
	query := `UPDATE auth_credentials SET password_hash = $1, salt = $2, modified_at = CURRENT_TIMESTAMP WHERE user_id = $3`

	result, err := config.DB.Exec(query, newPasswordHash, newSalt, userID)
	if err != nil {
		return false, err
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return false, err
	}

	if rowsAffected == 0 {
		return false, sql.ErrNoRows
	}

	return true, nil
}

// RecordLoginAttempt registra un tentativo di login
func (r *AuthRepository) RecordLoginAttempt(userID int, result models.LoginResult) error {
	query := `INSERT INTO auth_login_attempts (user_id, result) VALUES ($1, $2)`

	_, err := config.DB.Exec(query, userID, result)
	return err
}

// GetRecentFailedAttempts conta i tentativi falliti recenti
func (r *AuthRepository) GetRecentFailedAttempts(userID int, minutes int) (int, error) {
	query := `SELECT COUNT(*)
	FROM auth_login_attempts
	WHERE user_id = $1
	AND result = 'FAILURE'
	AND timestamp > NOW() - $2 * INTERVAL '1 minute'`

	var count int
	err := config.DB.QueryRow(query, userID, minutes).Scan(&count)
	if err != nil {
		return 0, nil
	}

	return count, nil
}

// CheckEmailExists verifica se un email é giá registrato
func (r *AuthRepository) CheckEmailExists(email string) (bool, error) {
	query := `SELECT EXISTS(SELECT 1 FROM users WHERE email = $1)`

	var exists bool
	err := config.DB.QueryRow(query, email).Scan(&exists)
	if err != nil {
		return false, err
	}

	return exists, nil
}