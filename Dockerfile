# Этап 1: Сборка приложения
# Используем ту же версию Go, что и в go.mod
FROM golang:1.25-alpine AS builder

WORKDIR /app

# Копируем файлы управления зависимостями.
COPY go.mod go.sum ./

# Загружаем зависимости.
RUN go mod download

# Копируем остальной исходный код.
COPY . .

# Собираем исполняемый файл.
RUN CGO_ENABLED=0 GOOS=linux go build -o /app/main ./cmd/app

# Этап 2: Создание минимального образа для запуска
FROM alpine:latest

WORKDIR /app

# Копируем только исполняемый файл.
COPY --from=builder /app/main .

# Копируем миграции.
COPY migrations ./migrations

EXPOSE 8080
CMD ["/app/main"]