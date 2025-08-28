package models

// User rappresenta il modello del record nella tabella Users
type User struct {
	ID   int    `json:"ID"`
	Name string `json:"name"`
	Email string `json:"email"`
	RoleID *int `json:"role_id"`
	ManagerID *int `json:"manager_id"`
}

// Struct per la richiesta di creazione User dal front-end
type CreateUserRequest struct {
	Name string `json:"name" binding:"required"`
	Email string `json:"email" binding:"required"`
	RoleID *int `json:"role_id"`
	ManagerID *int `json:"manager_id"`
}

// Struct per la risposta del back-end alla richiesta di creazione dell'User del front-end
type CreateUserResponse struct {
	ID   int    `json:"ID"`
	Name string `json:"name"`
	Email string `json:"email"`
	RoleID *int `json:"role_id"`
	ManagerID *int `json:"manager_id"`
}