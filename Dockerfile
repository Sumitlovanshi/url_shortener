# syntax=docker/dockerfile:1

FROM golang:1.23-alpine AS build

WORKDIR /src

COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -trimpath -ldflags="-s -w" -o /bin/url-shortener ./cmd/server

FROM alpine:3.21

RUN addgroup -S app && adduser -S app -G app

WORKDIR /app
RUN mkdir -p /app/data && chown -R app:app /app

COPY --from=build /bin/url-shortener /usr/local/bin/url-shortener

USER app

EXPOSE 8080

ENV PORT=8080
ENV BASE_URL=http://localhost:8080
ENV DATA_PATH=/app/data/url_shortener.db

ENTRYPOINT ["url-shortener"]
