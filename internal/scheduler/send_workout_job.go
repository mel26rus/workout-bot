package scheduler

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	"workout-bot/internal/bot"
)

type SendWorkoutJob struct {
	Handler *bot.Handler
	Hour    int
	Minute  int
	Loc     *time.Location
}

func NewSendWorkoutJob(handler *bot.Handler, hour int, minute int, loc *time.Location) *SendWorkoutJob {
	return &SendWorkoutJob{
		Handler: handler,
		Hour:    hour,
		Minute:  minute,
		Loc:     loc,
	}
}

func (j *SendWorkoutJob) Name() string {
	return "Send Workout Job"
}

func (j *SendWorkoutJob) NextRun(now time.Time) time.Time {

	n := now.In(j.Loc)
	next := time.Date(
		n.Year(),
		n.Month(),
		n.Day(),
		j.Hour, j.Minute, 0, 0,
		j.Loc,
	)

	if !next.After(n) {
		next = next.AddDate(0, 0, 1)
	}
	slog.Debug("j.NextRun", "name", j.Name(), "Next", next)
	return next
}

func (j *SendWorkoutJob) Run(ctx context.Context) error {
	fn_name := "SendWorkoutJob.Run"
	slog.Debug(fmt.Sprintf("+%s", fn_name))
	j.Handler.HandleSendWorkouts(ctx)
	slog.Debug(fmt.Sprintf("-%s", fn_name))
	return nil
}
