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

// GetAllUserRoles recupera tutti i user_roles
func (s *UserRoleService) GetAllUserRoles() ([]models.UserRole, error) {
	// Chiamo la repo
	userRoles, err := s.repository.GetAll()
	if err != nil {
		return nil, fmt.Errorf("error fetching user roles: %w", err)
	}

	return userRoles, nil
}

// GetUserRoleByID recupera un user_role per ID
func (s *UserRoleService) GetUserRoleByID(id int) (*models.UserRole, error) {
	// Se id é 0 o minore torna un errore
	if id <= 0 {
		return nil, errors.New("invalid ID: must be greater than 0")
	}

	// Chiamo la repo
	userRole, err := s.repository.GetByID(id)
	if err != nil {
		return nil, fmt.Errorf("error fetching user role with id:%w", err)
	}

	// se userRole é nil c'é un'errore di business logic
	if userRole == nil {
		// Business error: non trovato
		return nil, errors.New("user role not found")
	}

	return userRole, nil
}

// UpdateUserRole modifica un user_role esistente
func (s *UserRoleService) UpdateUserRole(id int, request *models.CreateUserRoleRequest) (*models.UserRole, error) {
	// Se id é 0 o minore torna un errore
	if id <= 0 {
		return nil, errors.New("invalid ID: must be greater than 0")
	}

	if request.Name == "" {
		return nil, errors.New("name cannot be empty")
	}

	if request.HierarchyLevel < 0 {
		return nil, errors.New("invalid hierarchy level: must 0 or higher")
	}

	// Business logic: verifica se esiste
	existingRole, err := s.repository.GetByID(id)
	if err != nil {
		return nil, fmt.Errorf("user role with id:%w does not exist", id)
	}

	if existingRole == nil {
		return nil, errors.New("user role not found")
	}

	// Busioness logic: vreifica hierarchy level unico
	if request.HierarchyLevel != existingRole.HierarchyLevel {
		roleWithLevel, err := s.repository.GetByHierarchyLevel(request.HierarchyLevel)
		if err != nil {
			return nil, fmt.Errorf("error checking role with hierarchy level: %w", err)
		}
		if roleWithLevel != nil {
			return nil, errors.New("hierarchy level already existing")
		}
	}	

	// aggiorno i dati del ruolo esistente
	existingRole.Name = request.Name
	existingRole.HierarchyLevel = request.HierarchyLevel

	// salva tramite repository
	success, err := s.repository.Update(existingRole)
	if err != nil {
		return nil, fmt.Errorf("error updating user role: %w", err)
	}
	if !success {
		return nil, errors.New("failed to update user role")
	}

	return existingRole, nil
}

// DeleteUserRole elimina un user_role
func (s *UserRoleService) DeleteUserRole(id int) (bool, error) {
	// Business logic: validazione ID
	if id <= 0 {
		return false, errors.New("invalid id, must be greater than 0")
	}

	existingRole, err := s.repository.GetByID(id)
	if err != nil {
		return false, fmt.Errorf("user role with id: %w does not exist", id)
	}

	if existingRole == nil {
		return false, errors.New("error fetching user role")
	}

	// TODO: Da aggiungere la verifica se esistono users con il ruolo che si vuole eliminare, e nel caso esistano, fermare la funzione

	success, err := s.repository.Delete(id)
	if err != nil {
		return false, fmt.Errorf("error deleting user role: %w", err)
	}
	if !success {
		return false, errors.New("failed to delete user role")
	}

	return true, nil
}