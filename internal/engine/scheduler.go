package engine

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/robfig/cron/v3"
)

type Scheduler struct {
	cron      *cron.Cron
	flows     map[string]*Flow
	schedules map[string]cron.EntryID
	mu        sync.RWMutex
	ctx       context.Context
	cancel    context.CancelFunc
}

type Schedule struct {
	FlowID   string
	CronExpr string
	Interval time.Duration
	Type     string // "cron", "interval", "webhook"
	Enabled  bool
}

func NewScheduler() *Scheduler {
	ctx, cancel := context.WithCancel(context.Background())

	return &Scheduler{
		cron:      cron.New(),
		flows:     make(map[string]*Flow),
		schedules: make(map[string]cron.EntryID),
		ctx:       ctx,
		cancel:    cancel,
	}
}

func (s *Scheduler) Start() error {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.cron.Start()
	return nil
}

func (s *Scheduler) Stop() error {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.cancel()
	s.cron.Stop()
	return nil
}

func (s *Scheduler) AddCronTrigger(flowID, cronExpr string, flow *Flow) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if _, exists := s.schedules[flowID]; exists {
		return fmt.Errorf("schedule already exists for flow %s", flowID)
	}

	entryID, err := s.cron.AddFunc(cronExpr, func() {
		s.executeFlow(flowID)
	})

	if err != nil {
		return fmt.Errorf("failed to add cron trigger: %w", err)
	}

	s.flows[flowID] = flow
	s.schedules[flowID] = entryID

	return nil
}

func (s *Scheduler) AddIntervalTrigger(flowID string, interval time.Duration, flow *Flow) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if _, exists := s.schedules[flowID]; exists {
		return fmt.Errorf("schedule already exists for flow %s", flowID)
	}

	cronExpr := fmt.Sprintf("@every %s", interval.String())
	entryID, err := s.cron.AddFunc(cronExpr, func() {
		s.executeFlow(flowID)
	})

	if err != nil {
		return fmt.Errorf("failed to add interval trigger: %w", err)
	}

	s.flows[flowID] = flow
	s.schedules[flowID] = entryID

	return nil
}

func (s *Scheduler) RemoveTrigger(flowID string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	entryID, exists := s.schedules[flowID]
	if !exists {
		return fmt.Errorf("no schedule found for flow %s", flowID)
	}

	s.cron.Remove(entryID)
	delete(s.schedules, flowID)
	delete(s.flows, flowID)

	return nil
}

func (s *Scheduler) executeFlow(flowID string) {
	s.mu.RLock()
	flow, exists := s.flows[flowID]
	s.mu.RUnlock()

	if !exists {
		return
	}

	if err := flow.Start(s.ctx); err != nil {
		fmt.Printf("Failed to execute scheduled flow %s: %v\n", flowID, err)
	}
}

func (s *Scheduler) GetSchedules() []Schedule {
	s.mu.RLock()
	defer s.mu.RUnlock()

	schedules := make([]Schedule, 0, len(s.schedules))

	for flowID := range s.schedules {
		schedules = append(schedules, Schedule{
			FlowID:  flowID,
			Enabled: true,
			Type:    "cron",
		})
	}

	return schedules
}

func (s *Scheduler) UpdateSchedule(flowID string, schedule Schedule) error {
	if err := s.RemoveTrigger(flowID); err != nil {
		return err
	}

	s.mu.RLock()
	flow := s.flows[flowID]
	s.mu.RUnlock()

	if flow == nil {
		return fmt.Errorf("flow %s not found", flowID)
	}

	switch schedule.Type {
	case "cron":
		return s.AddCronTrigger(flowID, schedule.CronExpr, flow)
	case "interval":
		return s.AddIntervalTrigger(flowID, schedule.Interval, flow)
	default:
		return fmt.Errorf("unsupported schedule type: %s", schedule.Type)
	}
}
