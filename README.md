# Нутри Таня

Telegram-бот рассылки тренировок и веб-модуль для тренера.

## Запуск в Docker

Postgres стоит на хосте (база `workout_db` уже создана по `schema.sql`), в Docker его нет.
Postgres должен принимать подключения из Docker: `listen_addresses` в `postgresql.conf`,
запись для сети Docker в `pg_hba.conf`, порт 5432 открыт в брандмауэре Windows.

1. `copy .env.example .env` и заполнить: имя и токен DuckDNS, имя сети VPN
   (`docker network ls`, сеть из compose expenseBot).
2. `copy config.example.yaml config.yaml` и заполнить: `web.password`, `bot_key`, пароль в `database.url`.
   Файл должен существовать до запуска, иначе Docker создаст вместо него папку.
3. Запустить стек expenseBot (в нём VPN), затем здесь: `docker compose up -d --build`.
4. Админы выставляются в БД: `UPDATE users SET is_admin = true WHERE tg_id = ...;`
5. Создать закрытый канал тренера, добавить бота админом, написать пост: ID канала появится в логе
   (`docker compose logs bot`). Вписать его в `channel_id` в `config.yaml` и выполнить `docker compose restart bot`.

Веб-модуль: `https://<имя>.duckdns.org:<внешний порт роутера>`.

## Обновление

Изменился код: `docker compose up -d --build bot`. Изменился только `config.yaml`: `docker compose restart bot`.

## Бэкап

База: `pg_dump -U postgres workout_db > backup.sql` на хосте.
Видео хранятся в канале Telegram, в БД их нет.

## Схема БД

`schema.sql` применяется к базе вручную.
