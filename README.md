# Нутри Таня

Telegram-бот рассылки тренировок и веб-модуль для тренера.

## Запуск в Docker

1. `copy .env.example .env` и заполнить: пароль Postgres, имя и токен DuckDNS, имя сети VPN
   (`docker network ls`, сеть из compose expenseBot).
2. `copy config.example.yaml config.yaml` и заполнить: `web.password`, `bot_key`, пароль в `database.url`.
   Файл должен существовать до запуска, иначе Docker создаст вместо него папку.
3. Запустить стек expenseBot (в нём VPN), затем здесь: `docker compose up -d --build`.
4. Выставить админов в БД, порт `127.0.0.1:5433`, база `workout_db`: `UPDATE users SET is_admin = true WHERE tg_id = ...;`
5. Создать закрытый канал тренера, добавить бота админом, написать пост: ID канала появится в логе
   (`docker compose logs bot`). Вписать его в `channel_id` в `config.yaml` и выполнить `docker compose restart bot`.

Веб-модуль: `https://<имя>.duckdns.org:<внешний порт роутера>`.

## Обновление

Изменился код: `docker compose up -d --build bot`. Изменился только `config.yaml`: `docker compose restart bot`.

## Бэкап

База: `docker compose exec postgres pg_dump -U postgres workout_db > backup.sql`.
Видео хранятся в канале Telegram, в БД их нет.

## Схема БД

`schema.sql` выполняется автоматически при создании пустой базы. Если база уже есть, схему применяют вручную.
