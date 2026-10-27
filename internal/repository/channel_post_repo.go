package repository

import (
	"context"

	"workout-bot/internal/model"

	"github.com/jackc/pgx/v5/pgxpool"
)

type ChannelPostRepo struct {
	db *pgxpool.Pool
}

func NewChannelPostRepo(db *pgxpool.Pool) *ChannelPostRepo {
	return &ChannelPostRepo{db: db}
}

// Save добавляет новый пост канала или обновляет отредактированный
func (r *ChannelPostRepo) Save(ctx context.Context, p model.ChannelPost) error {
	_, err := r.db.Exec(ctx, `
		INSERT INTO channel_posts (chat_id, message_id, text, media_type, file_name, file_size, post_date)
		VALUES ($1, $2, $3, $4, NULLIF($5, ''), NULLIF($6::bigint, 0), $7)
		ON CONFLICT (chat_id, message_id) DO UPDATE SET
			text       = EXCLUDED.text,
			media_type = EXCLUDED.media_type,
			file_name  = EXCLUDED.file_name,
			file_size  = EXCLUDED.file_size
	`, p.ChatID, p.MessageID, p.Text, p.MediaType, p.FileName, p.FileSize, p.PostDate)

	return err
}

// GetAllInfo возвращает все посты: сначала не привязанные (новые сверху), потом привязанные по челленджам и дням
func (r *ChannelPostRepo) GetAllInfo(ctx context.Context) ([]model.ChannelPostInfo, error) {
	rows, err := r.db.Query(ctx, `
		SELECT p.id, p.chat_id, p.message_id, p.text, p.media_type,
		       COALESCE(p.file_name, ''), COALESCE(p.file_size, 0), p.post_date,
		       COALESCE(t.doc_workout_id, 0), COALESCE(t.day_number, 0)
		FROM channel_posts p
		LEFT JOIN doc_tab_workouts t ON t.channel_post_id = p.id
		ORDER BY (t.id IS NOT NULL), t.doc_workout_id, t.day_number, p.post_date DESC
	`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var list []model.ChannelPostInfo
	for rows.Next() {
		var i model.ChannelPostInfo
		err := rows.Scan(
			&i.ID, &i.ChatID, &i.MessageID, &i.Text, &i.MediaType,
			&i.FileName, &i.FileSize, &i.PostDate,
			&i.DocWorkoutID, &i.DayNumber,
		)
		if err != nil {
			return nil, err
		}
		list = append(list, i)
	}

	return list, rows.Err()
}
