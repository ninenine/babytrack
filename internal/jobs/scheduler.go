package jobs

import (
	"context"
	"log"
	"sync"
	"time"
)

type Scheduler struct {
	jobs    []Job
	wg      sync.WaitGroup
	ctx     context.Context
	cancel  context.CancelFunc
	running bool
	mu      sync.Mutex
}

type Job interface {
	Name() string
	Interval() time.Duration
	Run(ctx context.Context) error
}

func NewScheduler() *Scheduler {
	ctx, cancel := context.WithCancel(context.Background())
	return &Scheduler{
		jobs:   make([]Job, 0),
		ctx:    ctx,
		cancel: cancel,
	}
}

func (s *Scheduler) Register(job Job) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.jobs = append(s.jobs, job)
}

func (s *Scheduler) Start() {
	s.mu.Lock()
	if s.running {
		s.mu.Unlock()
		return
	}
	s.running = true
	s.mu.Unlock()

	for _, job := range s.jobs {
		s.wg.Add(1)
		go s.runJob(job)
	}

	log.Printf("Scheduler started with %d jobs", len(s.jobs))
}

func (s *Scheduler) Stop() {
	s.cancel()
	s.wg.Wait()
	log.Println("Scheduler stopped")
}

func (s *Scheduler) runJob(job Job) {
	defer s.wg.Done()

	ticker := time.NewTicker(job.Interval())
	defer ticker.Stop()

	// Run immediately on start
	if err := job.Run(s.ctx); err != nil {
		log.Printf("Job %s failed: %v", job.Name(), err)
	}

	for {
		select {
		case <-s.ctx.Done():
			return
		case <-ticker.C:
			if err := job.Run(s.ctx); err != nil {
				log.Printf("Job %s failed: %v", job.Name(), err)
			}
		}
	}
}
