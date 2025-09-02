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
	query := `SELECT id, user_id, start_date, end_date, request_type, notes, created_at FROM requests WHERE user_id = $1 ORDER BY created_at DESC  LIMIT $2 OFFSET $3`
	
	rows,err := config.DB.Query(query,limit,offset)
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