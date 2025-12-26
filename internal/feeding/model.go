package feeding

import "time"

type FeedingType string

const (
	FeedingTypeBreast  FeedingType = "breast"
	FeedingTypeBottle  FeedingType = "bottle"
	FeedingTypeFormula FeedingType = "formula"
	FeedingTypeSolid   FeedingType = "solid"
)

type Feeding struct {
	ID        string      `json:"id"`
	ChildID   string      `json:"child_id"`
	Type      FeedingType `json:"type"`
	StartTime time.Time   `json:"start_time"`
	EndTime   *time.Time  `json:"end_time,omitempty"`
	Amount    *float64    `json:"amount,omitempty"` // in ml or oz
	Unit      string      `json:"unit,omitempty"`   // ml, oz
	Side      string      `json:"side,omitempty"`   // left, right, both (for breast)
	Notes     string      `json:"notes,omitempty"`
	CreatedAt time.Time   `json:"created_at"`
	UpdatedAt time.Time   `json:"updated_at"`
	SyncedAt  *time.Time  `json:"synced_at,omitempty"`
}

type CreateFeedingRequest struct {
	ChildID   string      `json:"child_id" binding:"required"`
	Type      FeedingType `json:"type" binding:"required"`
	StartTime time.Time   `json:"start_time" binding:"required"`
	EndTime   *time.Time  `json:"end_time,omitempty"`
	Amount    *float64    `json:"amount,omitempty"`
	Unit      string      `json:"unit,omitempty"`
	Side      string      `json:"side,omitempty"`
	Notes     string      `json:"notes,omitempty"`
}

type FeedingFilter struct {
	ChildID   string
	StartDate *time.Time
	EndDate   *time.Time
	Type      *FeedingType
}
