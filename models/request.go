package models

import "time"

type RequestType string
const (
	RequestHolidays RequestType = "FERIE"
	RequestPermits RequestType = "PERMESSO" 
)

type Request struct {
	ID        int `json:"id"`
	UserID    int `json:"user_id"`
	StartDate time.Time  `json:"start_date"`
	EndDate time.Time  `json:"end_date"`
	RequestType RequestType  `json:"request_type"`
	Notes *string  `json:"notes"`
	CreatedAt time.Time  `json:"created_at"`
}

// Request front-end -> back-end
// NO ID & ModifiedAt => generated from the database by defaulta
type CreateRequest struct {
	UserID int  `json:"user_id"`
	StartDate time.Time  `json:"start_date" binding:"required"`
	EndDate time.Time  `json:"end_date" binding:"required"`
	RequestType RequestType  `json:"request_type" binding:"required"`
	Notes *string  `json:"notes"`
}