package repository

import (
	"context"
	"errors"

	"workout-bot/internal/model"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type DocWorkoutRepo struct {
	db *pgxpool.Pool
}

func NewDocWorkoutRepo(db *pgxpool.Pool) *DocWorkoutRepo {
	return &DocWorkoutRepo{db: db}
}

func (r *DocWorkoutRepo) GetAll(ctx context.Context) ([]model.DocWorkout, error) {
	rows, err := r.db.Query(ctx, `
		SELECT id, name, description, payment_url
		FROM doc_workouts
		ORDER BY id
	`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var list []model.DocWorkout
	for rows.Next() {
		var w model.DocWorkout
		if err := rows.Scan(&w.ID, &w.Name, &w.Description, &w.PaymentURL); err != nil {
			return nil, err
		}
		list = append(list, w)
	}

	return list, rows.Err()
}

func (r *DocWorkoutRepo) GetByID(ctx context.Context, id int64) (model.DocWorkout, error) {
	var w model.DocWorkout
	err := r.db.QueryRow(ctx, `
		SELECT id, name, description, payment_url
		FROM doc_workouts
		WHERE id = $1
	`, id).Scan(&w.ID, &w.Name, &w.Description, &w.PaymentURL)
	if errors.Is(err, pgx.ErrNoRows) {
		return w, model.ErrChallengeNotFound
	}

	return w, err
}

func (r *DocWorkoutRepo) Create(ctx context.Context, w model.DocWorkout) (model.DocWorkout, error) {
	err := r.db.QueryRow(ctx, `
		INSERT INTO doc_workouts (name, description, payment_url)
		VALUES ($1, $2, $3)
		RETURNING id
	`, w.Name, w.Description, w.PaymentURL).Scan(&w.ID)

	return w, err
}
