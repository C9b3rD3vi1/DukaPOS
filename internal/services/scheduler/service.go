package scheduler

import (
	"log"
	"sync"
	"time"
)

// Task represents a scheduled task
type Task struct {
	Name        string
	Schedule    time.Duration // or cron expression
	IsCron      bool
	CronExpr    string
	Handler     func() error
	LastRun     time.Time
	NextRun     time.Time
	IsRunning   bool
	IsActive    bool
}

// Scheduler manages scheduled tasks
type Scheduler struct {
	tasks   map[string]*Task
	mu      sync.RWMutex
	stopCh  chan struct{}
	wg      sync.WaitGroup
}

// New creates a new scheduler
func New() *Scheduler {
	return &Scheduler{
		tasks:  make(map[string]*Task),
		stopCh: make(chan struct{}),
	}
}

// AddTask adds a new task to the scheduler
func (s *Scheduler) AddTask(name string, interval time.Duration, handler func() error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	task := &Task{
		Name:      name,
		Schedule:  interval,
		Handler:   handler,
		NextRun:   time.Now().Add(interval),
		IsActive:  true,
	}

	s.tasks[name] = task
	log.Printf("📅 Scheduled task: %s every %v", name, interval)
}

// AddCronTask adds a cron-based task
func (s *Scheduler) AddCronTask(name, cronExpr string, handler func() error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	task := &Task{
		Name:     name,
		IsCron:   true,
		CronExpr: cronExpr,
		Handler:  handler,
		IsActive: true,
		// Note: Full cron support would need a cron parser library
		// For now, using simple interval-based approach
		NextRun: time.Now().Add(1 * time.Hour), // Default
	}

	s.tasks[name] = task
	log.Printf("📅 Scheduled cron task: %s (%s)", name, cronExpr)
}

// RemoveTask removes a task from scheduler
func (s *Scheduler) RemoveTask(name string) {
	s.mu.Lock()
	defer s.mu.Unlock()

	if task, exists := s.tasks[name]; exists {
		task.IsActive = false
		delete(s.tasks, name)
		log.Printf("📅 Removed task: %s", name)
	}
}

// Start begins executing scheduled tasks
func (s *Scheduler) Start() {
	s.wg.Add(1)
	go s.run()
	log.Printf("📅 Scheduler started with %d tasks", len(s.tasks))
}

// Stop halts the scheduler
func (s *Scheduler) Stop() {
	close(s.stopCh)
	s.wg.Wait()
	log.Printf("📅 Scheduler stopped")
}

func (s *Scheduler) run() {
	defer s.wg.Done()

	ticker := time.NewTicker(1 * time.Minute)
	defer ticker.Stop()

	for {
		select {
		case <-s.stopCh:
			return
		case <-ticker.C:
			s.runPendingTasks()
		}
	}
}

func (s *Scheduler) runPendingTasks() {
	now := time.Now()

	s.mu.RLock()
	// Copy task names to avoid lock during execution
	taskNames := make([]string, 0, len(s.tasks))
	for name, task := range s.tasks {
		if task.IsActive && now.After(task.NextRun) && !task.IsRunning {
			taskNames = append(taskNames, name)
		}
	}
	s.mu.RUnlock()

	for _, name := range taskNames {
		s.runTask(name)
	}
}

func (s *Scheduler) runTask(name string) {
	s.mu.Lock()
	task, exists := s.tasks[name]
	if !exists {
		s.mu.Unlock()
		return
	}

	if task.IsRunning {
		s.mu.Unlock()
		log.Printf("⚠️ Task %s already running, skipping", name)
		return
	}

	task.IsRunning = true
	task.LastRun = time.Now()
	task.NextRun = task.LastRun.Add(task.Schedule)
	s.mu.Unlock()

	log.Printf("📅 Running task: %s", name)

	err := task.Handler()
	if err != nil {
		log.Printf("❌ Task %s failed: %v", name, err)
		// Don't update next run on failure - retry sooner
		s.mu.Lock()
		task.NextRun = time.Now().Add(1 * time.Minute)
		s.mu.Unlock()
	} else {
		log.Printf("✅ Task %s completed", name)
	}

	s.mu.Lock()
	task.IsRunning = false
	s.mu.Unlock()
}

// GetTaskStatus returns status of all tasks
func (s *Scheduler) GetTaskStatus() []map[string]interface{} {
	s.mu.RLock()
	defer s.mu.RUnlock()

	status := make([]map[string]interface{}, 0, len(s.tasks))
	for _, task := range s.tasks {
		status = append(status, map[string]interface{}{
			"name":        task.Name,
			"is_active":  task.IsActive,
			"is_running":  task.IsRunning,
			"last_run":   task.LastRun,
			"next_run":   task.NextRun,
			"schedule":   task.Schedule.String(),
		})
	}
	return status
}

// DailyTask creates a task that runs once daily
func DailyTask(name string, hour, minute int, handler func() error) {
	now := time.Now()
	nextRun := time.Date(now.Year(), now.Month(), now.Day(), hour, minute, 0, 0, now.Location())
	
	if nextRun.Before(now) {
		nextRun = nextRun.Add(24 * time.Hour)
	}

	_ = nextRun.Sub(now) // Calculate duration but don't use
	
	// Create a recursive task that reschedules itself daily
	sched := New()
	
	var createDailyTask func()
	createDailyTask = func() {
		sched.AddTask(name, 24*time.Hour, func() error {
			err := handler()
			// Reschedule for next day
			createDailyTask()
			return err
		})
	}
	
	createDailyTask()
	sched.Start()
}

// CommonTasks creates default system tasks
func CommonTasks(
	cleanup func() error,
	backup func() error,
	report func() error,
) *Scheduler {
	sched := New()

	// Clean up expired sessions every hour
	sched.AddTask("cleanup_expired", 1*time.Hour, func() error {
		log.Println("🧹 Running cleanup task...")
		return cleanup()
	})

	// Backup database daily at 2 AM
	sched.AddTask("backup_database", 24*time.Hour, func() error {
		log.Println("💾 Running backup task...")
		return backup()
	})

	// Send daily reports at 8 PM
	sched.AddTask("daily_reports", 24*time.Hour, func() error {
		log.Println("📊 Running daily report task...")
		return report()
	})

	return sched
}
