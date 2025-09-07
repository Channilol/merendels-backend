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

// RequestWithStatus per response API con status e approval_id
type RequestWithStatus struct {
	ID          int         `json:"id"`
	UserID      int         `json:"user_id"`
	StartDate   time.Time   `json:"start_date"`
	EndDate     time.Time   `json:"end_date"`
	RequestType RequestType `json:"request_type"`
	Notes       *string     `json:"notes"`
	CreatedAt   time.Time   `json:"created_at"`
	Status      string      `json:"status"`       // PENDING, APPROVED, REJECTED, REVOKED
	ApprovalID  *int        `json:"approval_id"`  // null se PENDING
	ApproverName *string     `json:"approver_name"`
}