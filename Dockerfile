# Сборка
FROM golang:1.24-alpine AS builder
WORKDIR /app
COPY . .
RUN go mod download
RUN CGO_ENABLED=0 GOOS=linux go build -o /loadbalancer ./cmd/main.go
# Финальный образ
FROM alpine:latest
COPY --from=builder /loadbalancer /loadbalancer
# Создаем директорию и копируем конфиг
RUN mkdir -p /configs
COPY ./configs/config.json /configs/config.json
EXPOSE 8888
CMD ["/loadbalancer"]