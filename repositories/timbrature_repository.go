package repositories

import (
	"database/sql"
	"merendels-backend/config"
	"merendels-backend/models"
	"time"
)

type TimbratureRepository struct {}

// NewTimbratureRepository crea la nuova istanza della repo
func NewTimbratureRepository() *TimbratureRepository {
	return &TimbratureRepository{}
}

// Create inserisce una nuova timbratura nel database
func (r *TimbratureRepository) Create(timbratura *models.Timbrature) error {
	query := `INSERT INTO timbrature (user_id, timestamp, action_type, location, geolocation) VALUES ($1, $2, $3, $4, $5) RETURNING id`
	
	err := config.DB.QueryRow(query, timbratura.UserID, timbratura.Timestamp, timbratura.ActionType, timbratura.Location, timbratura.Geolocation).Scan(&timbratura.ID)
	if err != nil {
		return err
	}

	return nil
}

// GetAll recupera tutte le timbrature con eventuali filtri di limite e offset
func (r *TimbratureRepository) GetAll(limit, offset int) ([]models.Timbrature, error) {
	// Query con ordinamento per data decrescente e paginazione tramite LIMIT e OFFSET
	query := `SELECT id, user_id, timestamp, action_type, location, geolocation 
			  FROM timbrature 
			  ORDER BY timestamp DESC 
			  LIMIT $1 OFFSET $2`

	// Esegue la query, passando i parametri limit e offset
	rows, err := config.DB.Query(query, limit, offset)
	if err != nil {
		// Se c’è un errore nell’esecuzione della query lo ritorna
		return nil, err
	}
	// Chiude le righe una volta usciti dalla funzione
	defer rows.Close()
	
	// Slice che conterrà tutte le timbrature recuperate
	var timbrature []models.Timbrature

	// Itera su ogni riga del risultato
	for rows.Next() {
		var t models.Timbrature

		// Popola la struct Timbrature con i valori delle colonne della riga corrente
		err := rows.Scan(&t.ID, &t.UserID, &t.Timestamp, &t.ActionType, &t.Location, &t.Geolocation)
		if err != nil {
			// Se c’è un errore nello scan ritorna l’errore
			return nil, err
		}

		// Aggiunge la timbratura alla slice
		timbrature = append(timbrature, t)
	}

	// Controlla se ci sono errori nell’iterazione delle righe
	if err = rows.Err(); err != nil {
		return nil, err
	}

	// Ritorna la lista di timbrature e nessun errore
	return timbrature, nil
}

// GetByUserID recupera timbrature di un utente specifico
func (r *TimbratureRepository) GetByUserID(userID, limit, offset int) ([]models.Timbrature, error) {
	// Query con filtro per userID, ordinamento per data decrescente e paginazione tramite LIMIT e OFFSET
	query := `
		SELECT id, user_id, timestamp, action_type, location, geolocation 
		FROM timbrature 
		WHERE user_id = $1 
		ORDER BY timestamp DESC 
		LIMIT $2 OFFSET $3`

	// Esegue la query, passando i parametri userID, limit e offset
	rows, err := config.DB.Query(query, userID, limit, offset)
	if err != nil {
		return nil, err
	}
	// Chiude le righe una volta usciti dalla funzione
	defer rows.Close()

		// Slice che conterrà tutte le timbrature recuperate
	var timbrature []models.Timbrature

	// Itera su ogni riga del risultato
	for rows.Next() {
		var t models.Timbrature

		// Popola la struct Timbrature con i valori delle colonne della riga corrente
		err := rows.Scan(&t.ID, &t.UserID, &t.Timestamp, &t.ActionType, &t.Location, &t.Geolocation)
		if err != nil {
			// Se c’è un errore nello scan ritorna l’errore
			return nil, err
		}

		// Aggiunge la timbratura alla slice
		timbrature = append(timbrature, t)
	}

	// Controlla se ci sono errori nell’iterazione delle righe
	if err = rows.Err(); err != nil {
		return nil, err
	}

	// Ritorna la lista di timbrature e nessun errore
	return timbrature, nil
}

// GetByUserIDAndDate recupera tutte le timbrature di un utente in una data specifica
func (r *TimbratureRepository) GetByUserIDAndDate(userID int, date time.Time) ([]models.Timbrature, error) {
	// Query che filtra per user_id e per data (solo la parte "giorno" del timestamp)
	query := `
		SELECT id, user_id, timestamp, action_type, location, geolocation 
		FROM timbrature 
		WHERE user_id = $1 
		AND DATE(timestamp) = DATE($2)
		ORDER BY timestamp ASC`
	
	// Esegue la query con i parametri userID e date
	rows, err := config.DB.Query(query, userID, date)
	if err != nil {
		return nil, err
	}
	// Chiude le righe al termine della funzione
	defer rows.Close()
	
	// Slice per contenere le timbrature trovate
	var timbrature []models.Timbrature
	
	// Itera sulle righe del risultato
	for rows.Next() {
		var t models.Timbrature
		
		// Legge i valori della riga e li mappa nella struct
		err := rows.Scan(&t.ID, &t.UserID, &t.Timestamp, &t.ActionType, &t.Location, &t.Geolocation)
		if err != nil {
			return nil, err
		}
		
		// Aggiunge la timbratura alla lista
		timbrature = append(timbrature, t)
	}
	
	// Controlla eventuali errori emersi durante l’iterazione
	if err = rows.Err(); err != nil {
		return nil, err
	}
	
	// Ritorna l’elenco di timbrature
	return timbrature, nil
}

// GetLastTimbratureByUserID recupera l'ultima timbratura registrata da un utente
func (r *TimbratureRepository) GetLastTimbratureByUserID(userID int) (*models.Timbrature, error) {
	// Query che prende l'ultima timbratura per user_id ordinando in ordine decrescente e limitando a 1
	query := `
		SELECT id, user_id, timestamp, action_type, location, geolocation 
		FROM timbrature 
		WHERE user_id = $1 
		ORDER BY timestamp DESC 
		LIMIT 1`
	
	// Struct che conterrà il risultato
	var t models.Timbrature
	
	// Usa QueryRow perché ci aspettiamo un solo record
	err := config.DB.QueryRow(query, userID).Scan(
		&t.ID, &t.UserID, &t.Timestamp, &t.ActionType, &t.Location, &t.Geolocation)
	
	if err != nil {
		// Caso in cui non ci sono righe (utente senza timbrature)
		if err == sql.ErrNoRows {
			return nil, nil
		}
		// Qualsiasi altro errore viene ritornato
		return nil, err
	}
	
	// Ritorna la timbratura trovata
	return &t, nil
}

// Delete elimina una timbratura dal database (usata solo per correzioni)
func (r *TimbratureRepository) Delete(id int) error {
	// Query SQL per cancellare una timbratura dato l'id
	query := `DELETE FROM timbrature WHERE id = $1`
	
	// Esegue la query di DELETE
	result, err := config.DB.Exec(query, id)
	if err != nil {
		return err
	}
	
	// Controlla quante righe sono state effettivamente eliminate
	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return err
	}
	
	// Se nessuna riga è stata cancellata, ritorna errore "nessuna riga trovata"
	if rowsAffected == 0 {
		return sql.ErrNoRows
	}
	
	// Nessun errore → operazione riuscita
	return nil
}

// CountByUserID conta il numero totale di timbrature associate a un utente
func (r *TimbratureRepository) CountByUserID(userID int) (int, error) {
	// Query SQL che conta tutte le righe per un determinato user_id
	query := `SELECT COUNT(*) FROM timbrature WHERE user_id = $1`
	
	// Variabile che conterrà il risultato
	var count int
	
	// QueryRow perché ci aspettiamo un solo valore (il conteggio)
	err := config.DB.QueryRow(query, userID).Scan(&count)
	if err != nil {
		return 0, err
	}
	
	// Ritorna il totale delle timbrature
	return count, nil
}
