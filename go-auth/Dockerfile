# Сборка бинарника
FROM golang:1.24 AS build
WORKDIR /app

# Копируем необходимые файлы для компиляции
COPY go-auth/go.mod go-auth/go.sum ./
RUN go mod download

# Копируем proto-файлы и внутренние файлы go-auth
COPY api/ /api/
COPY go-auth/ ./

# Устанавливаем protoc и необходимые плагины
RUN apt-get update && apt-get install -y protobuf-compiler
RUN go install google.golang.org/protobuf/cmd/protoc-gen-go@latest
RUN go install google.golang.org/grpc/cmd/protoc-gen-go-grpc@latest

# Создаем директорию для сгенерированных файлов, если она еще не существует
RUN mkdir -p internal/generated

# Генерируем protobuf файлы
RUN protoc --proto_path=/api \
    --go_out=internal/generated --go_opt=paths=source_relative \
    --go-grpc_out=internal/generated --go-grpc_opt=paths=source_relative \
    auth.proto

# СТАТИЧЕСКАЯ СБОРКА: отключаем CGO и задаем static linking
ENV CGO_ENABLED=0
# Собираем основной сервер
RUN go build -ldflags="-s -w" -o auth_server ./cmd/server
# Собираем утилиту администратора
RUN go build -ldflags="-s -w" -o addadmin ./internal/addAdmin

# Минималистичный рантайм с musl libc
FROM alpine:latest
RUN apk --no-cache add ca-certificates tzdata

WORKDIR /app
COPY --from=build /app/auth_server .
COPY --from=build /app/addadmin .

# Создаем команду add
RUN echo '#!/bin/sh' > /usr/local/bin/add && \
    echo '/app/addadmin' >> /usr/local/bin/add && \
    chmod +x /usr/local/bin/add

EXPOSE 50051
ENTRYPOINT ["/app/auth_server"]