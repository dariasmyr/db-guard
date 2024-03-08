FROM golang:alpine AS build
RUN apk update && apk add --no-cache postgresql-client
ENV CGO_ENABLED=0
WORKDIR /app
COPY . .
RUN go build -o db-guard -ldflags="-s -w" cmd/db-guard.go

FROM alpine:latest
RUN apk update && apk add --no-cache postgresql-client
WORKDIR /app
COPY --from=build /app/db-guard /app/db-guard
CMD ["./db-guard"]
