package services

import (
	"errors"
	"fmt"
	"log"
	"merendels-backend/models"
	"merendels-backend/repositories"
	"time"
)

type TimbratureService struct {
	repository *repositories.TimbratureRepository
}

// NewTimbratureService crea la nuova istanza della repo
func NewTimbratureService() *TimbratureService {
	return &TimbratureService{
		repository: repositories.NewTimbratureRepository(),
	}
}

// CreateTimbrature crea una nuova timbratura con validazioni business
func (s *TimbratureService) CreateTimbrature(userID int, request *models.CreateTimbratureRequest) (*models.TimbratureResponse, error) {
	// Validazioni base
	if request.ActionType != models.ActionEnter && request.ActionType != models.ActionExit {
		// Azione non valida
		return nil, errors.New("invalid action type")
	}

	if request.Location != models.LocationOffice && request.Location != models.LocationSmart {
		// Location non valida
		return nil, errors.New("invalid location")
	}

	// Verifica sequenza ENTRATA -> USCITA
	lastTimbrature, err := s.repository.GetLastTimbratureByUserID(userID)
	if err != nil {
		return nil, fmt.Errorf("error checking last timbrature: %w", err)
	}

	// Validazione sequenza logica
	if lastTimbrature != nil {
		if lastTimbrature.ActionType == request.ActionType {
			if request.ActionType == models.ActionEnter {
				return nil, errors.New("cannot enter twice in a row - you must exit first")
			} else {
				return nil, errors.New("cannot exit twice in a row - you must enter first")
			}
		}
	} else {
		// Prima timbratura ever - deve essere ENTRATA
		if request.ActionType == models.ActionExit {
			return nil, errors.New("first timbratura must be ENTRATA")
		}
	}

	// Timestamp generato dal server (anti-frode)
	now := time.Now()

	// CreateRequest -> Timbrature model
	timbrature := &models.Timbrature{
		UserID: userID,
		Timestamp: now,
		ActionType: request.ActionType,
		Location: request.Location,
		Geolocation: request.Geolocation,
	}

	// Salva nel database
	err = s.repository.Create(timbrature)
	if err != nil {
		return nil, fmt.Errorf("error creating timbrature: %w", err)
	}

	// Timbrature → Response
	response := &models.TimbratureResponse{
		ID:          timbrature.ID,
		UserID:      timbrature.UserID,
		Timestamp:   timbrature.Timestamp,
		ActionType:  timbrature.ActionType,
		Location:    timbrature.Location,
		Geolocation: timbrature.Geolocation,
	}

	// Log per audit
	log.Printf("User %d created %s timbratura at %s", 
		userID, request.ActionType, now.Format("2006-01-02 15:04:05"))

	return response, nil
}

// GetUserTimbrature recupera le timbrature dell'utente autenticato
func (s * TimbratureService) GetUserTimbrature(userID, limit, offset int) ([]models.TimbratureResponse, error) {
	// Validazioni base
	if limit <= 0 || limit >= 100 {
		limit = 20 // Default
	}
	if offset < 0 {
		offset = 0
	}

	// Recupera dalla repository
	timbrature, err := s.repository.GetByUserID(userID, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("error fetching timbrature: %w", err)
	}

	// []Timbrature -> []TimbratureResponse
	var responses []models.TimbratureResponse

	for _, t := range timbrature {
		response := models.TimbratureResponse(t)
		responses = append(responses, response)
	}
	return responses, nil
}

// GetUserTimbratureByDate recupera timbrature utente per data specifica
func (s *TimbratureService) GetUserTimbratureByDate(userID int, date time.Time) ([]models.TimbratureResponse, error) {
	timbrature, err := s.repository.GetByUserIDAndDate(userID, date)
	if err != nil {
		return nil, fmt.Errorf("error fetching timbrature by date: %w", err)
	}

	var responses []models.TimbratureResponse
	for _, t := range timbrature {
		response := models.TimbratureResponse(t)
		responses = append(responses, response)
	}

	return responses, nil
}

// GetTodayTimbrature recupera le timbrature di oggi dell'utente
func (s *TimbratureService) GetTodayTimbrature(userID int) ([]models.TimbratureResponse, error) {
	today := time.Now()
	timbrature, err := s.repository.GetByUserIDAndDate(userID, today)
	if err != nil {
		return nil, fmt.Errorf("error fetching today timbrature: %w", err)
	}

	var responses []models.TimbratureResponse
	for _, t := range timbrature {
		response := models.TimbratureResponse(t)
		responses = append(responses, response)
	}

	return responses, nil
}

// GetLastTimbratura recupera l'ultima timbratura dell'utente
func (s *TimbratureService) GetLastTimbrature(userID int) (*models.TimbratureResponse, error) {
	timbratura, err := s.repository.GetLastTimbratureByUserID(userID)
	if err != nil {
		return nil, fmt.Errorf("error fetching last timbratura: %w", err)
	}

	if timbratura == nil {
		return nil, nil // Nessuna timbratura precedente
	}

	response := &models.TimbratureResponse{
		ID:          timbratura.ID,
		UserID:      timbratura.UserID,
		Timestamp:   timbratura.Timestamp,
		ActionType:  timbratura.ActionType,
		Location:    timbratura.Location,
		Geolocation: timbratura.Geolocation,
	}

	return response, nil
}

// GetAllTimbrature recupera tutte le timbrature (solo per admin)
func (s *TimbratureService) GetAllTimbrature(limit, offset int) ([]models.TimbratureResponse, error) {
	//  Validazioni paginazione
	if limit <= 0 || limit > 100 {
		limit = 20
	}
	if offset < 0 {
		offset = 0
	}

	timbrature, err := s.repository.GetAll(limit, offset)
	if err != nil {
		return nil, fmt.Errorf("error fetching all timbrature: %w", err)
	}

	var responses []models.TimbratureResponse
	for _, t := range timbrature {
		response := models.TimbratureResponse(t)
		responses = append(responses, response)
	}

	return responses, nil
}

// GetWorkingStatus determina lo stato lavorativo corrente dell'utente
func (s *TimbratureService) GetWorkingStatus(userID int) (*WorkingStatusResponse, error) {
	lastTimbratura, err := s.repository.GetLastTimbratureByUserID(userID)
	if err != nil {
		return nil, fmt.Errorf("error fetching last timbratura: %w", err)
	}

	status := &WorkingStatusResponse{
		UserID:    userID,
		IsWorking: false,
		LastTimbratura: nil,
	}

	if lastTimbratura != nil {
		status.IsWorking = (lastTimbratura.ActionType == models.ActionEnter)
		status.LastTimbratura = &models.TimbratureResponse{
			ID:          lastTimbratura.ID,
			UserID:      lastTimbratura.UserID,
			Timestamp:   lastTimbratura.Timestamp,
			ActionType:  lastTimbratura.ActionType,
			Location:    lastTimbratura.Location,
			Geolocation: lastTimbratura.Geolocation,
		}
	}

	return status, nil
}

// WorkingStatusResponse rappresenta lo stato lavorativo dell'utente
type WorkingStatusResponse struct {
	UserID         int                       `json:"user_id"`
	IsWorking      bool                      `json:"is_working"`
	LastTimbratura *models.TimbratureResponse `json:"last_timbratura"`
}

// DeleteTimbratura elimina una timbratura (solo per correzioni admin)
func (s *TimbratureService) DeleteTimbratura(id int) error {
	//  Solo admin dovrebbero poter eliminare timbrature
	// La validazione del ruolo admin verrà fatta nell'handler

	err := s.repository.Delete(id)
	if err != nil {
		return fmt.Errorf("error deleting timbratura: %w", err)
	}

	log.Printf("Timbratura %d deleted", id)
	return nil
}