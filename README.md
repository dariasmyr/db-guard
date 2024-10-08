![Logo](docs/logo.jpg)
# DB Guard

DB Guard is a Golang project aimed at streamlining database backup procedures with automated scheduling. 

*Disclaimer: This project is designed for PostgreSQL databases only, but it can be extended to support other database management systems.*

### Key Advantages:
- **Automated Backups**: Streamline database backup processes with scheduled intervals.
- **Real-Time Notifications**: Receive instant updates on backup statuses via webhooks.
- **Optimized Storage**: Utilize parallel file archiving to maximize disk space efficiency.

## Prerequisites
- PostgreSQL client (for PostgreSQL databases)

## How to Run
First but not least, clone the lib:
```bash
go get github.com/dariasmyr/db-guard
```

### Option 1: Run via Command Line
You can run the application directly from the command line. 
```bash
go run cmd/db-guard.go \
  --host=<your-db-host> \
  --port=<your-db-port> \
  --user=<your-db-user> \
  --password=<your-db-password> \
  --compress \ ## optional
  --compression-level=<your-compression-level> \  ## optional
  --database=<your-db-name> \
  --max-backup-count=<your-max-backup-count> \
  --interval-seconds=<your-interval-in-seconds> \
  --dir=<your-backups-dir> ## optional
  --webhook-url=<your-webhook-url> ## optional
```


### Option 2: Run via Docker
**In this docker container example we pull ready-made Docker Image (db-guard) from Docker Hub**

**docker-compose.yml**
```bash
db_backup:
      image: dashasmyr/db-guard:latest
      restart: unless-stopped
      environment:
          - HOST=db
          - PORT=${DATABASE_PORT}
          - USER=${DATABASE_USER}
          - PASSWORD=${DATABASE_PASSWORD}
          - DATABASE=${DATABASE_NAME}
          - MAX_BACKUP_COUNT=42 # keep one week
          - INTERVAL_SECONDS=14400 # every 4 hours
          - WEBHOOK_URL=${WEBHOOK_URL} # optional
      volumes:
        - ./data/db_backups:/app/backups:rw
```

You can also run the application using Docker by executing the following command:
```bash
docker compose up -d --build
```
**!! Note that if your database is running outside, you need to create a metwork so that DB Guard could access the docker container with your database.**

### Option 3: Pull ready-made Docker Image from Docker Hub
```bash
docker pull dashasmyr/db-guard
```

Ensure you have provided necessary configurations in a `.env` file. Refer to the `.env-sample` for an example.

## Parameters
- `--host`: Hostname of the database server
- `--port`: Port number of the database server
- `--user`: Username for authentication
- `--password`: Password for authentication
- `--compress`: Enable compression for backup files (Optional)
- `--compression-level`: Level of compression (Default: -1) (Optional)
- `--database`: Name of the database to backup
- `--max-backup-count`: Maximum number of backup files to retain
- `--interval-seconds`: Interval between each backup (in seconds)
- `--dir`: Directory to store backup files (Optional)
- `--webhook-url`: Webhook URL for notifications (Optional)

## Compression Levels
You can specify the compression level using the following options:
- NoCompression (0)
- BestSpeed (1)
- BestCompression (9)
- DefaultCompression (-1) (Default)
- HuffmanOnly (-2)

## Show your support
Give a ⭐️ if this project helped you!

## License
Copyright © 2023 [dasha.smyr@gmail.com](https://github.com/dariasmyr).<br />
This project is [MIT](LICENSE) licensed.
