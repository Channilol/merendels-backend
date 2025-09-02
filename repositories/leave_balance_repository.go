package repositories

import (
	"database/sql"
	"fmt"
	"log"
	"merendels-backend/config"
	"merendels-backend/models"
)

type LeaveBalanceRepository struct{}

// NewLeaveBalanceRepository crea una nuova istanza del repository
func NewLeaveBalanceRepository() *LeaveBalanceRepository {
	return &LeaveBalanceRepository{}
}

// Create inserisce un nuovo record di saldo ferie nel database
func (r *LeaveBalanceRepository) Create(leaveBalance *models.LeaveBalance) (*models.LeaveBalance, error) {
	query := `
		INSERT INTO leave_balance (user_id, accumulated_holidays, accumulated_permits) 
		VALUES ($1, $2, $3) 
		RETURNING id, modified_at`

	err := config.DB.QueryRow(
		query,
		leaveBalance.UserID,
		leaveBalance.AccumulatedHolidays,
		leaveBalance.AccumulatedPermits,
	).Scan(&leaveBalance.ID, &leaveBalance.ModifiedAt)

	if err != nil {
		return nil, fmt.Errorf("errore nella creazione del saldo ferie: %w", err)
	}

	log.Printf("Nuovo saldo ferie creato con ID %d per user %d", leaveBalance.ID, leaveBalance.UserID)
	return leaveBalance, nil
}

// GetByUserID recupera il saldo ferie di un utente specifico
func (r *LeaveBalanceRepository) GetByUserID(userID int) (*models.LeaveBalance, error) {
	query := `
		SELECT id, user_id, accumulated_holidays, accumulated_permits, modified_at 
		FROM leave_balance 
		WHERE user_id = $1`

	var balance models.LeaveBalance
	err := config.DB.QueryRow(query, userID).Scan(
		&balance.ID,
		&balance.UserID,
		&balance.AccumulatedHolidays,
		&balance.AccumulatedPermits,
		&balance.ModifiedAt,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil // Saldo non trovato
		}
		return nil, err
	}

	return &balance, nil
}

// GetAll recupera tutti i saldi ferie con paginazione
func (r *LeaveBalanceRepository) GetAll(limit, offset int) ([]models.LeaveBalance, error) {
	query := `
		SELECT id, user_id, accumulated_holidays, accumulated_permits, modified_at 
		FROM leave_balance 
		ORDER BY modified_at DESC 
		LIMIT $1 OFFSET $2`

	rows, err := config.DB.Query(query, limit, offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var balances []models.LeaveBalance

	for rows.Next() {
		var balance models.LeaveBalance
		err := rows.Scan(
			&balance.ID,
			&balance.UserID,
			&balance.AccumulatedHolidays,
			&balance.AccumulatedPermits,
			&balance.ModifiedAt,
		)
		if err != nil {
			return nil, err
		}
		balances = append(balances, balance)
	}

	if err = rows.Err(); err != nil {
		return nil, err
	}

	return balances, nil
}

// Update aggiorna il saldo ferie di un utente
func (r *LeaveBalanceRepository) Update(leaveBalance *models.LeaveBalance) (bool, error) {
	query := `
		UPDATE leave_balance 
		SET accumulated_holidays = $1, accumulated_permits = $2, modified_at = CURRENT_TIMESTAMP 
		WHERE user_id = $3`

	result, err := config.DB.Exec(
		query,
		leaveBalance.AccumulatedHolidays,
		leaveBalance.AccumulatedPermits,
		leaveBalance.UserID,
	)
	if err != nil {
		return false, err
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return false, fmt.Errorf("errore nel controllare le righe aggiornate: %w", err)
	}

	if rowsAffected == 0 {
		return false, fmt.Errorf("nessun saldo ferie aggiornato per user %d", leaveBalance.UserID)
	}

	log.Printf("Saldo ferie aggiornato per user %d", leaveBalance.UserID)
	return true, nil
}

// AdjustBalance modifica il saldo ferie di un utente (aggiunge o sottrae giorni)
func (r *LeaveBalanceRepository) AdjustBalance(userID int, holidaysDelta, permitsDelta float32, reason string) error {
	// Inizia transazione per atomicità
	tx, err := config.DB.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	// Recupera il saldo attuale
	var currentHolidays, currentPermits float32
	query := `SELECT accumulated_holidays, accumulated_permits FROM leave_balance WHERE user_id = $1 FOR UPDATE`
	err = tx.QueryRow(query, userID).Scan(&currentHolidays, &currentPermits)
	
	if err != nil {
		if err == sql.ErrNoRows {
			// Se non esiste un record, creane uno nuovo
			newBalance := &models.LeaveBalance{
				UserID:              userID,
				AccumulatedHolidays: holidaysDelta,
				AccumulatedPermits:  permitsDelta,
			}
			
			insertQuery := `
				INSERT INTO leave_balance (user_id, accumulated_holidays, accumulated_permits) 
				VALUES ($1, $2, $3)`
			_, err = tx.Exec(insertQuery, newBalance.UserID, newBalance.AccumulatedHolidays, newBalance.AccumulatedPermits)
			if err != nil {
				return err
			}
		} else {
			return err
		}
	} else {
		// Aggiorna il saldo esistente
		newHolidays := currentHolidays + holidaysDelta
		newPermits := currentPermits + permitsDelta

		// Verifica che il saldo non diventi negativo
		if newHolidays < 0 {
			return fmt.Errorf("saldo ferie insufficiente: tentativo di sottrarre %f da %f", -holidaysDelta, currentHolidays)
		}
		if newPermits < 0 {
			return fmt.Errorf("saldo permessi insufficiente: tentativo di sottrarre %f da %f", -permitsDelta, currentPermits)
		}

		updateQuery := `
			UPDATE leave_balance 
			SET accumulated_holidays = $1, accumulated_permits = $2, modified_at = CURRENT_TIMESTAMP 
			WHERE user_id = $3`
		_, err = tx.Exec(updateQuery, newHolidays, newPermits, userID)
		if err != nil {
			return err
		}
	}

	// Commit della transazione
	err = tx.Commit()
	if err != nil {
		return err
	}

	log.Printf("Saldo aggiustato per user %d: holidays %+.1f, permits %+.1f (reason: %s)", 
		userID, holidaysDelta, permitsDelta, reason)
	return nil
}

// DeductLeave sottrae giorni dal saldo ferie quando una richiesta viene approvata
func (r *LeaveBalanceRepository) DeductLeave(userID int, requestType models.RequestType, days float32) error {
	var holidaysDelta, permitsDelta float32

	switch requestType {
	case models.RequestHolidays:
		holidaysDelta = -days // Sottrae dalle ferie
		permitsDelta = 0
	case models.RequestPermits:
		holidaysDelta = 0
		permitsDelta = -days // Sottrae dai permessi
	default:
		return fmt.Errorf("tipo richiesta non valido: %s", requestType)
	}

	return r.AdjustBalance(userID, holidaysDelta, permitsDelta, fmt.Sprintf("Deduzione per richiesta %s di %.1f giorni", requestType, days))
}

// RestoreLeave ripristina giorni nel saldo ferie quando una richiesta viene revocata
func (r *LeaveBalanceRepository) RestoreLeave(userID int, requestType models.RequestType, days float32) error {
	var holidaysDelta, permitsDelta float32

	switch requestType {
	case models.RequestHolidays:
		holidaysDelta = days // Aggiunge alle ferie
		permitsDelta = 0
	case models.RequestPermits:
		holidaysDelta = 0
		permitsDelta = days // Aggiunge ai permessi
	default:
		return fmt.Errorf("tipo richiesta non valido: %s", requestType)
	}

	return r.AdjustBalance(userID, holidaysDelta, permitsDelta, fmt.Sprintf("Ripristino per revoca richiesta %s di %.1f giorni", requestType, days))
}

// AddAnnualLeave aggiunge il saldo annuale di ferie a un utente
func (r *LeaveBalanceRepository) AddAnnualLeave(userID int, holidayDays, permitDays float32) error {
	return r.AdjustBalance(userID, holidayDays, permitDays, "Accredito saldo annuale")
}

// GetUsersWithLowBalance trova utenti con saldo ferie basso
func (r *LeaveBalanceRepository) GetUsersWithLowBalance(holidayThreshold, permitThreshold float32) ([]models.LeaveBalance, error) {
	query := `
		SELECT id, user_id, accumulated_holidays, accumulated_permits, modified_at 
		FROM leave_balance 
		WHERE accumulated_holidays < $1 OR accumulated_permits < $2
		ORDER BY accumulated_holidays ASC, accumulated_permits ASC`

	rows, err := config.DB.Query(query, holidayThreshold, permitThreshold)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var balances []models.LeaveBalance

	for rows.Next() {
		var balance models.LeaveBalance
		err := rows.Scan(
			&balance.ID,
			&balance.UserID,
			&balance.AccumulatedHolidays,
			&balance.AccumulatedPermits,
			&balance.ModifiedAt,
		)
		if err != nil {
			return nil, err
		}
		balances = append(balances, balance)
	}

	if err = rows.Err(); err != nil {
		return nil, err
	}

	return balances, nil
}

// Delete elimina un record di saldo ferie (raramente usato)
func (r *LeaveBalanceRepository) Delete(userID int) error {
	query := `DELETE FROM leave_balance WHERE user_id = $1`
	result, err := config.DB.Exec(query, userID)
	if err != nil {
		return err
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("errore nel controllare le righe eliminate: %w", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("nessun saldo ferie eliminato per user %d", userID)
	}

	log.Printf("Saldo ferie eliminato per user %d", userID)
	return nil
}

// InitializeUserBalance inizializza il saldo ferie per un nuovo utente
func (r *LeaveBalanceRepository) InitializeUserBalance(userID int) error {
	// Saldo standard italiano: 22 giorni di ferie + 4 giorni di permesso annuali
	standardBalance := &models.LeaveBalance{
		UserID:              userID,
		AccumulatedHolidays: 22.0, // Giorni di ferie standard
		AccumulatedPermits:  4.0,  // Giorni di permesso standard
	}

	_, err := r.Create(standardBalance)
	if err != nil {
		return fmt.Errorf("errore nell'inizializzazione del saldo per user %d: %w", userID, err)
	}

	return nil
}