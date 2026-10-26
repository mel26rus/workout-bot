package repository

import (
	"context"

	"workout-bot/internal/model"

	"github.com/jackc/pgx/v5/pgxpool"
)

type DocTabWorkoutRepo struct {
	db *pgxpool.Pool
}

func NewDocTabWorkoutRepo(db *pgxpool.Pool) *DocTabWorkoutRepo {
	return &DocTabWorkoutRepo{db: db}
}

// Save создаёт тренировку дня или обновляет существующую, replaced == true при обновлении.
// Если файл не приложен (File == nil), старый файл остаётся.
// Если приложен новый, tg_file_id сбрасывается.
func (r *DocTabWorkoutRepo) Save(ctx context.Context, w model.DocTabWorkout) (bool, error) {
	var inserted bool
	err := r.db.QueryRow(ctx, `
		INSERT INTO doc_tab_workouts (doc_workout_id, day_number, description, file, file_name, mime_type)
		VALUES ($1, $2, $3, $4, NULLIF($5, ''), NULLIF($6, ''))
		ON CONFLICT (doc_workout_id, day_number) DO UPDATE SET
			description = EXCLUDED.description,
			tg_file_id  = CASE WHEN EXCLUDED.file IS NOT NULL THEN NULL ELSE doc_tab_workouts.tg_file_id END,
			file_name   = CASE WHEN EXCLUDED.file IS NOT NULL THEN EXCLUDED.file_name ELSE doc_tab_workouts.file_name END,
			mime_type   = CASE WHEN EXCLUDED.file IS NOT NULL THEN EXCLUDED.mime_type ELSE doc_tab_workouts.mime_type END,
			file        = COALESCE(EXCLUDED.file, doc_tab_workouts.file)
		RETURNING (xmax = 0)
	`, w.DocWorkoutID, w.DayNumber, w.Description, w.File, w.FileName, w.MimeType).Scan(&inserted)

	return !inserted, err
}

func (r *DocTabWorkoutRepo) GetAllInfo(ctx context.Context) ([]model.DocTabWorkoutInfo, error) {
	rows, err := r.db.Query(ctx, `
		SELECT t.id, t.doc_workout_id, dw.name, t.day_number, t.description,
		       COALESCE(t.file_name, ''), COALESCE(octet_length(t.file), 0)
		FROM doc_tab_workouts t
		JOIN doc_workouts dw ON dw.id = t.doc_workout_id
		ORDER BY dw.id, t.day_number
	`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var list []model.DocTabWorkoutInfo
	for rows.Next() {
		var i model.DocTabWorkoutInfo
		if err := rows.Scan(&i.ID, &i.DocWorkoutID, &i.DocWorkoutName, &i.DayNumber, &i.Description, &i.FileName, &i.FileSize); err != nil {
			return nil, err
		}
		list = append(list, i)
	}

	return list, rows.Err()
}

// GetFile возвращает файл тренировки, грузим только когда его надо отправить
func (r *DocTabWorkoutRepo) GetFile(ctx context.Context, id int64) ([]byte, error) {
	var file []byte
	err := r.db.QueryRow(ctx, `
		SELECT file FROM doc_tab_workouts WHERE id = $1
	`, id).Scan(&file)

	return file, err
}

func (r *DocTabWorkoutRepo) SetTgFileID(ctx context.Context, id int64, tgFileID string) error {
	_, err := r.db.Exec(ctx, `
		UPDATE doc_tab_workouts SET tg_file_id = $2 WHERE id = $1
	`, id, tgFileID)

	return err
}
