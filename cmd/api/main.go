package main

import (
	"fmt"
	"go-url-shortener/cfg"
	"go-url-shortener/internal/storage"
	"go-url-shortener/internal/storage/inmemory"
	"go-url-shortener/internal/storage/postgres"
	shortenerproto "go-url-shortener/pkg/proto"
	"go-url-shortener/shortener"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
	"log"
	"net"
	"os"
)

func main() {
	env, err := cfg.ReadCFG()
	if err != nil {
		log.Fatal(err)
	}

	listener, _ := net.Listen("tcp", ":"+env.ServerPort)
	s := grpc.NewServer()

	// Creating storage depending on startup preference. In-memory is default.
	var keystorage storage.UrlStorage
	var shortener *shortenerserver.LinkShortener

	// Setting up our storage. It depends on cmd arguments ('command' field in docker-compose)
	if len(os.Args) == 1 || os.Args[1] == "in-memory" {
		keystorage = inmemory.New()
		shortener = shortenerserver.New(keystorage, env)
		log.Println("Server will run with in-memory storage")
	} else if os.Args[1] == "postgres" {
		connectionString := fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=disable", env.DBHost, env.DBPort, env.DBUser, env.DBPass, env.DBName)
		keystorage = postgres.New(connectionString)
		shortener = shortenerserver.New(keystorage, env)
		log.Println("Server will run with PostgreSQL storage")
	}

	defer func(keystorage storage.UrlStorage) {
		err = keystorage.Close()
		if err != nil {
			log.Fatal(err)
		}
	}(keystorage)

	// In order to see endpoints in Postman :)
	reflection.Register(s)

	shortenerproto.RegisterLinkShortenerServer(s, shortener)
	log.Println("Started gRPC server on localhost:", env.ServerPort)
	err = s.Serve(listener)
	if err != nil {
		log.Fatal(err)
	}
}
