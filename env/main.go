package main

import (
	"fmt"
	"log"
	"os"

	"github.com/joho/godotenv"
)

func goDotEnvVariable(key string) string {
	err := godotenv.Load()

	if err != nil {
		log.Fatalln("Error loading .env file", err)
	}

	return os.Getenv(key)
}

func main() {
	dotenv := goDotEnvVariable("STRONGEST_AVEENGER")

	fmt.Println(dotenv)
}
