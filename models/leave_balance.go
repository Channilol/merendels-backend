package models

import "time"

type LeaveBalance struct {
	ID                  int     `json:"id"`
	UserID              int     `json:"user_id"`
	AccumulatedHolidays float32 `json:"accumulated_holidays"`
	AccumulatedPermits  float32 `json:"accumulated_permits"`
	ModifiedAt          time.Time  `json:"modified_at"`
}

// Request front-end -> back-end
// NO ID & ModifiedAt => generated from the database by default
type CreateLeaveBalanceRequest struct {
	UserID              int     `json:"user_id" binding:"required"`
	AccumulatedHolidays float32 `json:"accumulated_holidays" binding:"required"`
	AccumulatedPermits  float32 `json:"accumulated_permits" binding:"required"`
}

// Response back-end -> front-end
type LeaveBalanceResponse struct {
	ID                  int     `json:"id"`
	UserID              int     `json:"user_id"`
	AccumulatedHolidays float32 `json:"accumulated_holidays"`
	AccumulatedPermits  float32 `json:"accumulated_permits"`
	ModifiedAt          time.Time  `json:"modified_at"`
}