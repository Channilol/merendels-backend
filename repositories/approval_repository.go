package repositories

import (
	"database/sql"
	"fmt"
	"log"
	"merendels-backend/config"
	"merendels-backend/models"
)

type ApprovalRepository struct {}

// NewApprovalRepository crea una nuova istanza del repository
func NewApprovalRepository() *ApprovalRepository {
	return &ApprovalRepository{}
}

// Create inserisce una nuova approvazione nel database
func (r *ApprovalRepository) Create(approval *models.Approval) (*models.Approval, error) {
	query := `INSERT INTO approvals (request_id, approver_id, status, comments) VALUES ($1, $2, $3, $4) 
		RETURNING id, approved_at`

	err := config.DB.QueryRow(query, approval.RequestID, approval.ApproverID, approval.Status, approval.Comments).Scan(&approval.ID, &approval.ApprovedAt)

	if err != nil {
		return nil, fmt.Errorf("errore nella creazione dell'approvazione: %w", err)
	}

	log.Printf("Nuova richiesta creata con ID %d per la richiesta %d", approval.ID, approval.RequestID)
	return approval, nil
}

// GetAll recupera tutte le approvazioni con paginazione
func (r *ApprovalRepository) GetAll(limit, offset int) ([]models.Approval, error) {
	query := `
		SELECT id, request_id, approver_id, status, comments, approved_at 
		FROM approvals 
		ORDER BY approved_at DESC 
		LIMIT $1 OFFSET $2`

		rows, err := config.DB.Query(query, limit, offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var approvals []models.Approval

	for rows.Next() {
		var approval models.Approval
		err := rows.Scan(
			&approval.ID,
			&approval.RequestID,
			&approval.ApproverID,
			&approval.Status,
			&approval.Comments,
			&approval.ApprovedAt,
		)
		if err != nil {
			return nil, err
		}
		approvals = append(approvals, approval)
	}

	if err = rows.Err(); err != nil {
		return nil, err
	}

	return approvals, nil
}

// GetByID recupera un'approvazione per ID
func (r *ApprovalRepository) GetByID(id int) (*models.Approval, error) {
	query := `
		SELECT id, request_id, approver_id, status, comments, approved_at 
		FROM approvals 
		WHERE id = $1`

	var approval models.Approval
	err := config.DB.QueryRow(query, id).Scan(
		&approval.ID,
		&approval.RequestID,
		&approval.ApproverID,
		&approval.Status,
		&approval.Comments,
		&approval.ApprovedAt,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil // Approvazione non trovata
		}
		return nil, err
	}

	return &approval, nil
}

// GetByRequestID recupera tutte le approvazioni per una richiesta specifica
func (r *ApprovalRepository) GetByRequestID(requestID int) ([]models.Approval, error) {
	query := `
		SELECT id, request_id, approver_id, status, comments, approved_at 
		FROM approvals 
		WHERE request_id = $1 
		ORDER BY approved_at ASC`

	rows, err := config.DB.Query(query, requestID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var approvals []models.Approval

	for rows.Next() {
		var approval models.Approval
		err := rows.Scan(
			&approval.ID,
			&approval.RequestID,
			&approval.ApproverID,
			&approval.Status,
			&approval.Comments,
			&approval.ApprovedAt,
		)
		if err != nil {
			return nil, err
		}
		approvals = append(approvals, approval)
	}

	if err = rows.Err(); err != nil {
		return nil, err
	}

	return approvals, nil
}

// GetByApproverID recupera tutte le approvazioni fatte da un approvatore specifico
func (r *ApprovalRepository) GetByApproverID(approverID, limit, offset int) ([]models.Approval, error) {
	query := `
		SELECT id, request_id, approver_id, status, comments, approved_at 
		FROM approvals 
		WHERE approver_id = $1 
		ORDER BY approved_at DESC 
		LIMIT $2 OFFSET $3`

	rows, err := config.DB.Query(query, approverID, limit, offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var approvals []models.Approval

	for rows.Next() {
		var approval models.Approval
		err := rows.Scan(
			&approval.ID,
			&approval.RequestID,
			&approval.ApproverID,
			&approval.Status,
			&approval.Comments,
			&approval.ApprovedAt,
		)
		if err != nil {
			return nil, err
		}
		approvals = append(approvals, approval)
	}

	if err = rows.Err(); err != nil {
		return nil, err
	}

	return approvals, nil
}

// GetByStatus recupera approvazioni per stato specifico
func (r *ApprovalRepository) GetByStatus(status models.ApprovalStatus, limit, offset int) ([]models.Approval, error) {
	query := `
		SELECT id, request_id, approver_id, status, comments, approved_at 
		FROM approvals 
		WHERE status = $1 
		ORDER BY approved_at DESC 
		LIMIT $2 OFFSET $3`

	rows, err := config.DB.Query(query, status, limit, offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var approvals []models.Approval

	for rows.Next() {
		var approval models.Approval
		err := rows.Scan(
			&approval.ID,
			&approval.RequestID,
			&approval.ApproverID,
			&approval.Status,
			&approval.Comments,
			&approval.ApprovedAt,
		)
		if err != nil {
			return nil, err
		}
		approvals = append(approvals, approval)
	}

	if err = rows.Err(); err != nil {
		return nil, err
	}

	return approvals, nil
}

// CheckExistingApproval controlla se esiste già un'approvazione per una richiesta da parte di un approvatore
func (r *ApprovalRepository) CheckExistingApproval(requestID, approverID int) (bool, error) {
	query := `
		SELECT COUNT(*) 
		FROM approvals 
		WHERE request_id = $1 AND approver_id = $2`

	var count int
	err := config.DB.QueryRow(query, requestID, approverID).Scan(&count)
	if err != nil {
		return false, err
	}

	return count > 0, nil
}

// GetRequestWithApprovals recupera una richiesta con tutte le sue approvazioni (JOIN)
func (r *ApprovalRepository) GetRequestWithApprovals(requestID int) (*RequestWithApprovals, error) {
	// Prima recuperiamo la richiesta
	requestQuery := `
		SELECT id, user_id, start_date, end_date, request_type, notes, created_at 
		FROM requests 
		WHERE id = $1`

	var req models.Request
	err := config.DB.QueryRow(requestQuery, requestID).Scan(
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
		return nil, fmt.Errorf("errore nel recuperare la richiesta: %w", err)
	}

	// Poi recuperiamo tutte le approvazioni associate
	approvals, err := r.GetByRequestID(requestID)
	if err != nil {
		return nil, fmt.Errorf("errore nel recuperare le approvazioni: %w", err)
	}

	result := &RequestWithApprovals{
		Request:   req,
		Approvals: approvals,
	}

	return result, nil
}

// Update aggiorna un'approvazione esistente (principalmente per cambiare status e commenti)
func (r *ApprovalRepository) Update(approval *models.Approval) (bool, error) {
	query := `
		UPDATE approvals 
		SET status = $1, comments = $2, approved_at = CURRENT_TIMESTAMP 
		WHERE id = $3`

	result, err := config.DB.Exec(query, approval.Status, approval.Comments, approval.ID)
	if err != nil {
		return false, err
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return false, fmt.Errorf("errore nel controllare le righe aggiornate: %w", err)
	}

	if rowsAffected == 0 {
		return false, fmt.Errorf("nessuna approvazione aggiornata con ID %d", approval.ID)
	}

	log.Printf("Approvazione con ID %d aggiornata a status %s", approval.ID, approval.Status)
	return true, nil
}

// UpdateStatus aggiorna solo lo status di un'approvazione (metodo di convenienza)
func (r *ApprovalRepository) UpdateStatus(id int, status models.ApprovalStatus, comments *string) error {
	query := `
		UPDATE approvals 
		SET status = $1, comments = $2, approved_at = CURRENT_TIMESTAMP 
		WHERE id = $3`

	result, err := config.DB.Exec(query, status, comments, id)
	if err != nil {
		return err
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("errore nel controllare le righe aggiornate: %w", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("nessuna approvazione trovata con ID %d", id)
	}

	log.Printf("Status approvazione ID %d aggiornato a %s", id, status)
	return nil
}

// Delete elimina un'approvazione dal database
func (r *ApprovalRepository) Delete(id int) (bool, error) {
	query := `DELETE FROM approvals WHERE id = $1`
	result, err := config.DB.Exec(query, id)
	if err != nil {
		return false, err
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return false, fmt.Errorf("errore nel controllare le righe eliminate: %w", err)
	}

	if rowsAffected == 0 {
		return false, fmt.Errorf("nessuna approvazione eliminata con ID %d", id)
	}

	log.Printf("Approvazione con ID %d eliminata", id)
	return true, nil
}

// DeleteByRequestID elimina tutte le approvazioni per una richiesta specifica
func (r *ApprovalRepository) DeleteByRequestID(requestID int) error {
	query := `DELETE FROM approvals WHERE request_id = $1`
	_, err := config.DB.Exec(query, requestID)
	if err != nil {
		return fmt.Errorf("errore nell'eliminazione delle approvazioni per richiesta %d: %w", requestID, err)
	}

	log.Printf("Eliminate tutte le approvazioni per richiesta ID %d", requestID)
	return nil
}

// CountByApproverID conta il numero totale di approvazioni fatte da un approvatore
func (r *ApprovalRepository) CountByApproverID(approverID int) (int, error) {
	query := `SELECT COUNT(*) FROM approvals WHERE approver_id = $1`

	var count int
	err := config.DB.QueryRow(query, approverID).Scan(&count)
	if err != nil {
		return 0, err
	}

	return count, nil
}

// CountByStatus conta le approvazioni per status
func (r *ApprovalRepository) CountByStatus(status models.ApprovalStatus) (int, error) {
	query := `SELECT COUNT(*) FROM approvals WHERE status = $1`

	var count int
	err := config.DB.QueryRow(query, status).Scan(&count)
	if err != nil {
		return 0, err
	}

	return count, nil
}

// Struct helper per query con JOIN
type RequestWithApprovals struct {
	Request   models.Request   `json:"request"`
	Approvals []models.Approval `json:"approvals"`
}