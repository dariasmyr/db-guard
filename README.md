<div style="text-align: center;">
    <img src="docs/logo.jpg"
         alt="Logo"
         style="width: 50%; height: auto;" />
</div>
# DB Dump

DB Dump is a Golang project designed to automate the backup process of databases at specified intervals. It includes a Telegram notification feature to provide status updates for each backup and utilizes parallel file archiving to save disk space.

Currently, the project supports backups of PostgreSQL databases using `pg_dump`.

## Prerequisites
- PostgreSQL client (including `pg_dump`)
- Telegram bot API token - Create a bot using [BotFather](https://t.me/BotFather) and obtain the token in the format `123456789:ABC-DEF1234ghIkl-zyx57W2v1u123ew11`
- Chat ID for receiving messages - You can find this out using [chat_id_echo_bot](https://t.me/chat_id_echo_bot)

## How to Run
```bash
TELEGRAM_BOT_TOKEN=<Your Telegram Bot Token> CHANNEL_ID=<Your Chat ID> go run cmd/db-dump.go --host=localhost --port=5433 --user=postgres --password=postgres --compress --database=betting --max-backup-count=5 --interval-seconds=20 --dir=backups --telegram-notifications
```

## Parameters
- `--host`: Hostname of the database server
- `--port`: Port number of the database server
- `--user`: Username for authentication
- `--password`: Password for authentication
- `--compress`: Enable compression for backup files
- `--database`: Name of the database to backup
- `--max-backup-count`: Maximum number of backup files to retain
- `--interval-seconds`: Interval between each backup (in seconds)
- `--dir`: Directory to store backup files
- `--telegram-notifications`: Enable Telegram notifications

## Example
```bash
TELEGRAM_BOT_TOKEN=123456789:ABC-DEF1234ghIkl-zyx57W2v1u123ew11 CHANNEL_ID=11123456789 go run cmd/db-dump.go --host=localhost --port=5432 --user=postgres --password=postgres --compress --database=betting --max-backup-count=5 --interval-seconds=86400 --dir=backups --telegram-notifications
```


## Show your support

Give a ⭐️ if this project helped you!

## License

Copyright © 2023 [dasha.smyr@gmail.com](https://github.com/dariasmyr).<br />
This project is [MIT](LICENSE) licensed.