package main

import (
	"github.com/joho/godotenv"
	"github.com/sitetester/info-center/route"
	"log"
	"os"
)

func init() {
	err := godotenv.Load(".env")
	if err != nil {
		panic("Error loading .env file")
	}
}

func main() {
	engine := route.SetupRouter()
	err := engine.Run(":" + getEnvVar("GIN_PORT"))
	if err != nil {
		panic(err)
	}
}

func getEnvVar(key string) string {
	value, exists := os.LookupEnv(key)
	if !exists {
		log.Fatalf("Environment variable (`%s`) not found!", key)
	}
	return value
}
