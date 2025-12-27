package vaccination

import "time"

type Vaccination struct {
	ID           string     `json:"id"`
	ChildID      string     `json:"child_id"`
	Name         string     `json:"name"`
	Dose         int        `json:"dose"` // 1st, 2nd, 3rd, etc.
	ScheduledAt  time.Time  `json:"scheduled_at"`
	AdministeredAt *time.Time `json:"administered_at,omitempty"`
	Provider     string     `json:"provider,omitempty"`
	Location     string     `json:"location,omitempty"`
	LotNumber    string     `json:"lot_number,omitempty"`
	Notes        string     `json:"notes,omitempty"`
	Completed    bool       `json:"completed"`
	CreatedAt    time.Time  `json:"created_at"`
	UpdatedAt    time.Time  `json:"updated_at"`
}

type VaccinationSchedule struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	Description string `json:"description"`
	AgeWeeks    int    `json:"age_weeks"`  // recommended age in weeks (0 = birth)
	AgeMonths   int    `json:"age_months"` // recommended age in months (for display)
	AgeLabel    string `json:"age_label"`  // human-readable age (e.g., "Birth", "6 weeks")
	Dose        int    `json:"dose"`
}

type CreateVaccinationRequest struct {
	ChildID     string    `json:"child_id" binding:"required"`
	Name        string    `json:"name" binding:"required"`
	Dose        int       `json:"dose" binding:"required"`
	ScheduledAt time.Time `json:"scheduled_at" binding:"required"`
}

type RecordVaccinationRequest struct {
	AdministeredAt time.Time `json:"administered_at" binding:"required"`
	Provider       string    `json:"provider,omitempty"`
	Location       string    `json:"location,omitempty"`
	LotNumber      string    `json:"lot_number,omitempty"`
	Notes          string    `json:"notes,omitempty"`
}

type VaccinationFilter struct {
	ChildID      string
	Completed    *bool
	UpcomingOnly bool
}
