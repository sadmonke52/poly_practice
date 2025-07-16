package main

import (
	"log"
	"os"
	"processor/config"
	"processor/internal/app"
)

// CONFIG_PATH=D:\GOlangProject\poly_practice_2\collector\config\local.yaml
func main() {
	//configPath := "config/local.yaml"
	configPath := os.Getenv("CONFIG_PATH")
	if configPath == "" {
		configPath = "/etc/myapp/local.yaml"
	}

	cfg, err := config.Load(configPath)
	if err != nil {
		log.Fatal(err)
	}

	app.MustRun(cfg)
}
