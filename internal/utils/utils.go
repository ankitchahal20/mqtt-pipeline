package utils

import (
	"encoding/json"
	"fmt"
	"log"
	"time"

	mqtt "github.com/eclipse/paho.mqtt.golang"
	"github.com/gin-gonic/gin"
	"github.com/go-redis/redis/v7"
	"github.com/mqtt-pipeline/internal/config"
	"github.com/mqtt-pipeline/internal/constants"
	"github.com/mqtt-pipeline/internal/models"
	"github.com/mqtt-pipeline/internal/mqtterror"
	"go.uber.org/zap"
)

var Logger *zap.Logger
var MQTTClient mqtt.Client
var SpeedChannel = make(chan int)

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

func InitMQTT() {
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

func InitMQTTSubscribe() {

	cfg := config.GetConfig()
	if token := MQTTClient.Subscribe(cfg.MQTTConfig.Topic, 0, func(client mqtt.Client, msg mqtt.Message) {
		var speedData models.SpeedData
		if err := json.Unmarshal(msg.Payload(), &speedData); err == nil {
			SpeedChannel <- *speedData.Speed
		}
	}); token.Wait() && token.Error() != nil {
		close(SpeedChannel)
		Logger.Error(fmt.Sprint("Unable to fetch data from the topic", nil))
		Logger.Fatal("unable Unable to fetch data from the topic")
	}
	Logger.Info(fmt.Sprint("successfully to send the fetch data from the topic", nil))
}

func RespondWithError(c *gin.Context, statusCode int, message string) {
	c.AbortWithStatusJSON(statusCode, mqtterror.MQTTPipelineError{
		Trace:   c.Request.Header.Get(constants.TransactionID),
		Code:    statusCode,
		Message: message,
	})
}
