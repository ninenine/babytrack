package jobs

import (
	"context"
	"sync/atomic"
	"testing"
	"time"
)

// mockJob is a test double for Job interface
type mockJob struct {
	name     string
	interval time.Duration
	runCount int32
	runErr   error
}

func newMockJob(name string, interval time.Duration) *mockJob {
	return &mockJob{
		name:     name,
		interval: interval,
	}
}

func (m *mockJob) Name() string {
	return m.name
}

func (m *mockJob) Interval() time.Duration {
	return m.interval
}

func (m *mockJob) Run(ctx context.Context) error {
	atomic.AddInt32(&m.runCount, 1)
	return m.runErr
}

func (m *mockJob) RunCount() int {
	return int(atomic.LoadInt32(&m.runCount))
}

func TestNewScheduler(t *testing.T) {
	scheduler := NewScheduler()

	if scheduler == nil {
		t.Fatal("NewScheduler() returned nil")
	}

	if scheduler.jobs == nil {
		t.Error("NewScheduler() jobs should be initialized")
	}

	if scheduler.running {
		t.Error("NewScheduler() should not be running initially")
	}
}

func TestScheduler_Register(t *testing.T) {
	scheduler := NewScheduler()

	job1 := newMockJob("job1", time.Minute)
	job2 := newMockJob("job2", time.Hour)

	scheduler.Register(job1)
	scheduler.Register(job2)

	if len(scheduler.jobs) != 2 {
		t.Errorf("Scheduler should have 2 jobs, got %d", len(scheduler.jobs))
	}
}

func TestScheduler_StartAndStop(t *testing.T) {
	scheduler := NewScheduler()

	job := newMockJob("test-job", 50*time.Millisecond)
	scheduler.Register(job)

	scheduler.Start()

	// Wait for the job to run at least once (runs immediately on start)
	time.Sleep(20 * time.Millisecond)

	if !scheduler.running {
		t.Error("Scheduler should be running after Start()")
	}

	scheduler.Stop()

	if job.RunCount() < 1 {
		t.Errorf("Job should have run at least once, ran %d times", job.RunCount())
	}
}

func TestScheduler_StartTwice(t *testing.T) {
	scheduler := NewScheduler()

	job := newMockJob("test-job", time.Hour)
	scheduler.Register(job)

	scheduler.Start()
	scheduler.Start() // Should be a no-op

	time.Sleep(10 * time.Millisecond)
	scheduler.Stop()

	// Should only have run once (immediate run on first Start)
	if job.RunCount() != 1 {
		t.Errorf("Job should have run exactly once, ran %d times", job.RunCount())
	}
}

func TestScheduler_JobRunsOnInterval(t *testing.T) {
	scheduler := NewScheduler()

	job := newMockJob("interval-job", 30*time.Millisecond)
	scheduler.Register(job)

	scheduler.Start()

	// Wait for multiple intervals
	time.Sleep(100 * time.Millisecond)

	scheduler.Stop()

	// Should have run multiple times (immediate + interval runs)
	if job.RunCount() < 2 {
		t.Errorf("Job should have run multiple times, ran %d times", job.RunCount())
	}
}

func TestScheduler_MultipleJobs(t *testing.T) {
	scheduler := NewScheduler()

	job1 := newMockJob("fast-job", 20*time.Millisecond)
	job2 := newMockJob("slow-job", 100*time.Millisecond)

	scheduler.Register(job1)
	scheduler.Register(job2)

	scheduler.Start()
	time.Sleep(80 * time.Millisecond)
	scheduler.Stop()

	// Fast job should run more times than slow job
	if job1.RunCount() <= job2.RunCount() {
		t.Errorf("Fast job (%d) should run more than slow job (%d)", job1.RunCount(), job2.RunCount())
	}

	// Both should have run at least once
	if job1.RunCount() < 1 || job2.RunCount() < 1 {
		t.Errorf("Both jobs should run at least once: job1=%d, job2=%d", job1.RunCount(), job2.RunCount())
	}
}

func TestScheduler_StopWaitsForJobs(t *testing.T) {
	scheduler := NewScheduler()

	job := newMockJob("test-job", time.Hour)
	scheduler.Register(job)

	scheduler.Start()
	time.Sleep(10 * time.Millisecond)

	// Stop should wait for jobs to complete
	done := make(chan bool)
	go func() {
		scheduler.Stop()
		done <- true
	}()

	select {
	case <-done:
		// Success - Stop completed
	case <-time.After(1 * time.Second):
		t.Error("Stop() should complete within timeout")
	}
}

func TestScheduler_EmptyJobs(t *testing.T) {
	scheduler := NewScheduler()

	// Starting with no jobs should not panic
	scheduler.Start()
	time.Sleep(10 * time.Millisecond)
	scheduler.Stop()
}
