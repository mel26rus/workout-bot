package scheduler

import (
	"context"
	"log/slog"
	"time"
)

type scheduledJob struct {
	job     Job
	nextRun time.Time
}

type Scheduler struct {
	jobs []*scheduledJob
}

func New() *Scheduler {
	return &Scheduler{
		jobs: make([]*scheduledJob, 0),
	}
}

func (s *Scheduler) Add(job Job) {

	s.jobs = append(s.jobs, &scheduledJob{
		job:     job,
		nextRun: job.NextRun(time.Now()),
	})
}

func (s *Scheduler) Run(ctx context.Context) {

	slog.Debug("Scheduler Run")
	// if s.jobs == nil {
	// 	slog.Debug("test", "jobs", nil)
	// } else {
	// 	slog.Debug("test", "jobs", s.jobs)
	// }

	for {

		var next *scheduledJob

		for _, job := range s.jobs {

			if next == nil || job.nextRun.Before(next.nextRun) {
				next = job
			}
		}

		timer := time.NewTimer(time.Until(next.nextRun))

		select {

		case <-ctx.Done():

			timer.Stop()
			slog.Info("Scheduler stopped")
			return

		case <-timer.C:

			slog.Debug(
				"Running job",
				"name", next.job.Name(),
			)

			oldNextRun := next.nextRun
			next.nextRun = next.job.NextRun(oldNextRun)

			go func(job *scheduledJob) {

				if err := job.job.Run(ctx); err != nil {
					slog.Error(
						"Job failed",
						"name", job.job.Name(),
						"error", err,
					)
				}

			}(next)
		}
	}
}
