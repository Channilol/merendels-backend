package models

import "time"

type LoginResult string

const (
	LoginSuccess LoginResult = "SUCCESS"
	LoginFailure LoginResult = "FAILURE"
)

// DATI SENSIBILI - Solo per uso interno al back-end
type AuthCredential struct {
	ID           int    `json:"id"`
	UserID       int    `json:"user_id"`
	PasswordHash string `json:"-"`
	Salt         string `json:"-"`
	CreatedAt    time.Time `json:"created_at"`
	ModifiedAt    time.Time `json:"modified_at"`
}
 
// Request front-end -> back-end
type CreateAuthCredentialRequest struct {
	UserID int `json:"user_id" binding:"required"`
	Password string `json:"password" binding:"required"`

}

// Log del tentativo di login
type AuthLoginAttempt struct {
	ID           int    `json:"id"`
	UserID       int    `json:"user_id"`
	Timestamp time.Time `json:"timestamp"`
	Result LoginResult `json:"result"`
}

// Input login dal front-end
type LoginRequest struct {
	Email string `json:"email" binding:"required"`
	Password string `json:"password" binding:"required"`
}

// Response dal back-end, NO PASSWORD & SALT
type LoginResponse struct {
	Token string `json:"token"` //* Token JWT
	User struct {
		ID int `json:"id"`
		Name string `json:"name"`
		Email string `json:"email"`
	} `json:"user"`
}
