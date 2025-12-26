package jobs

import (
	"context"
	"log"
	"time"
)

type MedicationReminderJob struct {
	// Add dependencies as needed
}

func NewMedicationReminderJob() *MedicationReminderJob {
	return &MedicationReminderJob{}
}

func (j *MedicationReminderJob) Name() string {
	return "medication-reminder"
}

func (j *MedicationReminderJob) Interval() time.Duration {
	return 15 * time.Minute
}

func (j *MedicationReminderJob) Run(ctx context.Context) error {
	log.Println("Running medication reminder job")

	// TODO: implement
	// 1. Get all active medications
	// 2. Check which ones are due
	// 3. Send notifications to family members

	return nil
}
