package main

import (
	"filemanager/server"
	"log"
	"os"

	"github.com/joho/godotenv"
)

func main() {
	if err := godotenv.Load("./.env"); err != nil {
		if _, envExist := os.LookupEnv("ENV_LOADED"); !envExist {
			log.Fatal("environment file not found")
		}
	}

	server.RunServer()
}
