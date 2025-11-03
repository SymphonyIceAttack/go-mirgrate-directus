package main

import (
	"fmt"
	"log"
	"os"
	"strconv"

	"github.com/joho/godotenv"

	gomigratedirectus "github.com/SymphonyIceAttack/go-mirgrate-directus/go-mirgrate-directus"
)

func main() {
	err := godotenv.Load()
	if err != nil {
		log.Fatal("Error loading .env file")
	}

	baseURL := os.Getenv("BASE_URL")
	baseToken := os.Getenv("BASE_TOKEN")
	targetURL := os.Getenv("TARGET_URL")
	targetToken := os.Getenv("TARGET_TOKEN")
	force, err := strconv.ParseBool(os.Getenv("FORCE"))
	if err != nil {
		log.Fatal("Error parsing FORCE from .env file")
	}

	if err := gomigratedirectus.Migrate(baseURL, baseToken, targetURL, targetToken, force); err != nil {
		fmt.Fprintf(os.Stderr, "Migration failed: %v\n", err)
		os.Exit(1)
	}
}
