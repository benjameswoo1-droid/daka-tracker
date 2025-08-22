package main

import (
	"github.com/benjameswoo1-droid/daka-tracker/cmd"
	"github.com/joho/godotenv"
	"log"
)

func main() {
	if err := godotenv.Load(); err != nil {
		log.Println("No .env file found (loading environment variables from the system instead)")
	}

	cmd.Execute()
}
