package repository

import (
	"context"
	"time"

	"workout-bot/internal/model"

	"github.com/jackc/pgx/v5/pgxpool"
)

type DocTabUserWorkoutRepo struct {
	db *pgxpool.Pool
}

func NewDocTabUserWorkoutRepo(db *pgxpool.Pool) *DocTabUserWorkoutRepo {
	return &DocTabUserWorkoutRepo{db: db}
}

// GetDue возвращает тренировки, которые надо отправить на дату date и ещё не отправлены.
// Сам файл не грузим, он нужен только если у тренировки ещё нет tg_file_id.
func (r *DocTabUserWorkoutRepo) GetDue(ctx context.Context, date time.Time) ([]model.DueWorkout, error) {
	rows, err := r.db.Query(ctx, `
		SELECT tuw.id, u.tg_id, t.id, t.day_number, t.description,
		       COALESCE(t.file_name, ''), COALESCE(t.mime_type, ''), COALESCE(t.tg_file_id, ''),
		       (t.file IS NOT NULL)
		FROM doc_tab_user_workouts tuw
		JOIN doc_user_workouts uw ON uw.id = tuw.doc_user_workout_id
		JOIN users u ON u.id = uw.user_id
		JOIN doc_tab_workouts t ON t.id = tuw.doc_tab_workout_id
		WHERE tuw.send_date = $1 AND tuw.sent_at IS NULL
		ORDER BY t.id, tuw.id
	`, date)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var list []model.DueWorkout
	for rows.Next() {
		var w model.DueWorkout
		err := rows.Scan(
			&w.ID, &w.TgID, &w.DocTabWorkoutID, &w.DayNumber, &w.Description,
			&w.FileName, &w.MimeType, &w.TgFileID, &w.HasFile,
		)
		if err != nil {
			return nil, err
		}
		list = append(list, w)
	}

	return list, rows.Err()
}

func (r *DocTabUserWorkoutRepo) MarkSent(ctx context.Context, id int64) error {
	_, err := r.db.Exec(ctx, `
		UPDATE doc_tab_user_workouts SET sent_at = now() WHERE id = $1
	`, id)

	return err
}

// MarkDone отмечает выполнение, только если запись принадлежит этому пользователю Telegram
func (r *DocTabUserWorkoutRepo) MarkDone(ctx context.Context, id int64, tgID int64) error {
	tag, err := r.db.Exec(ctx, `
		UPDATE doc_tab_user_workouts tuw
		SET done_at = COALESCE(tuw.done_at, now())
		FROM doc_user_workouts uw
		JOIN users u ON u.id = uw.user_id
		WHERE tuw.id = $1 AND uw.id = tuw.doc_user_workout_id AND u.tg_id = $2
	`, id, tgID)
	if err != nil {
		return err
	}
	if tag.RowsAffected() == 0 {
		return model.ErrWorkoutDayNotFound
	}

	return nil
}
