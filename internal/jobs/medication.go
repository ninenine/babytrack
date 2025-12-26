package jobs

import (
	"context"
	"log"
	"time"

	"family-tracker/internal/medication"
)

// MedicationReminderJob checks for medications that are due and logs reminders.
// In a production system, this would integrate with a notification service.
type MedicationReminderJob struct {
	medicationService medication.Service
}

func NewMedicationReminderJob(medicationService medication.Service) *MedicationReminderJob {
	return &MedicationReminderJob{
		medicationService: medicationService,
	}
}

func (j *MedicationReminderJob) Name() string {
	return "medication-reminder"
}

func (j *MedicationReminderJob) Interval() time.Duration {
	return 15 * time.Minute
}

func (j *MedicationReminderJob) Run(ctx context.Context) error {
	log.Println("[MedicationReminderJob] Checking for due medications...")

	// Get all active medications
	meds, err := j.medicationService.List(ctx, &medication.MedicationFilter{ActiveOnly: true})
	if err != nil {
		return err
	}

	now := time.Now()
	dueCount := 0

	for _, med := range meds {
		// Get the last log for this medication
		lastLog, err := j.medicationService.GetLastLog(ctx, med.ID)
		if err != nil {
			log.Printf("[MedicationReminderJob] Error getting last log for %s: %v", med.Name, err)
			continue
		}

		// Calculate if medication is due
		isDue := j.isMedicationDue(med, lastLog, now)
		if isDue {
			dueCount++
			log.Printf("[MedicationReminderJob] Medication due: %s (Child: %s, Frequency: %s)",
				med.Name, med.ChildID, med.Frequency)
			// TODO: In production, send notification via push/email service
		}
	}

	log.Printf("[MedicationReminderJob] Check complete. %d medications due out of %d active", dueCount, len(meds))
	return nil
}

// isMedicationDue determines if a medication is due based on its frequency and last administration
func (j *MedicationReminderJob) isMedicationDue(med medication.Medication, lastLog *medication.MedicationLog, now time.Time) bool {
	// If never given, it's due (unless it's as_needed)
	if lastLog == nil {
		return med.Frequency != "as_needed"
	}

	// Calculate the expected interval based on frequency
	var expectedInterval time.Duration
	switch med.Frequency {
	case "once_daily":
		expectedInterval = 24 * time.Hour
	case "twice_daily":
		expectedInterval = 12 * time.Hour
	case "three_times_daily":
		expectedInterval = 8 * time.Hour
	case "four_times_daily":
		expectedInterval = 6 * time.Hour
	case "every_4_hours":
		expectedInterval = 4 * time.Hour
	case "every_6_hours":
		expectedInterval = 6 * time.Hour
	case "every_8_hours":
		expectedInterval = 8 * time.Hour
	case "as_needed":
		return false // Never automatically due
	default:
		expectedInterval = 24 * time.Hour // Default to daily
	}

	// Add a 30-minute grace period before considering it due
	timeSinceLastDose := now.Sub(lastLog.GivenAt)
	return timeSinceLastDose >= expectedInterval-30*time.Minute
}
