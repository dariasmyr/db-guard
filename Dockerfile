FROM golang:alpine AS build
RUN apk update && apk add --no-cache postgresql-client make
ENV CGO_ENABLED=0
WORKDIR /app
COPY . .
RUN make build

FROM alpine:latest
RUN apk update && apk add --no-cache postgresql-client
WORKDIR /app
COPY --from=build /app/build/db-guard /app/db-guard
CMD ["./db-guard"]
