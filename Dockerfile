# Build stage
FROM golang:1.21-alpine AS builder

# Устанавливаем зависимости для сборки
RUN apk add --no-cache git ca-certificates tzdata

# Создаем рабочую директорию
WORKDIR /app

# Копируем go.mod и go.sum
COPY go.mod go.sum ./

# Скачиваем зависимости
RUN go mod download

# Копируем исходный код
COPY . .

# Собираем приложение с оптимизациями
RUN CGO_ENABLED=0 GOOS=linux go build \
    -ldflags="-w -s -X main.Version=1.0.0" \
    -o /app/iot-metrics-service \
    ./cmd/server/main.go

# Final stage - минимальный образ
FROM alpine:3.18

# Устанавливаем необходимые пакеты
RUN apk add --no-cache ca-certificates tzdata

# Создаем не-root пользователя для безопасности
RUN addgroup -g 1000 -S appgroup && \
    adduser -u 1000 -S appuser -G appgroup

WORKDIR /app

# Копируем бинарник из builder stage
COPY --from=builder --chown=appuser:appgroup /app/iot-metrics-service /app/

# Переключаемся на не-root пользователя
USER appuser

# Открываем порт
EXPOSE 8080

# Здоровье-чек
HEALTHCHECK --interval=30s --timeout=3s --start-period=5s --retries=3 \
    CMD wget --no-verbose --tries=1 --spider http://localhost:8080/api/health || exit 1

# Запускаем приложение
ENTRYPOINT ["/app/iot-metrics-service"]