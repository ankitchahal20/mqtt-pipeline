package utils

import (
	"fmt"
	"log"
	"time"

	mqtt "github.com/eclipse/paho.mqtt.golang"
	"github.com/go-redis/redis/v7"
	"github.com/mqtt-pipeline/internal/config"
	"go.uber.org/zap"
)

var Logger *zap.Logger
var MQTTClient mqtt.Client

func InitRedis() *redis.Client {
	cfg := config.GetConfig()
	client := redis.NewClient(&redis.Options{
		Addr:        cfg.RedisConfig.URL,
		Password:    "",
		DB:          cfg.RedisConfig.DBNum,
		IdleTimeout: time.Duration(cfg.RedisConfig.IdleTimeout),
	})

	return client
}

func InitLogClient() {
	Logger, _ = zap.NewDevelopment()
}

func InitMQTT(){
	cfg := config.GetConfig()
	var mqttBroker string = cfg.MQTTConfig.MQTTBroker
	if mqttBroker == "" {
		mqttBroker = "tcp://broker.emqx.io:1883"
	}

	opts := mqtt.NewClientOptions().AddBroker(mqttBroker)
	opts.SetClientID("go-app")
	opts.SetCleanSession(true)
	MQTTClient = mqtt.NewClient(opts)
	if token := MQTTClient.Connect(); token.Wait() && token.Error() != nil {
		fmt.Println("Failed to connect to MQTT broker")
		fmt.Println("MQTT Broker Address:", mqttBroker)
		fmt.Println("Error:", token.Error())
		log.Fatal(token.Error())
	}
}