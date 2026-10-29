# ===========================
# Stage 1 - Build
# ===========================
FROM golang:1.25 AS builder

WORKDIR /src

# Сначала зависимости (для кеширования)
COPY go.mod go.sum ./
RUN go mod download

# Исходники
COPY . .

# Собираем статический бинарник
RUN CGO_ENABLED=0 \
    GOOS=linux \
    GOARCH=amd64 \
    go build \
    -ldflags="-s -w" \
    -o workout-bot \
    ./cmd/bot


# ===========================
# Stage 2 - Runtime
# ===========================
FROM debian:bookworm-slim

WORKDIR /app

# SSL сертификаты для HTTPS
RUN apt-get update && \
    apt-get install -y ca-certificates && \
    rm -rf /var/lib/apt/lists/*

COPY --from=builder /src/workout-bot .

# Каталог для логов
RUN mkdir Logs

EXPOSE 8080

ENTRYPOINT ["./workout-bot"]
