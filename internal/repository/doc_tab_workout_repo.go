package repository

import (
	"context"
	"errors"
	"fmt"

	"workout-bot/internal/model"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
)

// код ошибки Postgres: нарушение внешнего ключа
const pgForeignKeyViolation = "23503"

type DocTabWorkoutRepo struct {
	db *pgxpool.Pool
}

func NewDocTabWorkoutRepo(db *pgxpool.Pool) *DocTabWorkoutRepo {
	return &DocTabWorkoutRepo{db: db}
}

// SaveAll в одной транзакции привязывает посты к дням челленджей.
// День, где уже был другой пост, получает новый (added - новые дни, replaced - заменённые).
// Если пост переезжает на другой день, его прежний день освобождается,
// но только если он ещё не попал в расписание людей.
func (r *DocTabWorkoutRepo) SaveAll(ctx context.Context, list []model.DocTabWorkout) (int, int, error) {
	tx, err := r.db.Begin(ctx)
	if err != nil {
		return 0, 0, err
	}
	defer tx.Rollback(ctx)

	var added, replaced int
	for _, w := range list {
		var curID int64
		var curWorkoutID int64
		var curWorkoutName string
		var curDay int
		err := tx.QueryRow(ctx, `
			SELECT t.id, t.doc_workout_id, dw.name, t.day_number
			FROM doc_tab_workouts t
			JOIN doc_workouts dw ON dw.id = t.doc_workout_id
			WHERE t.channel_post_id = $1
		`, w.ChannelPostID).Scan(&curID, &curWorkoutID, &curWorkoutName, &curDay)
		switch {
		case err == nil:
			if curWorkoutID == w.DocWorkoutID && curDay == w.DayNumber {
				continue
			}
			_, err = tx.Exec(ctx, `DELETE FROM doc_tab_workouts WHERE id = $1`, curID)
			var pgErr *pgconn.PgError
			if errors.As(err, &pgErr) && pgErr.Code == pgForeignKeyViolation {
				return 0, 0, fmt.Errorf("«%s», день %d: %w", curWorkoutName, curDay, model.ErrDayInSchedule)
			}
			if err != nil {
				return 0, 0, err
			}
		case errors.Is(err, pgx.ErrNoRows):
			// пост ещё не был привязан
		default:
			return 0, 0, err
		}

		var inserted bool
		err = tx.QueryRow(ctx, `
			INSERT INTO doc_tab_workouts (doc_workout_id, day_number, channel_post_id)
			VALUES ($1, $2, $3)
			ON CONFLICT (doc_workout_id, day_number) DO UPDATE SET
				channel_post_id = EXCLUDED.channel_post_id
			RETURNING (xmax = 0)
		`, w.DocWorkoutID, w.DayNumber, w.ChannelPostID).Scan(&inserted)
		if err != nil {
			return 0, 0, err
		}
		if inserted {
			added++
		} else {
			replaced++
		}
	}

	return added, replaced, tx.Commit(ctx)
}
