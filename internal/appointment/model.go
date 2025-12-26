package appointment

import "time"

type AppointmentType string

const (
	AppointmentTypeWellVisit  AppointmentType = "well_visit"
	AppointmentTypeSickVisit  AppointmentType = "sick_visit"
	AppointmentTypeSpecialist AppointmentType = "specialist"
	AppointmentTypeDental     AppointmentType = "dental"
	AppointmentTypeOther      AppointmentType = "other"
)

type Appointment struct {
	ID          string          `json:"id"`
	ChildID     string          `json:"child_id"`
	Type        AppointmentType `json:"type"`
	Title       string          `json:"title"`
	Provider    string          `json:"provider,omitempty"`
	Location    string          `json:"location,omitempty"`
	ScheduledAt time.Time       `json:"scheduled_at"`
	Duration    int             `json:"duration"` // in minutes
	Notes       string          `json:"notes,omitempty"`
	Completed   bool            `json:"completed"`
	Cancelled   bool            `json:"cancelled"`
	CreatedAt   time.Time       `json:"created_at"`
	UpdatedAt   time.Time       `json:"updated_at"`
}

type CreateAppointmentRequest struct {
	ChildID     string          `json:"child_id" binding:"required"`
	Type        AppointmentType `json:"type" binding:"required"`
	Title       string          `json:"title" binding:"required"`
	Provider    string          `json:"provider,omitempty"`
	Location    string          `json:"location,omitempty"`
	ScheduledAt time.Time       `json:"scheduled_at" binding:"required"`
	Duration    int             `json:"duration"`
	Notes       string          `json:"notes,omitempty"`
}

type AppointmentFilter struct {
	ChildID      string
	Type         *AppointmentType
	UpcomingOnly bool
	StartDate    *time.Time
	EndDate      *time.Time
}
