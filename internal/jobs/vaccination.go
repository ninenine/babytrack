package jobs

import (
	"context"
	"fmt"
	"log"
	"time"

	"family-tracker/internal/notifications"
	"family-tracker/internal/vaccination"

	"github.com/google/uuid"
)

// VaccinationReminderJob checks for upcoming vaccinations and sends notifications.
type VaccinationReminderJob struct {
	vaccinationService vaccination.Service
	notificationHub    *notifications.Hub
}

func NewVaccinationReminderJob(vaccinationService vaccination.Service, hub *notifications.Hub) *VaccinationReminderJob {
	return &VaccinationReminderJob{
		vaccinationService: vaccinationService,
		notificationHub:    hub,
	}
}

func (j *VaccinationReminderJob) Name() string {
	return "vaccination-reminder"
}

func (j *VaccinationReminderJob) Interval() time.Duration {
	return 6 * time.Hour // Check 4 times a day
}

func (j *VaccinationReminderJob) Run(ctx context.Context) error {
	log.Println("[VaccinationReminderJob] Checking for upcoming vaccinations...")

	// Get all vaccinations due in the next 7 days (across all children)
	upcoming, err := j.vaccinationService.GetUpcoming(ctx, "", 7)
	if err != nil {
		return err
	}

	now := time.Now()
	notifiedCount := 0

	for _, vax := range upcoming {
		if vax.Completed {
			continue
		}

		daysUntil := int(vax.ScheduledAt.Sub(now).Hours() / 24)

		// Only notify for vaccinations due today or in next 3 days
		if daysUntil > 3 {
			continue
		}

		var message string
		if daysUntil <= 0 {
			message = fmt.Sprintf("%s (Dose %d) is due today", vax.Name, vax.Dose)
		} else if daysUntil == 1 {
			message = fmt.Sprintf("%s (Dose %d) is due tomorrow", vax.Name, vax.Dose)
		} else {
			message = fmt.Sprintf("%s (Dose %d) is due in %d days", vax.Name, vax.Dose, daysUntil)
		}

		log.Printf("[VaccinationReminderJob] %s (Child: %s)", message, vax.ChildID)
		notifiedCount++

		// Broadcast notification to connected clients
		if j.notificationHub != nil && j.notificationHub.ClientCount() > 0 {
			j.notificationHub.Broadcast(notifications.Event{
				ID:        uuid.New().String(),
				Type:      notifications.EventVaccinationDue,
				Title:     "Vaccination Reminder",
				Message:   message,
				ChildID:   vax.ChildID,
				Timestamp: now,
			})
		}
	}

	log.Printf("[VaccinationReminderJob] Check complete. %d vaccination reminders sent", notifiedCount)
	return nil
}
