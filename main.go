package main

import (
	"log"
	"os"
	"strconv"
	"time"

	"github.com/marcuscarr/appts/server"
)

func main() {
	host := os.Getenv("HOST")
	port, err := strconv.Atoi(os.Getenv("PORT"))
	if err != nil {
		log.Fatal(err)
	}

	dbHost := os.Getenv("DB_HOST")
	dbPort, err := strconv.Atoi(os.Getenv("DB_PORT"))
	if err != nil {
		log.Fatal(err)
	}

	dbUser := os.Getenv("DB_USER")
	dbPass := os.Getenv("DB_PASS")
	dbName := os.Getenv("DB_NAME")

	config := server.Config{
		Host:    host,
		Port:    port,
		DBHost:  dbHost,
		DBPort:  dbPort,
		DBUser:  dbUser,
		DBName:  dbName,
		DBPass:  dbPass,
		Timeout: time.Duration(5) * time.Second,
	}

	server := server.New(&config)

	server.Run()
}
