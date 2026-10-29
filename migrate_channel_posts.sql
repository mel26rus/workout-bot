-- Миграция тестовой БД на схему с каналом.
-- Порядок: сначала выполнить schema.sql (создаст channel_posts и settings), потом этот файл.
-- Старые тренировки с видео в БД в новую схему не переносятся:
-- удаляются тренировки, расписание людей и подтверждения оплаты (после этого оплату можно подтвердить заново).

BEGIN;

DELETE FROM doc_tab_user_workouts;
DELETE FROM doc_user_workouts;
DELETE FROM doc_tab_workouts;

ALTER TABLE doc_tab_workouts
    DROP COLUMN description,
    DROP COLUMN file,
    DROP COLUMN file_name,
    DROP COLUMN mime_type,
    DROP COLUMN tg_file_id,
    ADD COLUMN channel_post_id bigint NOT NULL UNIQUE REFERENCES channel_posts (id);

COMMIT;
