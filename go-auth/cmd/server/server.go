package main

import (
	"Auth/internal/db"
	pb "Auth/internal/generated"
	"Auth/internal/service"
	"context"
	"time"

	// iteraptor "Auth/internal/interaptors"
	// "Auth/internal/server"

	"net"
	"os"

	log "github.com/sirupsen/logrus"
	"google.golang.org/grpc"
)

func init() {
	log.SetLevel(log.TraceLevel)
}

func handler(address string) {
	lis, err := net.Listen("tcp", address)
	if err != nil {
		log.Fatalf("Ошибка запуска слушателя: %v", err)
	}

	// Инициализация БД
	ctx, cancel := context.WithTimeout(context.Background(), 20*time.Second)
	defer cancel()
	err = db.InitDB(ctx)
	if err != nil {
		log.Fatalf("Ошибка инициализациии базы данных: %v", err)
	}
	// Создаем дефолтного админа
	user := db.User{
		Login:    "Terwin",
		Password: "BEST",
		Role:     "ADMIN",
	}
	_, err = db.AddUser(ctx, user)
	if err != nil {
		log.Fatalf("Ошибка создания админа: %v", err)
	}
	// Создание grpc-сервера
	grpcServer := grpc.NewServer()

	// Регистрация UserService сервера
	pb.RegisterAuthenticationServer(grpcServer, &service.UserService{})

	log.Println("gRPC-сервер запущен по адрессу: ", address)
	// Запускаем обслуживание запросов (блокирует текущий поток)
	if err := grpcServer.Serve(lis); err != nil {
		log.Fatalf("Ошибка работы gRPC-сервера: %v", err)
	}
}

func main() {
	address := os.Getenv("AUTH_SERVER_ADDRESS")
	if address == "" {
		address = ":50051"
	}
	handler(address)

}
