package main

import (
	"github.com/joho/godotenv"
	"github.com/tdmdh/HETFs/cmd"
)

func main() {
	// Load .env file if present (ignored if missing).
	// SESSION_COOKIE and other secrets are read from here.
	_ = godotenv.Load()
	cmd.Execute()
}
