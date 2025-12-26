package medication

import "time"

type Medication struct {
	ID           string     `json:"id"`
	ChildID      string     `json:"child_id"`
	Name         string     `json:"name"`
	Dosage       string     `json:"dosage"`
	Unit         string     `json:"unit"`
	Frequency    string     `json:"frequency"` // daily, twice_daily, as_needed, etc.
	Instructions string     `json:"instructions,omitempty"`
	StartDate    time.Time  `json:"start_date"`
	EndDate      *time.Time `json:"end_date,omitempty"`
	Active       bool       `json:"active"`
	CreatedAt    time.Time  `json:"created_at"`
	UpdatedAt    time.Time  `json:"updated_at"`
}

type MedicationLog struct {
	ID           string    `json:"id"`
	MedicationID string    `json:"medication_id"`
	ChildID      string    `json:"child_id"`
	GivenAt      time.Time `json:"given_at"`
	GivenBy      string    `json:"given_by"` // user ID
	Dosage       string    `json:"dosage"`
	Notes        string    `json:"notes,omitempty"`
	CreatedAt    time.Time `json:"created_at"`
	SyncedAt     *time.Time `json:"synced_at,omitempty"`
}

type CreateMedicationRequest struct {
	ChildID      string     `json:"child_id" binding:"required"`
	Name         string     `json:"name" binding:"required"`
	Dosage       string     `json:"dosage" binding:"required"`
	Unit         string     `json:"unit" binding:"required"`
	Frequency    string     `json:"frequency" binding:"required"`
	Instructions string     `json:"instructions,omitempty"`
	StartDate    time.Time  `json:"start_date" binding:"required"`
	EndDate      *time.Time `json:"end_date,omitempty"`
}

type LogMedicationRequest struct {
	MedicationID string    `json:"medication_id" binding:"required"`
	GivenAt      time.Time `json:"given_at" binding:"required"`
	Dosage       string    `json:"dosage" binding:"required"`
	Notes        string    `json:"notes,omitempty"`
}

type MedicationFilter struct {
	ChildID    string
	ActiveOnly bool
}
