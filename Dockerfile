FROM golang:latest AS build
RUN apt-get update && \
    apt-get install -y postgresql-client make && \
    apt-get clean && \
    rm -rf /var/lib/apt/lists/*
ENV CGO_ENABLED=0
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN make build

FROM alpine:latest
RUN apk update && apk add --no-cache postgresql-client
WORKDIR /app
COPY --from=build /app/build/db-guard /app/db-guard
CMD ["./db-guard"]
