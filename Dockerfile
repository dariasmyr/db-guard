FROM golang:latest

# Install PostgreSQL client
RUN apt-get update && apt-get install -y postgresql-client

WORKDIR /app
COPY . .
RUN go build -o db-guard cmd/db-guard.go

CMD ["./db-guard"]
