package scheduler

import (
	"context"
	"time"
)

type Job interface {
	Name() string
	NextRun(now time.Time) time.Time
	Run(ctx context.Context) error
}
