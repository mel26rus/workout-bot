package scheduler

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	"workout-bot/internal/bot"
	"workout-bot/internal/service"
)

type SendWorkoutJob struct {
	Handler  *bot.Handler
	Settings *service.SettingsService
	Loc      *time.Location
}

func NewSendWorkoutJob(handler *bot.Handler, settings *service.SettingsService, loc *time.Location) *SendWorkoutJob {
	return &SendWorkoutJob{
		Handler:  handler,
		Settings: settings,
		Loc:      loc,
	}
}

func (j *SendWorkoutJob) Name() string {
	return "Send Workout Job"
}

// sendTime читает время рассылки при каждом расчёте, поэтому смена времени в веб-модуле подхватывается без перезапуска
func (j *SendWorkoutJob) sendTime() (int, int, error) {
	value, err := j.Settings.GetSendTime(context.Background())
	if err != nil {
		return 0, 0, err
	}

	return service.ParseSendTime(value)
}

func (j *SendWorkoutJob) NextRun(now time.Time) time.Time {

	n := now.In(j.Loc)

	hour, minute, err := j.sendTime()
	if err != nil {
		// не смогли прочитать время: пробуем снова через час, планировщик не должен остановиться
		slog.Error("j.NextRun sendTime", "name", j.Name(), "Error", err)
		return n.Add(time.Hour)
	}

	next := time.Date(
		n.Year(),
		n.Month(),
		n.Day(),
		hour, minute, 0, 0,
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
