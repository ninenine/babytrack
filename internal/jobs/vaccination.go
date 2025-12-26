package jobs

import (
	"context"
	"log"
	"time"
)

type VaccinationReminderJob struct {
	// Add dependencies as needed
}

func NewVaccinationReminderJob() *VaccinationReminderJob {
	return &VaccinationReminderJob{}
}

func (j *VaccinationReminderJob) Name() string {
	return "vaccination-reminder"
}

func (j *VaccinationReminderJob) Interval() time.Duration {
	return 24 * time.Hour
}

func (j *VaccinationReminderJob) Run(ctx context.Context) error {
	log.Println("Running vaccination reminder job")

	// TODO: implement
	// 1. Get all upcoming vaccinations in the next 7 days
	// 2. Send reminders to family members
	// 3. Mark reminders as sent

	return nil
}
