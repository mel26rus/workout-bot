-- Нутри Таня: инициализация БД (Postgres)
-- Префикс doc_ у заголовочной части, doc_tab_ у табличной части.
-- Скрипт можно запускать повторно.

BEGIN;

-- Автообновление updated_at
CREATE OR REPLACE FUNCTION set_updated_at() RETURNS trigger AS $$
BEGIN
    NEW.updated_at = now();
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

-- Челлендж (заголовок)
CREATE TABLE IF NOT EXISTS doc_workouts (
    id          bigserial   PRIMARY KEY,
    name        text        NOT NULL,
    description text        NOT NULL DEFAULT '',
    payment_url text        NOT NULL DEFAULT '',
    created_at  timestamptz NOT NULL DEFAULT now(),
    updated_at  timestamptz NOT NULL DEFAULT now()
);

-- Пользователи
CREATE TABLE IF NOT EXISTS users (
    id             bigserial   PRIMARY KEY,
    tg_id          bigint      NOT NULL UNIQUE,
    username       text,
    nickname       text,
    first_name     text        NOT NULL DEFAULT '',
    surname        text,
    language_code  text,
    is_admin       boolean     NOT NULL DEFAULT false,
    doc_workout_id bigint      REFERENCES doc_workouts (id) ON DELETE SET NULL, -- выбранный сейчас челлендж
    created_at     timestamptz NOT NULL DEFAULT now(),
    updated_at     timestamptz NOT NULL DEFAULT now()
);

-- Посты закрытого канала тренера: бот копирует их людям как тренировки
CREATE TABLE IF NOT EXISTS channel_posts (
    id          bigserial   PRIMARY KEY,
    chat_id     bigint      NOT NULL,
    message_id  integer     NOT NULL,
    text        text        NOT NULL DEFAULT '',     -- подпись к файлу или текст поста
    media_type  text        NOT NULL DEFAULT 'text', -- video, document, photo, text, other
    file_name   text,
    file_size   bigint,
    post_date   timestamptz NOT NULL,
    created_at  timestamptz NOT NULL DEFAULT now(),
    updated_at  timestamptz NOT NULL DEFAULT now(),
    UNIQUE (chat_id, message_id)
);

-- Тренировки челленджа (табличная часть): какой пост канала отправлять в какой день
CREATE TABLE IF NOT EXISTS doc_tab_workouts (
    id              bigserial   PRIMARY KEY,
    doc_workout_id  bigint      NOT NULL REFERENCES doc_workouts (id),
    day_number      int         NOT NULL CHECK (day_number > 0),
    channel_post_id bigint      NOT NULL UNIQUE REFERENCES channel_posts (id),
    created_at      timestamptz NOT NULL DEFAULT now(),
    updated_at      timestamptz NOT NULL DEFAULT now(),
    UNIQUE (doc_workout_id, day_number)
);

-- Участие человека в челлендже (заголовок), создаётся при подтверждении оплаты
CREATE TABLE IF NOT EXISTS doc_user_workouts (
    id             bigserial   PRIMARY KEY,
    user_id        bigint      NOT NULL REFERENCES users (id),
    doc_workout_id bigint      NOT NULL REFERENCES doc_workouts (id),
    paid_at        timestamptz NOT NULL DEFAULT now(),
    start_date     date        NOT NULL, -- следующий день после подтверждения оплаты (МСК)
    created_at     timestamptz NOT NULL DEFAULT now(),
    updated_at     timestamptz NOT NULL DEFAULT now(),
    UNIQUE (user_id, doc_workout_id)
);

-- Расписание и ход выполнения (табличная часть)
CREATE TABLE IF NOT EXISTS doc_tab_user_workouts (
    id                  bigserial   PRIMARY KEY,
    doc_user_workout_id bigint      NOT NULL REFERENCES doc_user_workouts (id),
    doc_tab_workout_id  bigint      NOT NULL REFERENCES doc_tab_workouts (id),
    send_date           date        NOT NULL, -- start_date + day_number - 1
    sent_at             timestamptz,          -- NULL: ещё не отправлено
    done_at             timestamptz,          -- NULL: не выполнено
    created_at          timestamptz NOT NULL DEFAULT now(),
    updated_at          timestamptz NOT NULL DEFAULT now(),
    UNIQUE (doc_user_workout_id, doc_tab_workout_id)
);

-- Настройки, которые тренер меняет в веб-модуле (send_time - время рассылки ЧЧ:ММ)
CREATE TABLE IF NOT EXISTS settings (
    key        text        PRIMARY KEY,
    value      text        NOT NULL,
    created_at timestamptz NOT NULL DEFAULT now(),
    updated_at timestamptz NOT NULL DEFAULT now()
);

-- Выборка шедулера: send_date = сегодня AND sent_at IS NULL
CREATE INDEX IF NOT EXISTS idx_doc_tab_user_workouts_unsent
    ON doc_tab_user_workouts (send_date)
    WHERE sent_at IS NULL;

-- Триггеры updated_at
DROP TRIGGER IF EXISTS trg_doc_workouts_updated_at ON doc_workouts;
CREATE TRIGGER trg_doc_workouts_updated_at
    BEFORE UPDATE ON doc_workouts FOR EACH ROW EXECUTE FUNCTION set_updated_at();

DROP TRIGGER IF EXISTS trg_users_updated_at ON users;
CREATE TRIGGER trg_users_updated_at
    BEFORE UPDATE ON users FOR EACH ROW EXECUTE FUNCTION set_updated_at();

DROP TRIGGER IF EXISTS trg_channel_posts_updated_at ON channel_posts;
CREATE TRIGGER trg_channel_posts_updated_at
    BEFORE UPDATE ON channel_posts FOR EACH ROW EXECUTE FUNCTION set_updated_at();

DROP TRIGGER IF EXISTS trg_settings_updated_at ON settings;
CREATE TRIGGER trg_settings_updated_at
    BEFORE UPDATE ON settings FOR EACH ROW EXECUTE FUNCTION set_updated_at();

DROP TRIGGER IF EXISTS trg_doc_tab_workouts_updated_at ON doc_tab_workouts;
CREATE TRIGGER trg_doc_tab_workouts_updated_at
    BEFORE UPDATE ON doc_tab_workouts FOR EACH ROW EXECUTE FUNCTION set_updated_at();

DROP TRIGGER IF EXISTS trg_doc_user_workouts_updated_at ON doc_user_workouts;
CREATE TRIGGER trg_doc_user_workouts_updated_at
    BEFORE UPDATE ON doc_user_workouts FOR EACH ROW EXECUTE FUNCTION set_updated_at();

DROP TRIGGER IF EXISTS trg_doc_tab_user_workouts_updated_at ON doc_tab_user_workouts;
CREATE TRIGGER trg_doc_tab_user_workouts_updated_at
    BEFORE UPDATE ON doc_tab_user_workouts FOR EACH ROW EXECUTE FUNCTION set_updated_at();

COMMIT;
