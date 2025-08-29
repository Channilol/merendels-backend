package repositories

import (
	"database/sql"
	"fmt"
	"log"
	"merendels-backend/config"
	"merendels-backend/models"
)

type UserRoleRepository struct {}

// Nuova istanza del repository
func NewUserRoleRepository() *UserRoleRepository {
	return &UserRoleRepository{}
}

// Operazioni CRUD ->

// Create fá l'INSERT nella tabella user_roles
func (r *UserRoleRepository) Create(userRole *models.UserRole) (*models.UserRole, error) {
	query := `
	INSERT INTO user_roles (name, hierarchy_level)
	VALUES ($1, $2)
	RETURNING id`

	// err Esegue la row e fa lo scan di quello che é stato creato dalla query
	err := config.DB.QueryRow(query, userRole.Name, userRole.HierarchyLevel).Scan(&userRole.ID)
	// Controllo errori
	if (err != nil) {
		return nil, fmt.Errorf("errore nella creazione di user_role: %w", err)
	}

	log.Printf("nuovo user_role creato con id %d", userRole.ID)
	return userRole, nil
}

// GetAll recupera tutti i record da user_roles
func (r *UserRoleRepository) GetAll() ([]models.UserRole, error) {
	query := `SELECT id, name, hierarchy_level FROM user_roles ORDER BY hierarchy_level`

	// Esegue la query per prendere tutti i record
	rows, err := config.DB.Query(query)
	if err != nil {
		return nil, err
	}
	//* Chiudere sempre le rows
	// defer esegue la funzione quando la funzione "wrapper" esegue il return
	defer rows.Close()

	var userRoles []models.UserRole

	// Itero su tutte le righe
	for rows.Next() {
		var userRole models.UserRole

		// Scan di ogni riga nei campi dello struct
		err := rows.Scan(&userRole.ID, &userRole.Name, &userRole.HierarchyLevel)
		if err != nil {
			return nil, err
		}
		
		// Aggiungi alla slice (lista di lunghezza indefinita)
		userRoles = append(userRoles, userRole)
	}

	// Controllo errori durante l'iterazione
	if err = rows.Err(); err != nil {
		return nil, err
	}

	return userRoles, nil
}

// GetById recupera un user_role per ID
func (r *UserRoleRepository) GetByID(id int) (*models.UserRole, error) {
	// Query per prendere la row da l'id
	query := `SELECT id, name, hierarchy_level FROM user_roles WHERE id = $1`

	var userRole models.UserRole

	// Inizializzazione di err -> esegue la query e fa lo scan del risultato puntando a userRole
	err := config.DB.QueryRow(query, id).Scan(&userRole.ID, &userRole.Name, &userRole.HierarchyLevel)
	// Controllo errori
	if err != nil {
		// Caso nessun record trovato, torna nil
		if err == sql.ErrNoRows {
			return nil,nil
		}
		// Caso errore reale, ritorna nil e l'errore
		return nil, err
	}

	return &userRole, nil
}

func (r *UserRoleRepository) GetByHierarchyLevel(level int) (*models.UserRole, error) {
	// Query per prendere la row dal livello di gerarchia
	query := `SELECT id, name, hierarchy_level FROM user_roles WHERE hierarchy_level = $1`

	var userRole models.UserRole

	// Inizializzazione di err -> esegue la query e fa lo scan del risultato puntando a userRole
	err := config.DB.QueryRow(query, level).Scan(&userRole.ID, &userRole.Name, &userRole.HierarchyLevel)
	// Controllo errori
	if err != nil {
		// Caso nessun record trovato, torna nil
		if err == sql.ErrNoRows {
			return nil, nil
		}
		// Caso errore reale, ritorna nil e l'errore
		return nil,err
	}

	return &userRole, nil
}

// Update user_roles esistente
func (r *UserRoleRepository) Update(userRole *models.UserRole) (bool, error) {
	// Query per fare l'update
	query := `UPDATE user_roles SET name = $1, hierarchy_level = $2 WHERE id = $3`

	// Execution della query
	res, err := config.DB.Exec(query, userRole.Name, userRole.HierarchyLevel, userRole.ID)
	// Controllo errori
	if err != nil {
		return false, err
	}
	// Controllo se é stato aggiornato il record
	rowsAffected, err := res.RowsAffected()
	if err != nil {
		return false, fmt.Errorf("errore nel controllare le righe aggiornate: %w", err)
	}
	if (rowsAffected == 0) {
		// Nessuna riga, non é un errore ma non é stato aggiornato nulla.
		return false, fmt.Errorf("nessun record aggiornato con id %d", userRole.ID)
	}

	// Log di record aggiornato con successo
	log.Printf("record con id %d aggiornato", userRole.ID)
	
	return true, nil
}

// Elimina user_role dal database
func (r *UserRoleRepository) Delete(id int) (bool, error) {
	query := `DELETE FROM user_roles WHERE id = $1`
	 _, err := config.DB.Exec(query, id) 

	 if err != nil {
		return false, err
	 }

	 // Log di record aggiornato con successo
	log.Printf("record con id %d eliminato", id)
	return true, nil
}