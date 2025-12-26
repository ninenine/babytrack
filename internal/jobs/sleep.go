package jobs

import (
	"context"
	"log"
	"time"

	"family-tracker/internal/sleep"
)

// SleepAnalyticsJob calculates daily sleep statistics.
// In a production system, this could store aggregated data for faster queries.
type SleepAnalyticsJob struct {
	sleepService sleep.Service
}

func NewSleepAnalyticsJob(sleepService sleep.Service) *SleepAnalyticsJob {
	return &SleepAnalyticsJob{
		sleepService: sleepService,
	}
}

func (j *SleepAnalyticsJob) Name() string {
	return "sleep-analytics"
}

func (j *SleepAnalyticsJob) Interval() time.Duration {
	return 6 * time.Hour
}

func (j *SleepAnalyticsJob) Run(ctx context.Context) error {
	log.Println("[SleepAnalyticsJob] Running sleep analytics...")

	// Get yesterday's date range for analysis
	now := time.Now()
	yesterday := now.AddDate(0, 0, -1)
	startOfYesterday := time.Date(yesterday.Year(), yesterday.Month(), yesterday.Day(), 0, 0, 0, 0, yesterday.Location())
	endOfYesterday := startOfYesterday.AddDate(0, 0, 1)

	// Get all sleep sessions that started yesterday
	sessions, err := j.sleepService.List(ctx, &sleep.SleepFilter{
		StartDate: &startOfYesterday,
		EndDate:   &endOfYesterday,
	})
	if err != nil {
		return err
	}

	// Group by child and calculate totals
	childSleep := make(map[string]time.Duration)
	childNaps := make(map[string]int)
	childNightSleep := make(map[string]time.Duration)

	for _, session := range sessions {
		if session.EndTime == nil {
			continue // Skip incomplete sessions
		}

		duration := session.EndTime.Sub(session.StartTime)
		childSleep[session.ChildID] += duration

		if session.Type == "nap" {
			childNaps[session.ChildID]++
		} else {
			childNightSleep[session.ChildID] += duration
		}
	}

	// Log results
	for childID, totalSleep := range childSleep {
		naps := childNaps[childID]
		nightSleep := childNightSleep[childID]
		log.Printf("[SleepAnalyticsJob] Child %s: Total=%.1fh, Night=%.1fh, Naps=%d",
			childID,
			totalSleep.Hours(),
			nightSleep.Hours(),
			naps,
		)
	}

	log.Printf("[SleepAnalyticsJob] Analyzed %d sleep sessions for %d children",
		len(sessions), len(childSleep))

	// TODO: Store aggregated data in a sleep_analytics table for faster dashboard queries
	return nil
}
