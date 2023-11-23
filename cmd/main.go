package main

import (
	"log"

	"github.com/mqtt-pipeline/internal/config"
	"github.com/mqtt-pipeline/internal/server"
	"github.com/mqtt-pipeline/internal/service"
	"github.com/mqtt-pipeline/internal/utils"
)

func main() {
	// Initializing the Log client
	utils.InitLogClient()

	// Initializing the GlobalConfig
	err := config.InitGlobalConfig()
	if err != nil {
		log.Fatalf("Unable to initialize global config")
	}

	utils.Logger.Info("main started")

	// Initialize Redis
	redisClient := utils.InitRedis()
	// create client
	service.NewMQTTPipelineService(redisClient)
	server.Start()
}
