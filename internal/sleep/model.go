package sleep

import "time"

type SleepType string

const (
	SleepTypeNap   SleepType = "nap"
	SleepTypeNight SleepType = "night"
)

type Sleep struct {
	ID        string     `json:"id"`
	ChildID   string     `json:"child_id"`
	Type      SleepType  `json:"type"`
	StartTime time.Time  `json:"start_time"`
	EndTime   *time.Time `json:"end_time,omitempty"`
	Quality   *int       `json:"quality,omitempty"` // 1-5 rating
	Notes     string     `json:"notes,omitempty"`
	CreatedAt time.Time  `json:"created_at"`
	UpdatedAt time.Time  `json:"updated_at"`
	SyncedAt  *time.Time `json:"synced_at,omitempty"`
}

type CreateSleepRequest struct {
	ChildID   string     `json:"child_id" binding:"required"`
	Type      SleepType  `json:"type" binding:"required"`
	StartTime time.Time  `json:"start_time" binding:"required"`
	EndTime   *time.Time `json:"end_time,omitempty"`
	Quality   *int       `json:"quality,omitempty"`
	Notes     string     `json:"notes,omitempty"`
}

type SleepFilter struct {
	ChildID   string
	StartDate *time.Time
	EndDate   *time.Time
	Type      *SleepType
}

type SleepStats struct {
	TotalSleep   time.Duration `json:"total_sleep"`
	AverageNap   time.Duration `json:"average_nap"`
	AverageNight time.Duration `json:"average_night"`
	NapCount     int           `json:"nap_count"`
}
