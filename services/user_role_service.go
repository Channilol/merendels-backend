package services

import (
	"errors"
	"fmt"
	"merendels-backend/models"
	"merendels-backend/repositories"
)

type UserRoleService struct {
	repository *repositories.UserRoleRepository
}

// Nuova istanza del service
func NewUserRoleRepository() *UserRoleService {
	return &UserRoleService{
		repository: repositories.NewUserRoleRepository(),
	}
}

// CreateUserRole applica la logica di business e crea un nuovo user_role
func (s *UserRoleService) CreateUserRole(request *models.CreateUserRoleRequest) (*models.UserRole, error) {
	// Validazione request
	if request.Name == "" {
		return nil, errors.New("name cannot be empty")
	}
	if request.HierarchyLevel < 0 {
		return nil, errors.New("hierarchy_level can be less than 0")
	}
	// Verifica dell'unicitá di Hierarchy level
	existingRole, err := s.repository.GetByHierarchyLevel(request.HierarchyLevel)
	if err != nil {
		// Errore tecnico database
		return nil, fmt.Errorf("error checking hierarchy level: %w", err)
	}
	if existingRole != nil {
		// Livello gerarchico giá esistente
		return nil, errors.New("hierarchy level already exists")
	}
	
	// Richiesta validata e controllata, trasformo i dati nello struct
	userRole := &models.UserRole{
		Name: request.Name,
		HierarchyLevel: request.HierarchyLevel,
	}

	// Salvo nel database tramite la repository
	createdRole, err := s.repository.Create(userRole)
	if err != nil {
		return nil, fmt.Errorf("error creating user role: %w", err)
	}

	response := &models.UserRole{
		ID: createdRole.ID,
		Name: createdRole.Name,
		HierarchyLevel: createdRole.HierarchyLevel,
	}
	
	return response, nil
}

// GetAllUserRoles applica la logica di business e 