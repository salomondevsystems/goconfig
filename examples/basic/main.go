package main

import (
	"fmt"
	"log"
	"os"

	"github.com/salomondevsystems/goconfig"
)

func main() {
	cfg := goconfig.New(goconfig.WithConfigDir("."))
	if err := cfg.LoadEnv(".env"); err != nil {
		log.Printf("warning: could not load .env file: %v", err)
	}

	// Parsea el archivo de configuración
	var appConfig Config
	if err := cfg.Parse("config", &appConfig); err != nil {
		log.Fatalf("failed to parse config: %v", err)
	}

	fmt.Printf("Server: %s (%s)\n", appConfig.App.Name, appConfig.App.Environment)
	fmt.Printf("Listening on port %d\n", appConfig.App.Port)
	fmt.Printf("Database: %s:%d\n", appConfig.Database.Host, appConfig.Database.Port)

	os.Exit(0)
}
