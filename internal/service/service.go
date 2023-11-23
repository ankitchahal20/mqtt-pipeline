package service

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/mail"
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
	secretKey          = []byte("SOME-SECRET-KEY-WHICH-NOT_SECRET-ANY-MORE")
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
			// Validate request body
			if emailInfo.Email == "" {
				err := errors.New("invalid request received")
				context.JSON(http.StatusBadRequest, gin.H{"email not found": err.Error()})
				return
			}

			_, parseErr := mail.ParseAddress(emailInfo.Email)
			if parseErr != nil {
				context.JSON(http.StatusBadRequest, gin.H{"invalid email found": parseErr.Error()})
				return
			}

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

	tokenString, err := token.SignedString(secretKey)
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
			// Validate request body
			if speedInfo.Speed == nil {
				err := errors.New("invalid request received")
				context.JSON(http.StatusBadRequest, gin.H{"email not found": err.Error()})
				return
			}

			if *speedInfo.Speed < 0 || *speedInfo.Speed > 100 {
				err := errors.New("speed should be range between 0 and 100")
				context.JSON(http.StatusBadRequest, gin.H{"invalid speed": err.Error()})
				return
			}

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