package jobs

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/ninenine/babytrack/internal/notifications"
	"github.com/ninenine/babytrack/internal/sleep"

	"github.com/google/uuid"
)

// Age-appropriate sleep recommendations (in hours)
var sleepRecommendations = map[string]struct {
	minHours float64
	maxHours float64
}{
	"newborn": {14, 17}, // 0-3 months
	"infant":  {12, 15}, // 4-11 months
	"toddler": {11, 14}, // 1-2 years
	"pre-k":   {10, 13}, // 3-5 years
	"child":   {9, 11},  // 6-12 years
	"default": {10, 14}, // Default for babies
}

// SleepAnalyticsJob calculates daily sleep statistics and sends insights.
type SleepAnalyticsJob struct {
	sleepService    sleep.Service
	notificationHub *notifications.Hub
}

func NewSleepAnalyticsJob(sleepService sleep.Service) *SleepAnalyticsJob {
	return &SleepAnalyticsJob{
		sleepService: sleepService,
	}
}

// WithNotificationHub adds notification capability to the job
func (j *SleepAnalyticsJob) WithNotificationHub(hub *notifications.Hub) *SleepAnalyticsJob {
	j.notificationHub = hub
	return j
}

func (j *SleepAnalyticsJob) Name() string {
	return "sleep-analytics"
}

func (j *SleepAnalyticsJob) Interval() time.Duration {
	return 1 * time.Hour // Check every hour
}

func (j *SleepAnalyticsJob) Run(ctx context.Context) error {
	log.Println("[SleepAnalyticsJob] Running sleep analytics...")

	now := time.Now()

	// Check for long ongoing sleeps first
	j.checkLongSleeps(ctx, now)

	// Only run daily summary once per day (between 7-8 AM)
	if now.Hour() == 7 {
		j.runDailySummary(ctx, now)
	}

	return nil
}

// checkLongSleeps alerts if a sleep session has been going for too long
func (j *SleepAnalyticsJob) checkLongSleeps(ctx context.Context, now time.Time) {
	// Get all sleep sessions without end time (ongoing)
	sessions, err := j.sleepService.List(ctx, &sleep.SleepFilter{})
	if err != nil {
		log.Printf("[SleepAnalyticsJob] Error fetching sleep sessions: %v", err)
		return
	}

	for _, session := range sessions {
		if session.EndTime != nil {
			continue // Skip completed sessions
		}

		duration := now.Sub(session.StartTime)

		// Alert thresholds
		var alertMessage string
		if session.Type == sleep.SleepTypeNap && duration > 3*time.Hour {
			alertMessage = fmt.Sprintf("Nap has been going for %.1f hours", duration.Hours())
		} else if session.Type == sleep.SleepTypeNight && duration > 14*time.Hour {
			alertMessage = fmt.Sprintf("Night sleep has been going for %.1f hours", duration.Hours())
		}

		if alertMessage != "" {
			log.Printf("[SleepAnalyticsJob] Long sleep alert: %s (Child: %s)", alertMessage, session.ChildID)

			if j.notificationHub != nil && j.notificationHub.ClientCount() > 0 {
				j.notificationHub.Broadcast(notifications.Event{
					ID:        uuid.New().String(),
					Type:      notifications.EventSleepInsight,
					Title:     "Sleep Alert",
					Message:   alertMessage,
					ChildID:   session.ChildID,
					Timestamp: now,
				})
			}
		}
	}
}

// runDailySummary calculates yesterday's sleep and sends insights
func (j *SleepAnalyticsJob) runDailySummary(ctx context.Context, now time.Time) {
	log.Println("[SleepAnalyticsJob] Generating daily sleep summary...")

	// Get yesterday's date range
	yesterday := now.AddDate(0, 0, -1)
	startOfYesterday := time.Date(yesterday.Year(), yesterday.Month(), yesterday.Day(), 0, 0, 0, 0, yesterday.Location())
	endOfYesterday := startOfYesterday.AddDate(0, 0, 1)

	// Get all sleep sessions that started yesterday
	sessions, err := j.sleepService.List(ctx, &sleep.SleepFilter{
		StartDate: &startOfYesterday,
		EndDate:   &endOfYesterday,
	})
	if err != nil {
		log.Printf("[SleepAnalyticsJob] Error fetching sleep sessions: %v", err)
		return
	}

	// Group by child and calculate totals
	type childStats struct {
		totalSleep time.Duration
		nightSleep time.Duration
		napCount   int
		napTime    time.Duration
	}
	childData := make(map[string]*childStats)

	for _, session := range sessions {
		if session.EndTime == nil {
			continue // Skip incomplete sessions
		}

		if _, exists := childData[session.ChildID]; !exists {
			childData[session.ChildID] = &childStats{}
		}

		duration := session.EndTime.Sub(session.StartTime)
		childData[session.ChildID].totalSleep += duration

		if session.Type == sleep.SleepTypeNap {
			childData[session.ChildID].napCount++
			childData[session.ChildID].napTime += duration
		} else {
			childData[session.ChildID].nightSleep += duration
		}
	}

	// Generate insights for each child
	rec := sleepRecommendations["default"]

	for childID, stats := range childData {
		totalHours := stats.totalSleep.Hours()

		log.Printf("[SleepAnalyticsJob] Child %s: Total=%.1fh, Night=%.1fh, Naps=%d (%.1fh)",
			childID, totalHours, stats.nightSleep.Hours(), stats.napCount, stats.napTime.Hours())

		// Generate insight message
		var message string
		var title string

		if totalHours >= rec.minHours && totalHours <= rec.maxHours {
			title = "Great Sleep!"
			message = fmt.Sprintf("Slept %.1f hours yesterday - right on track!", totalHours)
		} else if totalHours < rec.minHours {
			title = "Sleep Summary"
			deficit := rec.minHours - totalHours
			message = fmt.Sprintf("Slept %.1f hours yesterday (%.1f hours below recommended)", totalHours, deficit)
		} else {
			title = "Sleep Summary"
			message = fmt.Sprintf("Slept %.1f hours yesterday - plenty of rest!", totalHours)
		}

		// Add nap info if applicable
		if stats.napCount > 0 {
			message += fmt.Sprintf(" Had %d nap(s).", stats.napCount)
		}

		if j.notificationHub != nil && j.notificationHub.ClientCount() > 0 {
			j.notificationHub.Broadcast(notifications.Event{
				ID:        uuid.New().String(),
				Type:      notifications.EventSleepInsight,
				Title:     title,
				Message:   message,
				ChildID:   childID,
				Timestamp: now,
			})
		}
	}

	log.Printf("[SleepAnalyticsJob] Daily summary complete. Analyzed %d children", len(childData))
}
