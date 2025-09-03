package repositories

import (
	"database/sql"
	"fmt"
	"log"
	"merendels-backend/config"
	"merendels-backend/models"
	"time"
)

type RequestRepository struct {}

// NewRequestRepository crea una nuova istanza del repository
func NewRequestRepository() *RequestRepository {
	return &RequestRepository{}
}

// Create inserisce una nuova richiesta nel database
func (r *RequestRepository) Create(request *models.Request) (*models.Request, error) {
	query := `
		INSERT INTO requests (user_id, start_date, end_date, request_type, notes) 
		VALUES ($1, $2, $3, $4, $5) 
		RETURNING id, created_at`
	err := config.DB.QueryRow(
		query, 
		request.UserID, 
		request.StartDate, 
		request.EndDate, 
		request.RequestType, 
		request.Notes,
	).Scan(&request.ID, &request.CreatedAt)
	if err != nil {
		return nil, fmt.Errorf("errore nella creazione della richiesta: %w", err)
	}

	log.Printf("Nuova richiesta creata con ID %d per user %d", request.ID, request.UserID)
	return request, nil
}

// GetAll recupera tutte le richieste con paginazione
func (r *RequestRepository) GetAll(limit, offset int) ([]models.Request, error) {
	query := `
		SELECT id, user_id, start_date, end_date, request_type, notes, created_at 
		FROM requests 
		ORDER BY created_at DESC 
		LIMIT $1 OFFSET $2`

	rows, err := config.DB.Query(query, limit, offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var requests []models.Request

	for rows.Next() {
		var req models.Request
		err := rows.Scan(
			&req.ID,
			&req.UserID,
			&req.StartDate,
			&req.EndDate,
			&req.RequestType,
			&req.Notes,
			&req.CreatedAt,
		)
		if err != nil {
			return nil, err
		}
		requests = append(requests, req)
	}

	if err = rows.Err(); err != nil {
		return nil, err
	}

	return requests, nil
}

// GetByID recupera una richista per ID
func (r *RequestRepository) GetByID(id int) (*models.Request, error) {
	query := `SELECT id, user_id, start_date, end_date, request_type, notes, created_at FROM requests WHERE id = $1`

	var req models.Request
	err := config.DB.QueryRow(query,id).Scan(
		&req.ID,
		&req.UserID,
		&req.StartDate,
		&req.EndDate,
		&req.RequestType,
		&req.Notes,
		&req.CreatedAt,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}

	return &req, nil
}

// GetByUserID recupera tutte le richieste di un utente specifico
func (r *RequestRepository) GetByUserID(userID, limit, offset int) ([]models.Request, error) {
	query := `SELECT id, user_id, start_date, end_date, request_type, notes, created_at FROM requests WHERE user_id = $1 ORDER BY created_at DESC LIMIT $2 OFFSET $3`
	
	rows,err := config.DB.Query(query,userID,limit,offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var requests []models.Request

	for rows.Next() {
		var req models.Request
		err := rows.Scan(
			&req.ID,
			&req.UserID,
			&req.StartDate,
			&req.EndDate,
			&req.RequestType,
			&req.Notes,
			&req.CreatedAt,
		)
		if err != nil {
			return nil, err
		}
		requests = append(requests, req)
	}
	if err = rows.Err(); err != nil {
		return nil, err
	}
	return requests, nil
}

// GetByDateRange recupera richieste in un range di date specifico
func (r *RequestRepository) GetByDateRange(startDate, endDate time.Time) ([]models.Request, error) {
	query := `SELECT id, user_id, start_date, end_date, request_type, notes, created_at FROM requests WHERE (start_date <= $2 AND end_date >= $1) ORDER BY start_date ASC`

	rows, err := config.DB.Query(query,startDate,endDate)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var requests []models.Request

	for rows.Next() {
		var req models.Request
		err := rows.Scan(
			&req.ID,
			&req.UserID,
			&req.StartDate,
			&req.EndDate,
			&req.RequestType,
			&req.Notes,
			&req.CreatedAt,
		)
		if err != nil {
			return nil, err
		}
		requests = append(requests, req)
	}
	if err = rows.Err(); err != nil {
		return nil, err
	}
	return requests, nil
}

// GetByUserAndDateRange recupera richieste di un utente in un range di date
func (r *RequestRepository) GetByUserAndDateRange(userID int, startDate, endDate time.Time)  ([]models.Request, error) {
	query := `SELECT id, user_id, start_date, end_date, request_type, notes, created_at FROM requests WHERE user_id = $1 (start_date <= $3 AND end_date >= $2) ORDER BY start_date ASC`

	rows, err := config.DB.Query(query,startDate,endDate)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var requests []models.Request

	for rows.Next() {
		var req models.Request
		err := rows.Scan(
			&req.ID,
			&req.UserID,
			&req.StartDate,
			&req.EndDate,
			&req.RequestType,
			&req.Notes,
			&req.CreatedAt,
		)
		if err != nil {
			return nil, err
		}
		requests = append(requests, req)
	}
	if err = rows.Err(); err != nil {
		return nil, err
	}
	return requests, nil
}

// CheckOverlapForUser verifica se ci sono sovrapposizioni per un utente in un periodo
func (r *RequestRepository) CheckOverlapForUser(userID int, startDate, endDate time.Time, excludeID int) (bool, error) {
	query := `
		SELECT COUNT(*) 
		FROM requests 
		WHERE user_id = $1 
		AND id != $4
		AND (start_date <= $3 AND end_date >= $2)`
	var count int
	err := config.DB.QueryRow(query, userID, startDate, endDate, excludeID).Scan(&count)
	if err != nil {
		return false, err
	}

	return count > 0, nil;
}

// GetPendingRequests recupera richieste che non hanno ancora approvazioni
func (r *RequestRepository) GetPendingRequests() ([]models.Request, error) {
	query := `SELECT r.id, r.user_id, r.start_date, r.end_date, r.request_type, r.notes, r.created_at FROM requests r LEFT JOIN approvals a ON r.id = a.request_id WHERE a.id IS NULL ORDER BY r.created_at ASC`

	rows,err := config.DB.Query(query)
	if err != nil {
		return nil,err
	}
	defer rows.Close()

	var requests []models.Request

		for rows.Next() {
		var req models.Request
		err := rows.Scan(
			&req.ID,
			&req.UserID,
			&req.StartDate,
			&req.EndDate,
			&req.RequestType,
			&req.Notes,
			&req.CreatedAt,
		)
		if err != nil {
			return nil, err
		}
		requests = append(requests, req)
	}
	if err = rows.Err(); err != nil {
		return nil, err
	}
	return requests, nil
}

// Update aggiorna una richiesta esistente
func (r *RequestRepository) Update(request *models.Request) (bool, error) {
	query := `UPDATE requests SET start_date = $1, end_date = $2, request_type = $3, notes = $4 WHERE id = $5`

	result, err := config.DB.Exec(query, request.StartDate, request.EndDate, request.RequestType, request.Notes, request.ID)
	if err != nil {
		return false, err
	}
	
	 rowsAffected, err := result.RowsAffected()
	 if err != nil {
		return false, fmt.Errorf("errore nel controllare le righe aggiornate: %w", err)
	 }

	 if rowsAffected == 0 {
		return false, fmt.Errorf("nessuna richiesta aggiornata con ID %d", request.ID)
	 }

	 log.Printf("Richiesta con ID %d aggiornata", request.ID)
	 return true, nil
}

// Delete elimina una richiesta dal database
func (r *RequestRepository) Delete(id int) (bool, error) {
	// Prima elimino le approcazioni associate
	deleteApprovalsQuery := `DELETE FROM approvals WHERE request_id = $1`
	_, err := config.DB.Exec(deleteApprovalsQuery, id)
	if err != nil {
		return false, fmt.Errorf("errore nell'eliminazione delle approvazioni associate: %w", err)
	}

	// Poi elimino le richieste
	deleteRequestQuery := `DELETE FROM requests WHERE id = $1`
	result, err := config.DB.Exec(deleteRequestQuery, id)
	if err != nil {
		return false, err
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return false, fmt.Errorf("errore nel controllare le righe eliminate: %w", err)
	}

	if rowsAffected == 0 {
		return false, fmt.Errorf("nessuna richiesta eliminata con ID %d", id)
	 }

	 log.Printf("Richiesta con ID %d eliminata (incluse approvazioni associate)", id)
	return true, nil
}

// CountByUserID conta il numero totale di richieste per un utente
func (r *RequestRepository) CountByUserID(userID int) (int, error) {
	query := `SELECT COUNT(*) FROM requests WHERE user_id = $1`

	var count int
	err := config.DB.QueryRow(query, userID).Scan(&count)
	if err != nil {
		return 0, err
	}

	return count, nil
}

// CountTotal conta il numero totale di richieste nel sistema
func (r *RequestRepository) CountTotal() (int, error) {
	query := `SELECT COUNT(*) FROM requests`

	var count int
	err := config.DB.QueryRow(query).Scan(&count)
	if err != nil {
		return 0, err
	}

	return count, nil
}