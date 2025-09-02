package services

import (
	"errors"
	"fmt"
	"log"
	"merendels-backend/models"
	"merendels-backend/repositories"
	"time"
)

type RequestService struct {
	requestRepository *repositories.RequestRepository
	approvalRepository *repositories.ApprovalRepository
	leaveBalanceRepository *repositories.LeaveBalanceRepository
}

// NewRequestService crea una nuova istanza del servizio
func NewRequestService() *RequestService {
	return &RequestService{
		requestRepository: repositories.NewRequestRepository(),
		approvalRepository: repositories.NewApprovalRepository(),
		leaveBalanceRepository: repositories.NewLeaveBalanceRepository(),
	}
}

// CreateRequest crea una nuova richiesta con validazioni business
func (s *RequestService) CreateRequest(userID int, request *models.CreateRequest) (*models.Request, error) {
	// Validazioni base
	if request.StartDate.After(request.EndDate) {
		return nil, errors.New("data inizio non può essere successiva alla data fine")
	}

	if request.StartDate.Before(time.Now().Truncate(24 * time.Hour)) {
		return nil, errors.New("non è possibile richiedere ferie per date passate")
	}

	if request.RequestType != models.RequestHolidays && request.RequestType != models.RequestPermits {
		return nil, errors.New("tipo richiesta non valido")
	}

	// Calcola i giorni richiesti
	days := s.calculateWorkingDays(request.StartDate, request.EndDate)
	if days <= 0 {
		return nil, errors.New("la richiesta deve coprire almeno un giorno lavorativo")
	}

	// Validazioni specifiche per tipo
	if request.RequestType == models.RequestHolidays {
		if days > 30 {
			return nil, errors.New("non è possibile richiedere più di 30 giorni consecutivi di ferie")
		}
	} else if request.RequestType == models.RequestPermits {
		if days > 5 {
			return nil, errors.New("non è possibile richiedere più di 5 giorni consecutivi di permessi")
		}
	}

	// Controlla sovrapposizioni con altre richieste dello stesso utente
	hasOverlap, err := s.requestRepository.CheckOverlapForUser(
		userID, 
		request.StartDate, 
		request.EndDate, 
		-1, // -1 perché è una nuova richiesta
	)
	if err != nil {
		return nil, fmt.Errorf("errore nel controllo sovrapposizioni: %w", err)
	}
	if hasOverlap {
		return nil, errors.New("esiste già una richiesta per questo periodo")
	}

	// Controlla il saldo ferie disponibile
	if request.RequestType == models.RequestHolidays {
		hasEnoughBalance, err := s.checkLeaveBalance(userID, days, request.RequestType)
		if err != nil {
			return nil, fmt.Errorf("errore nel controllo saldo ferie: %w", err)
		}
		if !hasEnoughBalance {
			return nil, errors.New("saldo ferie insufficiente per questa richiesta")
		}
	}
	
	// Controlla anche il saldo permessi per richieste di tipo PERMESSO
	if request.RequestType == models.RequestPermits {
		hasEnoughBalance, err := s.checkLeaveBalance(userID, days, request.RequestType)
		if err != nil {
			return nil, fmt.Errorf("errore nel controllo saldo permessi: %w", err)
		}
		if !hasEnoughBalance {
			return nil, errors.New("saldo permessi insufficiente per questa richiesta")
		}
	}

	// Crea la richiesta nel database
	newRequest := &models.Request{
		UserID:      userID,
		StartDate:   request.StartDate,
		EndDate:     request.EndDate,
		RequestType: request.RequestType,
		Notes:       request.Notes,
	}

	createdRequest, err := s.requestRepository.Create(newRequest)
	if err != nil {
		return nil, fmt.Errorf("errore nella creazione della richiesta: %w", err)
	}

	log.Printf("Richiesta creata: User %d, tipo %s, giorni %d, periodo %s - %s", 
		userID, request.RequestType, days, 
		request.StartDate.Format("2006-01-02"), 
		request.EndDate.Format("2006-01-02"))

	return createdRequest, nil
}

// GetAllRequests recupera tutte le richieste con paginazione
func (s *RequestService) GetAllRequests(limit, offset int) ([]models.Request, error) {
	if limit <= 0 || limit > 100 {
		limit = 20
	}
	if offset < 0 {
		offset = 0
	}

	requests, err := s.requestRepository.GetAll(limit, offset)
	if err != nil {
		return nil, fmt.Errorf("errore nel recupero delle richieste: %w", err)
	}

	return requests, nil
}

// GetRequestByID recupera una richiesta specifica
func (s *RequestService) GetRequestByID(id int) (*models.Request, error) {
	if id <= 0 {
		return nil, errors.New("ID richiesta non valido")
	}

	request, err := s.requestRepository.GetByID(id)
	if err != nil {
		return nil, fmt.Errorf("errore nel recupero della richiesta: %w", err)
	}

	if request == nil {
		return nil, errors.New("richiesta non trovata")
	}

	return request, nil
}

// GetUserRequests recupera tutte le richieste di un utente
func (s *RequestService) GetUserRequests(userID, limit, offset int) ([]models.Request, error) {
	if userID <= 0 {
		return nil, errors.New("ID utente non valido")
	}

	if limit <= 0 || limit > 100 {
		limit = 20
	}
	if offset < 0 {
		offset = 0
	}

	requests, err := s.requestRepository.GetByUserID(userID, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("errore nel recupero delle richieste utente: %w", err)
	}

	return requests, nil
}

// GetRequestsByDateRange recupera richieste in un range di date
func (s *RequestService) GetRequestsByDateRange(startDate, endDate time.Time) ([]models.Request, error) {
	if startDate.After(endDate) {
		return nil, errors.New("data inizio non può essere successiva alla data fine")
	}

	requests, err := s.requestRepository.GetByDateRange(startDate, endDate)
	if err != nil {
		return nil, fmt.Errorf("errore nel recupero delle richieste per range date: %w", err)
	}

	return requests, nil
}

// GetPendingRequests recupera richieste in attesa di approvazione
func (s *RequestService) GetPendingRequests() ([]models.Request, error) {
	requests, err := s.requestRepository.GetPendingRequests()
	if err != nil {
		return nil, fmt.Errorf("errore nel recupero delle richieste in attesa: %w", err)
	}

	return requests, nil
}

// UpdateRequest aggiorna una richiesta esistente
func (s *RequestService) UpdateRequest(id int, userID int, request *models.CreateRequest) (*models.Request, error) {
	if id <= 0 {
		return nil, errors.New("ID richiesta non valido")
	}

	// Verifica che la richiesta esista e appartenga all'utente
	existingRequest, err := s.requestRepository.GetByID(id)
	if err != nil {
		return nil, fmt.Errorf("errore nel recupero della richiesta esistente: %w", err)
	}
	if existingRequest == nil {
		return nil, errors.New("richiesta non trovata")
	}
	if existingRequest.UserID != userID {
		return nil, errors.New("non autorizzato a modificare questa richiesta")
	}

	// Controlla se la richiesta ha già delle approvazioni
	approvals, err := s.approvalRepository.GetByRequestID(id)
	if err != nil {
		return nil, fmt.Errorf("errore nel controllo approvazioni esistenti: %w", err)
	}
	if len(approvals) > 0 {
		return nil, errors.New("non è possibile modificare una richiesta già approvata o rifiutata")
	}

	// Validazioni come per la creazione
	if request.StartDate.After(request.EndDate) {
		return nil, errors.New("data inizio non può essere successiva alla data fine")
	}

	if request.StartDate.Before(time.Now().Truncate(24 * time.Hour)) {
		return nil, errors.New("non è possibile richiedere ferie per date passate")
	}

	// Controlla sovrapposizioni escludendo la richiesta corrente
	hasOverlap, err := s.requestRepository.CheckOverlapForUser(
		userID, 
		request.StartDate, 
		request.EndDate, 
		id, // Escludi la richiesta corrente
	)
	if err != nil {
		return nil, fmt.Errorf("errore nel controllo sovrapposizioni: %w", err)
	}
	if hasOverlap {
		return nil, errors.New("le nuove date si sovrappongono con un'altra richiesta esistente")
	}

	// Aggiorna la richiesta
	existingRequest.StartDate = request.StartDate
	existingRequest.EndDate = request.EndDate
	existingRequest.RequestType = request.RequestType
	existingRequest.Notes = request.Notes

	_, err = s.requestRepository.Update(existingRequest)
	if err != nil {
		return nil, fmt.Errorf("errore nell'aggiornamento della richiesta: %w", err)
	}

	log.Printf("Richiesta ID %d aggiornata da user %d", id, userID)
	return existingRequest, nil
}

// DeleteRequest elimina una richiesta
func (s *RequestService) DeleteRequest(id int, userID int) error {
	if id <= 0 {
		return errors.New("ID richiesta non valido")
	}

	// Verifica che la richiesta esista e appartenga all'utente
	existingRequest, err := s.requestRepository.GetByID(id)
	if err != nil {
		return fmt.Errorf("errore nel recupero della richiesta: %w", err)
	}
	if existingRequest == nil {
		return errors.New("richiesta non trovata")
	}
	if existingRequest.UserID != userID {
		return errors.New("non autorizzato a eliminare questa richiesta")
	}

	// Controlla se la richiesta è stata approvata
	approvals, err := s.approvalRepository.GetByRequestID(id)
	if err != nil {
		return fmt.Errorf("errore nel controllo approvazioni: %w", err)
	}
	for _, approval := range approvals {
		if approval.Status == models.ApprovalAccepted {
			return errors.New("non è possibile eliminare una richiesta già approvata")
		}
	}

	// Elimina la richiesta (le approvazioni vengono eliminate automaticamente dalla repository)
	_, err = s.requestRepository.Delete(id)
	if err != nil {
		return fmt.Errorf("errore nell'eliminazione della richiesta: %w", err)
	}

	log.Printf("Richiesta ID %d eliminata da user %d", id, userID)
	return nil
}

// calculateWorkingDays calcola i giorni lavorativi tra due date (esclusi weekend)
func (s *RequestService) calculateWorkingDays(startDate, endDate time.Time) int {
	count := 0
	current := startDate

	for current.Before(endDate) || current.Equal(endDate) {
		// Escludi sabato e domenica
		if current.Weekday() != time.Saturday && current.Weekday() != time.Sunday {
			count++
		}
		current = current.AddDate(0, 0, 1)
	}

	return count
}

// checkLeaveBalance controlla se l'utente ha abbastanza giorni di ferie disponibili
func (s *RequestService) checkLeaveBalance(userID int, daysRequested int, requestType models.RequestType) (bool, error) {
	// Recupera il saldo ferie dell'utente
	balance, err := s.leaveBalanceRepository.GetByUserID(userID)
	if err != nil {
		return false, fmt.Errorf("errore nel recupero saldo ferie: %w", err)
	}

	// Se l'utente non ha ancora un saldo ferie, inizializzane uno
	if balance == nil {
		log.Printf("Utente %d non ha saldo ferie, inizializzazione...", userID)
		err = s.leaveBalanceRepository.InitializeUserBalance(userID)
		if err != nil {
			return false, fmt.Errorf("errore nell'inizializzazione saldo ferie: %w", err)
		}
		
		// Recupera il saldo appena creato
		balance, err = s.leaveBalanceRepository.GetByUserID(userID)
		if err != nil {
			return false, fmt.Errorf("errore nel recupero saldo ferie dopo inizializzazione: %w", err)
		}
	}

	// Controlla il saldo in base al tipo di richiesta
	daysRequestedFloat := float32(daysRequested)
	
	switch requestType {
	case models.RequestHolidays:
		if balance.AccumulatedHolidays < daysRequestedFloat {
			log.Printf("Saldo ferie insufficiente per user %d: richiesti %.1f, disponibili %.1f", 
				userID, daysRequestedFloat, balance.AccumulatedHolidays)
			return false, nil
		}
		log.Printf("Saldo ferie OK per user %d: richiesti %.1f, disponibili %.1f", 
			userID, daysRequestedFloat, balance.AccumulatedHolidays)
		
	case models.RequestPermits:
		if balance.AccumulatedPermits < daysRequestedFloat {
			log.Printf("Saldo permessi insufficiente per user %d: richiesti %.1f, disponibili %.1f", 
				userID, daysRequestedFloat, balance.AccumulatedPermits)
			return false, nil
		}
		log.Printf("Saldo permessi OK per user %d: richiesti %.1f, disponibili %.1f", 
			userID, daysRequestedFloat, balance.AccumulatedPermits)
		
	default:
		return false, fmt.Errorf("tipo richiesta non riconosciuto: %s", requestType)
	}

	return true, nil
}

// GetRequestWithApprovals recupera una richiesta con tutte le sue approvazioni
func (s *RequestService) GetRequestWithApprovals(id int) (*repositories.RequestWithApprovals, error) {
	if id <= 0 {
		return nil, errors.New("ID richiesta non valido")
	}

	requestWithApprovals, err := s.approvalRepository.GetRequestWithApprovals(id)
	if err != nil {
		return nil, fmt.Errorf("errore nel recupero della richiesta con approvazioni: %w", err)
	}

	if requestWithApprovals == nil {
		return nil, errors.New("richiesta non trovata")
	}

	return requestWithApprovals, nil
}