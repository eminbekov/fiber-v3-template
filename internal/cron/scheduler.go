package cron

import (
	"context"
	"log/slog"
	"time"
)

type Job struct {
	Name     string
	Schedule time.Duration
	Run      func(ctx context.Context) error
}

type Scheduler struct {
	jobs []Job
}

func NewScheduler() *Scheduler {
	return &Scheduler{}
}

func (scheduler *Scheduler) Register(job Job) {
	if job.Schedule <= 0 {
		slog.Warn("cron job skipped because schedule is not positive", "name", job.Name, "schedule", job.Schedule.String())
		return
	}
	if job.Run == nil {
		slog.Warn("cron job skipped because run function is nil", "name", job.Name)
		return
	}

	scheduler.jobs = append(scheduler.jobs, job)
}

func (scheduler *Scheduler) Start(ctx context.Context) {
	for _, registeredJob := range scheduler.jobs {
		go scheduler.runJob(ctx, registeredJob)
	}
}

func (scheduler *Scheduler) runJob(ctx context.Context, job Job) {
	ticker := time.NewTicker(job.Schedule)
	defer ticker.Stop()

	slog.Info("cron job registered", "name", job.Name, "schedule", job.Schedule.String())

	for {
		select {
		case <-ctx.Done():
			slog.Info("cron job stopped", "name", job.Name)
			return
		case <-ticker.C:
			slog.Info("cron job starting", "name", job.Name)
			startTime := time.Now()

			if runError := job.Run(ctx); runError != nil {
				slog.Error("cron job failed", "name", job.Name, "error", runError, "duration", time.Since(startTime).String())
				continue
			}

			slog.Info("cron job completed", "name", job.Name, "duration", time.Since(startTime).String())
		}
	}
}
