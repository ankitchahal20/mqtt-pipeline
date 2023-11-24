package service

import (
	"encoding/json"
	"fmt"
	"net/http"
	"runtime"
	"time"

	"github.com/dgrijalva/jwt-go"
	"github.com/gin-gonic/gin"
	"github.com/gin-gonic/gin/binding"
	"github.com/go-redis/redis/v7"
	"github.com/mqtt-pipeline/internal/config"
	"github.com/mqtt-pipeline/internal/constants"
	"github.com/mqtt-pipeline/internal/models"
	"github.com/mqtt-pipeline/internal/mqtterror"
	"github.com/mqtt-pipeline/internal/utils"
)

var (
	mqttPipelineClient *MQTTPipelineService
)

type MQTTPipelineService struct {
	redisClient *redis.Client
}

func NewMQTTPipelineService(redisClient *redis.Client) {
	mqttPipelineClient = &MQTTPipelineService{
		redisClient: redisClient,
	}
}

func GenerateToken() func(ctx *gin.Context) {
	return func(context *gin.Context) {
		var emailInfo models.Email
		if err := context.ShouldBindBodyWith(&emailInfo, binding.JSON); err == nil {
			txid := context.Request.Header.Get(constants.TransactionID)
			utils.Logger.Info(fmt.Sprintf("received request for generating token, txid : %v", txid))

			token, err := mqttPipelineClient.generateToken(context, emailInfo)
			if err != nil {
				utils.Logger.Error("unable to generate a token for the given email")
				context.Writer.WriteHeader(err.Code)
			} else {
				context.JSON(http.StatusOK, map[string]string{
					"token": token,
				})
			}
		} else {
			context.JSON(http.StatusBadRequest, gin.H{"Unable to marshal the request body": err.Error()})
		}
	}
}

func (service *MQTTPipelineService) generateToken(ctx *gin.Context, emailInfo models.Email) (string, *mqtterror.MQTTPipelineError) {
	token := jwt.New(jwt.SigningMethodHS256)
	claims := token.Claims.(jwt.MapClaims)
	claims["email"] = emailInfo.Email
	claims["exp"] = time.Now().Add(time.Minute * 5).Unix()

	tokenString, err := token.SignedString(constants.SecretKey)
	if err != nil {
		return "", &mqtterror.MQTTPipelineError{
			Code:    http.StatusInternalServerError,
			Message: fmt.Sprintf("Unable to generate the token, err %v", err),
			Trace:   ctx.Request.Header.Get(constants.TransactionID),
		}
	}
	return tokenString, nil
}

func Publish() func(ctx *gin.Context) {
	return func(context *gin.Context) {
		var speedInfo models.SpeedData
		if err := context.ShouldBindBodyWith(&speedInfo, binding.JSON); err == nil {
			txid := context.Request.Header.Get(constants.TransactionID)
			utils.Logger.Info(fmt.Sprintf("received request for publish the speed on mqtt, txid : %v", txid))

			err := mqttPipelineClient.publish(context, speedInfo)
			fmt.Println("runtime.NumGoroutine() : ", runtime.NumGoroutine())
			if err != nil {
				utils.Logger.Error("unable to generate a token for the given email")
				context.Writer.WriteHeader(err.Code)
			} else {
				context.JSON(http.StatusOK, map[string]string{
					"message": "Published speed data to MQTT Pipeline",
				})
			}
		} else {
			context.JSON(http.StatusBadRequest, gin.H{"Unable to marshal the request body": err.Error()})
		}
	}
}

func (service *MQTTPipelineService) publish(ctx *gin.Context, speedInfo models.SpeedData) *mqtterror.MQTTPipelineError {
	txid := ctx.Request.Header.Get(constants.TransactionID)
	payload, _ := json.Marshal(speedInfo)
	cfg := config.GetConfig()
	if token := utils.MQTTClient.Publish(cfg.MQTTConfig.Topic, 0, false, payload); token.Wait() && token.Error() != nil {
		utils.Logger.Error(fmt.Sprintf("unable to publish the message on the topic, txid : %v", txid))
		return &mqtterror.MQTTPipelineError{
			Code:    http.StatusInternalServerError,
			Message: fmt.Sprintf("Unable to send the speed data on the topic, err %v", token.Error()),
			Trace:   txid,
		}
	}
	utils.Logger.Info(fmt.Sprintf("succesfully publish the message on the topic, txid : %v", txid))
	go service.subscribeToMQTT(ctx)

	return nil
}

func (service *MQTTPipelineService) subscribeToMQTT(ctx *gin.Context) *mqtterror.MQTTPipelineError {
	//ctxx := context.Background()
	txid := ctx.Request.Header.Get(constants.TransactionID)

	// blocking call
	speed := <-utils.SpeedChannel
	// Save the latest speed data to Redis
	utils.Logger.Info(fmt.Sprintf("data successfully fetched from the topic, txid : %v", txid))
	err := service.storeInRedis(ctx, speed)
	if err != nil {
		utils.Logger.Error(fmt.Sprintf("unable to store data into redis, txid : %v", txid))
		return err
	}
	return nil
}

func (service *MQTTPipelineService) storeInRedis(ctx *gin.Context, speed int) *mqtterror.MQTTPipelineError {
	// Store the speed data in Redis
	txid := ctx.Request.Header.Get(constants.TransactionID)
	data := map[string]interface{}{"speed": speed}
	val, err := json.Marshal(data)
	if err != nil {
		return &mqtterror.MQTTPipelineError{
			Code:    http.StatusInternalServerError,
			Message: fmt.Sprintf("Unable to send the speed data on the topic, err %v", err.Error()),
			Trace:   txid,
		}
	}

	err = service.redisClient.Set("latest_speed_data", val, 0).Err()
	if err != nil {
		return &mqtterror.MQTTPipelineError{
			Code:    http.StatusInternalServerError,
			Message: fmt.Sprintf("Unable to send the speed data on the topic, err %v", err.Error()),
			Trace:   txid,
		}
	}
	utils.Logger.Info(fmt.Sprintf("data stored successfully in redis , txid : %v", txid))
	return nil
}

func GetSpeedData() func(ctx *gin.Context) {
	return func(ctx *gin.Context) {
		txid := ctx.Request.Header.Get(constants.TransactionID)
		utils.Logger.Info(fmt.Sprintf("received request to get the latest value from redis, txid : %v", txid))
		speed, err := mqttPipelineClient.getSpeedData(ctx)
		if err != nil {
			utils.Logger.Error("unable to get the latest speed data")
			ctx.Writer.WriteHeader(err.Code)
		} else {
			if speed == nil {
				utils.Logger.Info(fmt.Sprintf("no speed data exists in redis, txid : %v", txid))
				ctx.JSON(http.StatusOK, map[string]string{
					"latest_speed": "No speed data found in redis",
				})
			} else {
			ctx.JSON(http.StatusOK, map[string]int{
				"latest_speed": *speed,
			})
		}
		}
	}
}

func (service *MQTTPipelineService) getSpeedData(ctx *gin.Context) (*int, *mqtterror.MQTTPipelineError) {
	txid := ctx.Request.Header.Get(constants.TransactionID)
	val, err := service.redisClient.Get("latest_speed_data").Result()
		if err == redis.Nil {
			utils.Logger.Info(fmt.Sprintf("no stored in stored in redis , txid : %v", txid))
			return nil, nil
		}
		if err != nil {
			utils.Logger.Error(fmt.Sprintf("unable to fetch latest speed data from redis , txid : %v", txid))
			return nil, &mqtterror.MQTTPipelineError{
				Code:    http.StatusInternalServerError,
				Message: fmt.Sprintf("Unable to fetch latest speed data from redis, err %v", err.Error()),
				Trace:   txid,
			}
		}

		var data models.SpeedData
		if err := json.Unmarshal([]byte(val), &data); err != nil {
			utils.Logger.Error(fmt.Sprintf("unmarshalling error for redis data , txid : %v", txid))
			return nil ,  &mqtterror.MQTTPipelineError{
				Code:    http.StatusInternalServerError,
				Message: fmt.Sprintf("error while unmarshalling the response from the redis, err %v", err.Error()),
				Trace:   txid,
			}
		}
		return data.Speed, nil
}