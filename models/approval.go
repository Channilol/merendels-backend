package models

import "time"

type ApprovalStatus string

const (
	ApprovalAccepted ApprovalStatus = "APPROVED"
	ApprovalRejected ApprovalStatus = "REJECTED"
	ApprovalRevoked  ApprovalStatus = "REVOKED"
)

type Approval struct {
	ID         int            `json:"id"`
	RequestID  int            `json:"request_id"`
	ApproverID int            `json:"approver_id"`
	Status     ApprovalStatus `json:"status"`
	Comments   *string         `json:"comments"`
	ApprovedAt time.Time 	`json:"approved_at"`
}

// Request front-end -> back-end
// NO ID & ApprovedAt => generated from the database by default
type CreateApprovalRequest struct {
	RequestID  int            `json:"request_id" binding:"required"`
	ApproverID int            `json:"approver_id" binding:"required"`
	Status     ApprovalStatus `json:"status" binding:"required"`
	Comments   *string         `json:"comments"`
}