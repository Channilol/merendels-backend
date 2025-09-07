package services

import (
	"errors"
	"fmt"
	"log"
	"merendels-backend/models"
	"merendels-backend/repositories"
)

type ApprovalService struct {
	approvalRepository *repositories.ApprovalRepository
	requestRepository  *repositories.RequestRepository
	userRepository     *repositories.UserRoleRepository
}

// NewApprovalService crea una nuova istanza del servizio
func NewApprovalService() *ApprovalService {
	return &ApprovalService{
		approvalRepository: repositories.NewApprovalRepository(),
		requestRepository:  repositories.NewRequestRepository(),
		userRepository:     repositories.NewUserRoleRepository(),
	}
}

// CreateApproval crea una nuova approvazione con validazioni business
func (s *ApprovalService) CreateApproval(approverID int, request *models.CreateApprovalRequest) (*models.Approval, error) {
	// Validazioni base
	if request.RequestID <= 0 {
		return nil, errors.New("ID richiesta non valido")
	}

	if request.Status != models.ApprovalAccepted && 
	   request.Status != models.ApprovalRejected && 
	   request.Status != models.ApprovalRevoked {
		return nil, errors.New("status approvazione non valido")
	}

	// Verifica che la richiesta esista
	existingRequest, err := s.requestRepository.GetByID(request.RequestID)
	if err != nil {
		return nil, fmt.Errorf("errore nel controllo della richiesta: %w", err)
	}
	if existingRequest == nil {
		return nil, errors.New("richiesta non trovata")
	}

	// Verifica che l'approvatore non sia lo stesso utente della richiesta
	if existingRequest.UserID == approverID {
		return nil, errors.New("non è possibile approvare le proprie richieste")
	}

	// Verifica che non esista già un'approvazione per questa richiesta da questo approvatore
	hasExisting, err := s.approvalRepository.CheckExistingApproval(request.RequestID, approverID)
	if err != nil {
		return nil, fmt.Errorf("errore nel controllo approvazioni esistenti: %w", err)
	}
	if hasExisting {
		return nil, errors.New("hai già dato un'approvazione per questa richiesta")
	}

	// Business logic: un'approvazione rifiutata chiude la richiesta
	if request.Status == models.ApprovalRejected {
		log.Printf("Richiesta ID %d rifiutata da approver %d", request.RequestID, approverID)
	}

	// Crea l'approvazione
	newApproval := &models.Approval{
		RequestID:   request.RequestID,
		ApproverID:  approverID,
		Status:      request.Status,
		Comments:    request.Comments,
	}

	createdApproval, err := s.approvalRepository.Create(newApproval)
	if err != nil {
		return nil, fmt.Errorf("errore nella creazione dell'approvazione: %w", err)
	}

	// Log per audit
	log.Printf("Approvazione creata: Approver %d, Request %d, Status %s", 
		approverID, request.RequestID, request.Status)

	// TODO: Implementare notifiche all'utente della richiesta

	return createdApproval, nil
}

// GetAllApprovals recupera tutte le approvazioni con paginazione
func (s *ApprovalService) GetAllApprovals(limit, offset int) ([]models.Approval, error) {
	if limit <= 0 || limit > 100 {
		limit = 20
	}
	if offset < 0 {
		offset = 0
	}

	approvals, err := s.approvalRepository.GetAll(limit, offset)
	if err != nil {
		return nil, fmt.Errorf("errore nel recupero delle approvazioni: %w", err)
	}

	return approvals, nil
}

// GetApprovalByID recupera un'approvazione specifica
func (s *ApprovalService) GetApprovalByID(id int) (*models.Approval, error) {
	if id <= 0 {
		return nil, errors.New("ID approvazione non valido")
	}

	approval, err := s.approvalRepository.GetByID(id)
	if err != nil {
		return nil, fmt.Errorf("errore nel recupero dell'approvazione: %w", err)
	}

	if approval == nil {
		return nil, errors.New("approvazione non trovata")
	}

	return approval, nil
}

// GetApprovalsByRequestID recupera tutte le approvazioni per una richiesta
func (s *ApprovalService) GetApprovalsByRequestID(requestID int) ([]models.Approval, error) {
	if requestID <= 0 {
		return nil, errors.New("ID richiesta non valido")
	}

	approvals, err := s.approvalRepository.GetByRequestID(requestID)
	if err != nil {
		return nil, fmt.Errorf("errore nel recupero delle approvazioni per richiesta: %w", err)
	}

	return approvals, nil
}

// GetApprovalsByApproverID recupera tutte le approvazioni fatte da un approvatore
func (s *ApprovalService) GetApprovalsByApproverID(approverID, limit, offset int) ([]models.Approval, error) {
	if approverID <= 0 {
		return nil, errors.New("ID approvatore non valido")
	}

	if limit <= 0 || limit > 100 {
		limit = 20
	}
	if offset < 0 {
		offset = 0
	}

	approvals, err := s.approvalRepository.GetByApproverID(approverID, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("errore nel recupero delle approvazioni dell'approvatore: %w", err)
	}

	return approvals, nil
}

// GetApprovalsByStatus recupera approvazioni per status
func (s *ApprovalService) GetApprovalsByStatus(status models.ApprovalStatus, limit, offset int) ([]models.Approval, error) {
	if status != models.ApprovalAccepted && 
	   status != models.ApprovalRejected && 
	   status != models.ApprovalRevoked {
		return nil, errors.New("status approvazione non valido")
	}

	if limit <= 0 || limit > 100 {
		limit = 20
	}
	if offset < 0 {
		offset = 0
	}

	approvals, err := s.approvalRepository.GetByStatus(status, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("errore nel recupero delle approvazioni per status: %w", err)
	}

	return approvals, nil
}

// UpdateApprovalStatus aggiorna lo status di un'approvazione esistente
func (s *ApprovalService) UpdateApprovalStatus(id int, approverID int, status models.ApprovalStatus, comments *string) (*models.Approval, error) {
	if id <= 0 {
		return nil, errors.New("ID approvazione non valido")
	}

	if status != models.ApprovalAccepted && 
	   status != models.ApprovalRejected && 
	   status != models.ApprovalRevoked {
		return nil, errors.New("status approvazione non valido")
	}

	// Verifica che l'approvazione esista e appartenga all'approvatore
	existingApproval, err := s.approvalRepository.GetByID(id)
	if err != nil {
		return nil, fmt.Errorf("errore nel recupero dell'approvazione: %w", err)
	}
	if existingApproval == nil {
		return nil, errors.New("approvazione non trovata")
	}
	if existingApproval.ApproverID != approverID {
		return nil, errors.New("non autorizzato a modificare questa approvazione")
	}

	// Business logic: non permettere di cambiare un'approvazione già accettata
	if existingApproval.Status == models.ApprovalAccepted && status != models.ApprovalRevoked {
		return nil, errors.New("non è possibile modificare un'approvazione già accettata (solo revoca)")
	}

	// Aggiorna lo status
	err = s.approvalRepository.UpdateStatus(id, status, comments)
	if err != nil {
		return nil, fmt.Errorf("errore nell'aggiornamento dello status: %w", err)
	}

	// Recupera l'approvazione aggiornata
	updatedApproval, err := s.approvalRepository.GetByID(id)
	if err != nil {
		return nil, fmt.Errorf("errore nel recupero dell'approvazione aggiornata: %w", err)
	}

	log.Printf("Approvazione ID %d aggiornata da %s a %s da approver %d", 
		id, existingApproval.Status, status, approverID)

	return updatedApproval, nil
}

// RevokeApproval revoca un'approvazione esistente (solo per approvazioni accettate)
func (s *ApprovalService) RevokeApproval(id int, approverID int, reason string) (*models.Approval, error) {
	if id <= 0 {
		return nil, errors.New("ID approvazione non valido")
	}

	// Verifica che l'approvazione esista e sia accettata
	existingApproval, err := s.approvalRepository.GetByID(id)
	if err != nil {
		return nil, fmt.Errorf("errore nel recupero dell'approvazione: %w", err)
	}
	if existingApproval == nil {
		return nil, errors.New("approvazione non trovata")
	}
	if existingApproval.ApproverID != approverID {
		return nil, errors.New("non autorizzato a revocare questa approvazione")
	}
	if existingApproval.Status != models.ApprovalAccepted {
		return nil, errors.New("è possibile revocare solo approvazioni accettate")
	}

	// Revoca l'approvazione
	comments := &reason
	if reason == "" {
		defaultReason := "Approvazione revocata"
		comments = &defaultReason
	}

	return s.UpdateApprovalStatus(id, approverID, models.ApprovalRevoked, comments)
}

// DeleteApproval elimina un'approvazione (solo per admin o in casi eccezionali)
func (s *ApprovalService) DeleteApproval(id int, approverID int) error {
	if id <= 0 {
		return errors.New("ID approvazione non valido")
	}

	// Verifica che l'approvazione esista
	existingApproval, err := s.approvalRepository.GetByID(id)
	if err != nil {
		return fmt.Errorf("errore nel recupero dell'approvazione: %w", err)
	}
	if existingApproval == nil {
		return errors.New("approvazione non trovata")
	}

	// Business logic: solo l'approvatore originale può eliminare la sua approvazione
	if existingApproval.ApproverID != approverID {
		return errors.New("non autorizzato a eliminare questa approvazione")
	}

	// Non permettere eliminazione di approvazioni accettate
	if existingApproval.Status == models.ApprovalAccepted {
		return errors.New("non è possibile eliminare un'approvazione accettata (usa revoca)")
	}

	_, err = s.approvalRepository.Delete(id)
	if err != nil {
		return fmt.Errorf("errore nell'eliminazione dell'approvazione: %w", err)
	}

	log.Printf("Approvazione ID %d eliminata da approver %d", id, approverID)
	return nil
}

// GetRequestApprovalStatus restituisce lo stato di approvazione di una richiesta
func (s *ApprovalService) GetRequestApprovalStatus(requestID int) (*RequestApprovalStatus, error) {
	if requestID <= 0 {
		return nil, errors.New("ID richiesta non valido")
	}

	approvals, err := s.approvalRepository.GetByRequestID(requestID)
	if err != nil {
		return nil, fmt.Errorf("errore nel recupero delle approvazioni: %w", err)
	}

	status := &RequestApprovalStatus{
		RequestID:        requestID,
		TotalApprovals:   len(approvals),
		AcceptedCount:    0,
		RejectedCount:    0,
		RevokedCount:     0,
		HasAccepted:      false,
		HasRejected:      false,
		HasRevoked:       false,
		FinalStatus:      "PENDING",
		Approvals:        approvals,
	}

	// Analizza gli status
	for _, approval := range approvals {
		switch approval.Status {
		case models.ApprovalAccepted:
			status.AcceptedCount++
			status.HasAccepted = true
		case models.ApprovalRejected:
			status.RejectedCount++
			status.HasRejected = true
		case models.ApprovalRevoked:
			status.RevokedCount++
			status.HasRevoked = true
		}
	}

	// Determina lo status finale
	if status.HasRejected {
		status.FinalStatus = "REJECTED"
	} else if status.HasAccepted && !status.HasRevoked {
		status.FinalStatus = "APPROVED"
	} else if status.HasRevoked {
		status.FinalStatus = "REVOKED"
	}

	return status, nil
}

// GetApprovalStatistics restituisce statistiche sulle approvazioni
func (s *ApprovalService) GetApprovalStatistics() (*ApprovalStatistics, error) {
	acceptedCount, err := s.approvalRepository.CountByStatus(models.ApprovalAccepted)
	if err != nil {
		return nil, fmt.Errorf("errore nel conteggio approvazioni accettate: %w", err)
	}

	rejectedCount, err := s.approvalRepository.CountByStatus(models.ApprovalRejected)
	if err != nil {
		return nil, fmt.Errorf("errore nel conteggio approvazioni rifiutate: %w", err)
	}

	revokedCount, err := s.approvalRepository.CountByStatus(models.ApprovalRevoked)
	if err != nil {
		return nil, fmt.Errorf("errore nel conteggio approvazioni revocate: %w", err)
	}

	stats := &ApprovalStatistics{
		TotalApprovals: acceptedCount + rejectedCount + revokedCount,
		AcceptedCount:  acceptedCount,
		RejectedCount:  rejectedCount,
		RevokedCount:   revokedCount,
	}

	if stats.TotalApprovals > 0 {
		stats.AcceptanceRate = float64(acceptedCount) / float64(stats.TotalApprovals) * 100
		stats.RejectionRate = float64(rejectedCount) / float64(stats.TotalApprovals) * 100
		stats.RevocationRate = float64(revokedCount) / float64(stats.TotalApprovals) * 100
	}

	return stats, nil
}

// Struct helper per lo status di approvazione di una richiesta
type RequestApprovalStatus struct {
	RequestID        int                `json:"request_id"`
	TotalApprovals   int                `json:"total_approvals"`
	AcceptedCount    int                `json:"accepted_count"`
	RejectedCount    int                `json:"rejected_count"`
	RevokedCount     int                `json:"revoked_count"`
	HasAccepted      bool               `json:"has_accepted"`
	HasRejected      bool               `json:"has_rejected"`
	HasRevoked       bool               `json:"has_revoked"`
	FinalStatus      string             `json:"final_status"` // PENDING, APPROVED, REJECTED, REVOKED
	Approvals        []models.Approval  `json:"approvals"`
}

// Struct helper per statistiche approvazioni
type ApprovalStatistics struct {
	TotalApprovals  int     `json:"total_approvals"`
	AcceptedCount   int     `json:"accepted_count"`
	RejectedCount   int     `json:"rejected_count"`
	RevokedCount    int     `json:"revoked_count"`
	AcceptanceRate  float64 `json:"acceptance_rate"`
	RejectionRate   float64 `json:"rejection_rate"`
	RevocationRate  float64 `json:"revocation_rate"`
}