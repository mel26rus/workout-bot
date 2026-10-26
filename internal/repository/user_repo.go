package repository

import (
	"context"

	"workout-bot/internal/model"

	"github.com/jackc/pgx/v5/pgxpool"
)

type UserRepo struct {
	db *pgxpool.Pool
}

func NewUserRepo(db *pgxpool.Pool) *UserRepo {
	return &UserRepo{db: db}
}

// GetOrCreate создаёт пользователя или обновляет его данные из Telegram, isNew == true при создании
func (r *UserRepo) GetOrCreate(ctx context.Context, u model.User) (bool, error) {
	var isNew bool
	err := r.db.QueryRow(ctx, `
		INSERT INTO users (tg_id, username, first_name, surname, language_code)
		VALUES ($1, NULLIF($2, ''), $3, NULLIF($4, ''), NULLIF($5, ''))
		ON CONFLICT (tg_id) DO UPDATE SET
			username      = EXCLUDED.username,
			first_name    = EXCLUDED.first_name,
			surname       = EXCLUDED.surname,
			language_code = EXCLUDED.language_code
		RETURNING (xmax = 0)
	`, u.TgID, u.Username, u.FirstName, u.Surname, u.LanguageCode).Scan(&isNew)

	return isNew, err
}

// SetDocWorkout запоминает выбранный человеком челлендж
func (r *UserRepo) SetDocWorkout(ctx context.Context, tgID int64, docWorkoutID int64) error {
	tag, err := r.db.Exec(ctx, `
		UPDATE users SET doc_workout_id = $2 WHERE tg_id = $1
	`, tgID, docWorkoutID)
	if err != nil {
		return err
	}
	if tag.RowsAffected() == 0 {
		return model.ErrUserNotFound
	}

	return nil
}

func (r *UserRepo) GetAdminTgIDs(ctx context.Context) ([]int64, error) {
	rows, err := r.db.Query(ctx, `
		SELECT tg_id FROM users WHERE is_admin ORDER BY id
	`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var list []int64
	for rows.Next() {
		var id int64
		if err := rows.Scan(&id); err != nil {
			return nil, err
		}
		list = append(list, id)
	}

	return list, rows.Err()
}

// GetParticipants возвращает занимающихся: сначала ждущих оплаты, потом оплативших с прогрессом
func (r *UserRepo) GetParticipants(ctx context.Context) ([]model.Participant, error) {
	rows, err := r.db.Query(ctx, `
		SELECT u.id, u.tg_id, COALESCE(u.username, ''), u.first_name, COALESCE(u.surname, ''),
		       dw.name, true AS is_paid, uw.start_date,
		       COUNT(tuw.id), COUNT(tuw.sent_at), COUNT(tuw.done_at)
		FROM doc_user_workouts uw
		JOIN users u ON u.id = uw.user_id
		JOIN doc_workouts dw ON dw.id = uw.doc_workout_id
		LEFT JOIN doc_tab_user_workouts tuw ON tuw.doc_user_workout_id = uw.id
		GROUP BY u.id, dw.id, uw.id

		UNION ALL

		SELECT u.id, u.tg_id, COALESCE(u.username, ''), u.first_name, COALESCE(u.surname, ''),
		       dw.name, false, NULL::date, 0::bigint, 0::bigint, 0::bigint
		FROM users u
		JOIN doc_workouts dw ON dw.id = u.doc_workout_id
		WHERE NOT EXISTS (
			SELECT 1 FROM doc_user_workouts uw
			WHERE uw.user_id = u.id AND uw.doc_workout_id = u.doc_workout_id
		)

		ORDER BY 7, 1
	`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var list []model.Participant
	for rows.Next() {
		var p model.Participant
		err := rows.Scan(
			&p.UserID, &p.TgID, &p.Username, &p.FirstName, &p.Surname,
			&p.DocWorkoutName, &p.IsPaid, &p.StartDate,
			&p.Total, &p.Sent, &p.Done,
		)
		if err != nil {
			return nil, err
		}
		list = append(list, p)
	}

	return list, rows.Err()
}
