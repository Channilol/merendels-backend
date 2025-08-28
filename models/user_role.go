package models

// UserRole rappresenta il modello del record dentro la tabella user_roles
type UseRole struct {
	ID int `json:"id"`
	Name string `json:"name"`
	HierarchyLevel int `json:"hierarchy_level"`
}

// CreateUserRoleRequest - json che riceve dal front-end
type CreateUserRoleRequest struct {
	Name string `json:"name" binding:"required"`
	HierarchyLevel int `json:"hierarchy_level" binding:"required"`
}

// UserRoleResponse - risposta che il back-end manda al front-end
type UserRoleResponse struct {
	ID int `json:"id"`
	Name string `json:"name"`
	HierarchyLevel int `json:"hierarchy_level"`
}