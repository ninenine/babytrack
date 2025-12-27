package jobs

import (
	"context"
	"fmt"
	"log"
	"time"

	"family-tracker/internal/appointment"
	"family-tracker/internal/notifications"

	"github.com/google/uuid"
)

// AppointmentReminderJob checks for upcoming appointments and sends notifications.
type AppointmentReminderJob struct {
	appointmentService appointment.Service
	notificationHub    *notifications.Hub
}

func NewAppointmentReminderJob(appointmentService appointment.Service, hub *notifications.Hub) *AppointmentReminderJob {
	return &AppointmentReminderJob{
		appointmentService: appointmentService,
		notificationHub:    hub,
	}
}

func (j *AppointmentReminderJob) Name() string {
	return "appointment-reminder"
}

func (j *AppointmentReminderJob) Interval() time.Duration {
	return 30 * time.Minute // Check every 30 minutes
}

func (j *AppointmentReminderJob) Run(ctx context.Context) error {
	log.Println("[AppointmentReminderJob] Checking for upcoming appointments...")

	// Get all appointments in the next 2 days (across all children)
	upcoming, err := j.appointmentService.GetUpcoming(ctx, "", 2)
	if err != nil {
		return err
	}

	now := time.Now()
	notifiedCount := 0

	for _, apt := range upcoming {
		if apt.Completed || apt.Cancelled {
			continue
		}

		timeUntil := apt.ScheduledAt.Sub(now)
		hoursUntil := timeUntil.Hours()

		// Notify for appointments:
		// - Starting in the next hour
		// - Starting tomorrow (within 24 hours)
		var message string
		var shouldNotify bool

		if hoursUntil <= 1 && hoursUntil > 0 {
			// Starting soon (within 1 hour)
			minutes := int(timeUntil.Minutes())
			message = fmt.Sprintf("%s starts in %d minutes", apt.Title, minutes)
			shouldNotify = true
		} else if hoursUntil <= 24 && hoursUntil > 1 {
			// Today or tomorrow
			if apt.ScheduledAt.Day() == now.Day() {
				message = fmt.Sprintf("%s is today at %s", apt.Title, apt.ScheduledAt.Format("3:04 PM"))
			} else {
				message = fmt.Sprintf("%s is tomorrow at %s", apt.Title, apt.ScheduledAt.Format("3:04 PM"))
			}
			shouldNotify = true
		}

		if !shouldNotify {
			continue
		}

		log.Printf("[AppointmentReminderJob] %s (Child: %s)", message, apt.ChildID)
		notifiedCount++

		// Broadcast notification to connected clients
		if j.notificationHub != nil && j.notificationHub.ClientCount() > 0 {
			j.notificationHub.Broadcast(notifications.Event{
				ID:        uuid.New().String(),
				Type:      notifications.EventAppointmentSoon,
				Title:     "Appointment Reminder",
				Message:   message,
				ChildID:   apt.ChildID,
				Timestamp: now,
			})
		}
	}

	log.Printf("[AppointmentReminderJob] Check complete. %d appointment reminders sent", notifiedCount)
	return nil
}
