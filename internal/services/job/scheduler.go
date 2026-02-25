package job

import (
	"context"
	"fmt"
	"log"
	"sync"
	"time"
)

type Job struct {
	ID        string
	Name      string
	Interval  time.Duration
	Handler   func() error
	LastRun   time.Time
	LastError error
	RunCount  int
	IsRunning bool
	stopChan  chan struct{}
	mu        sync.RWMutex
}

type Scheduler struct {
	jobs      map[string]*Job
	jobChan   chan func() error
	workers   int
	wg        sync.WaitGroup
	ctx       context.Context
	cancel    context.CancelFunc
	isRunning bool
	mu        sync.RWMutex
}

var (
	defaultScheduler *Scheduler
	once             sync.Once
)

func NewScheduler(workers int) *Scheduler {
	ctx, cancel := context.WithCancel(context.Background())
	return &Scheduler{
		jobs:    make(map[string]*Job),
		jobChan: make(chan func() error, 100),
		workers: workers,
		ctx:     ctx,
		cancel:  cancel,
	}
}

func GetScheduler() *Scheduler {
	once.Do(func() {
		defaultScheduler = NewScheduler(5)
	})
	return defaultScheduler
}

func (s *Scheduler) Start() {
	s.mu.Lock()
	if s.isRunning {
		s.mu.Unlock()
		return
	}
	s.isRunning = true
	s.mu.Unlock()

	for i := 0; i < s.workers; i++ {
		s.wg.Add(1)
		go s.worker(i)
	}

	log.Printf("Job scheduler started with %d workers", s.workers)
}

func (s *Scheduler) Stop() {
	s.cancel()

	s.mu.Lock()
	for _, job := range s.jobs {
		if job.stopChan != nil {
			close(job.stopChan)
		}
	}
	s.mu.Unlock()

	s.wg.Wait()

	s.mu.Lock()
	s.isRunning = false
	s.mu.Unlock()

	log.Println("Job scheduler stopped")
}

func (s *Scheduler) worker(id int) {
	defer s.wg.Done()

	log.Printf("Job worker %d started", id)

	for {
		select {
		case <-s.ctx.Done():
			log.Printf("Job worker %d stopping", id)
			return
		case job := <-s.jobChan:
			if err := job(); err != nil {
				log.Printf("Worker %d: Job failed: %v", id, err)
			}
		}
	}
}

func (s *Scheduler) AddPeriodicJob(name string, interval time.Duration, handler func() error) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if _, exists := s.jobs[name]; exists {
		return fmt.Errorf("job %s already exists", name)
	}

	job := &Job{
		ID:       name,
		Name:     name,
		Interval: interval,
		Handler:  handler,
		stopChan: make(chan struct{}),
	}

	s.jobs[name] = job

	go s.runPeriodicJob(job)

	log.Printf("Periodic job '%s' scheduled with interval: %v", name, interval)
	return nil
}

func (s *Scheduler) runPeriodicJob(job *Job) {
	ticker := time.NewTicker(job.Interval)
	defer ticker.Stop()

	for {
		select {
		case <-job.stopChan:
			log.Printf("Job '%s' stopped", job.Name)
			return
		case <-s.ctx.Done():
			return
		case <-ticker.C:
			s.runJob(job)
		}
	}
}

func (s *Scheduler) AddDailyJob(name string, hour, minute int, handler func() error) error {
	now := time.Now()
	loc := now.Location()

	targetTime := time.Date(now.Year(), now.Month(), now.Day(), hour, minute, 0, 0, loc)
	if targetTime.Before(now) {
		targetTime = targetTime.Add(24 * time.Hour)
	}

	interval := targetTime.Sub(now)

	return s.AddPeriodicJob(name, interval, handler)
}

func (s *Scheduler) RunJob(name string) error {
	s.mu.RLock()
	job, ok := s.jobs[name]
	s.mu.RUnlock()

	if !ok {
		return fmt.Errorf("job %s not found", name)
	}

	return s.runJob(job)
}

func (s *Scheduler) runJob(job *Job) error {
	job.mu.Lock()
	if job.IsRunning {
		job.mu.Unlock()
		log.Printf("Job '%s' is already running, skipping", job.Name)
		return nil
	}
	job.IsRunning = true
	job.mu.Unlock()

	log.Printf("Running job: %s", job.Name)
	startTime := time.Now()

	err := job.Handler()

	job.mu.Lock()
	job.LastRun = startTime
	job.IsRunning = false
	job.RunCount++
	if err != nil {
		job.LastError = err
		log.Printf("Job '%s' failed: %v", job.Name, err)
	} else {
		job.LastError = nil
		log.Printf("Job '%s' completed successfully in %v", job.Name, time.Since(startTime))
	}
	job.mu.Unlock()

	return err
}

func (s *Scheduler) GetJob(name string) (*Job, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	job, ok := s.jobs[name]
	return job, ok
}

func (s *Scheduler) ListJobs() []*Job {
	s.mu.RLock()
	defer s.mu.RUnlock()

	jobs := make([]*Job, 0, len(s.jobs))
	for _, job := range s.jobs {
		jobs = append(jobs, job)
	}
	return jobs
}

func (s *Scheduler) GetStatus() map[string]interface{} {
	s.mu.RLock()
	defer s.mu.RUnlock()

	status := make(map[string]interface{})

	jobInfo := make([]map[string]interface{}, 0, len(s.jobs))
	for _, job := range s.jobs {
		job.mu.RLock()
		jobInfo = append(jobInfo, map[string]interface{}{
			"name":       job.Name,
			"interval":   job.Interval.String(),
			"last_run":   job.LastRun,
			"last_error": job.LastError,
			"run_count":  job.RunCount,
			"is_running": job.IsRunning,
		})
		job.mu.RUnlock()
	}

	status["jobs"] = jobInfo
	status["workers"] = s.workers
	status["running"] = s.isRunning

	return status
}

func (s *Scheduler) RemoveJob(name string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	job, ok := s.jobs[name]
	if !ok {
		return fmt.Errorf("job %s not found", name)
	}

	if job.stopChan != nil {
		close(job.stopChan)
	}

	delete(s.jobs, name)
	log.Printf("Job '%s' removed", name)

	return nil
}
