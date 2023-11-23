package service

import (
	"errors"
	"fmt"
	"net/http"
	"net/mail"
	"time"

	"github.com/dgrijalva/jwt-go"
	"github.com/gin-gonic/gin"
	"github.com/gin-gonic/gin/binding"
	"github.com/go-redis/redis/v7"
	"github.com/mqtt-pipeline/internal/constants"
	"github.com/mqtt-pipeline/internal/models"
	"github.com/mqtt-pipeline/internal/mqtterror"
	"github.com/mqtt-pipeline/internal/utils"
)

var mqttPipelineClient *MQTTPipelineService

var secretKey = []byte("SOME-SECRET-KEY-WHICH-NOT_SECRET-ANY-MORE")

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
			Code: http.StatusInternalServerError,
			Message: fmt.Sprintf("Unable to generate the token, err %v", err),
			Trace: ctx.Request.Header.Get(constants.TransactionID),
		}
	}
	return tokenString, nil
}
