package repository

import (
	"context"
	"errors"
	"time"

	"workout-bot/internal/model"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type DocUserWorkoutRepo struct {
	db *pgxpool.Pool
}

func NewDocUserWorkoutRepo(db *pgxpool.Pool) *DocUserWorkoutRepo {
	return &DocUserWorkoutRepo{db: db}
}

// ConfirmPayment в одной транзакции создаёт doc_user_workouts для выбранного челленджа
// и расписание doc_tab_user_workouts: send_date = startDate + day_number - 1
func (r *DocUserWorkoutRepo) ConfirmPayment(ctx context.Context, userID int64, startDate time.Time) (model.PaymentConfirmation, error) {
	result := model.PaymentConfirmation{StartDate: startDate}

	tx, err := r.db.Begin(ctx)
	if err != nil {
		return result, err
	}
	defer tx.Rollback(ctx)

	var docWorkoutID *int64
	err = tx.QueryRow(ctx, `
		SELECT tg_id, doc_workout_id FROM users WHERE id = $1 FOR UPDATE
	`, userID).Scan(&result.TgID, &docWorkoutID)
	if errors.Is(err, pgx.ErrNoRows) {
		return result, model.ErrUserNotFound
	}
	if err != nil {
		return result, err
	}
	if docWorkoutID == nil {
		return result, model.ErrNoChallengeSelected
	}

	var docUserWorkoutID int64
	err = tx.QueryRow(ctx, `
		INSERT INTO doc_user_workouts (user_id, doc_workout_id, start_date)
		VALUES ($1, $2, $3)
		ON CONFLICT (user_id, doc_workout_id) DO NOTHING
		RETURNING id
	`, userID, *docWorkoutID, startDate).Scan(&docUserWorkoutID)
	if errors.Is(err, pgx.ErrNoRows) {
		return result, model.ErrAlreadyPaid
	}
	if err != nil {
		return result, err
	}

	_, err = tx.Exec(ctx, `
		INSERT INTO doc_tab_user_workouts (doc_user_workout_id, doc_tab_workout_id, send_date)
		SELECT $1, id, $2::date + (day_number - 1)
		FROM doc_tab_workouts
		WHERE doc_workout_id = $3
	`, docUserWorkoutID, startDate, *docWorkoutID)
	if err != nil {
		return result, err
	}

	err = tx.QueryRow(ctx, `
		SELECT name FROM doc_workouts WHERE id = $1
	`, *docWorkoutID).Scan(&result.DocWorkoutName)
	if err != nil {
		return result, err
	}

	return result, tx.Commit(ctx)
}
