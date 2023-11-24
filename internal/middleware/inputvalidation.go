package middleware

import (
	"errors"
	"fmt"
	"net/http"
	"net/mail"

	"github.com/gin-gonic/gin"
	"github.com/gin-gonic/gin/binding"
	"github.com/google/uuid"
	"github.com/mqtt-pipeline/internal/constants"
	"github.com/mqtt-pipeline/internal/models"
	"github.com/mqtt-pipeline/internal/utils"
)

// This function gets the unique transactionID
func getTransactionID(c *gin.Context) string {
	transactionID := c.GetHeader(constants.TransactionID)
	_, err := uuid.Parse(transactionID)
	if err != nil {
		transactionID = uuid.New().String()
		c.Request.Header.Set(constants.TransactionID, transactionID)
	}
	return transactionID
}

// Set the transactionID from headers if not present create a new.
func SetTransactionId() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		_ = getTransactionID(ctx)
	}
}

func ValidatePublishEndpointRequest() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		txid := ctx.Request.Header.Get(constants.TransactionID)

		var speedData models.SpeedData
		err := ctx.ShouldBindBodyWith(&speedData, binding.JSON)
		if err != nil {
			utils.Logger.Error(fmt.Sprintf("error while unmarshaling the request field for create account data validation, txid : %v", txid))
			utils.RespondWithError(ctx, http.StatusBadRequest, constants.InvalidBody)
			return
		}

		// Validate request body
		if speedData.Speed == nil {
			utils.Logger.Error(fmt.Sprintf("request does not have speed field, txid : %v", txid))
			err := errors.New("invalid request received")
			utils.RespondWithError(ctx, http.StatusBadRequest, err.Error())
			return
		}

		if *speedData.Speed < 0 || *speedData.Speed > 100 {
			utils.Logger.Error(fmt.Sprintf("speed range is incorrect, it's range should be between 0 and 100, txid : %v", txid))
			err := errors.New("speed should be range between 0 and 100")
			utils.RespondWithError(ctx, http.StatusBadRequest, err.Error())
			return
		}

		ctx.Next()
	}
}

func ValidateGetTokenEndointRequest() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		txid := ctx.Request.Header.Get(constants.TransactionID)
		var emailInfo models.Email
		err := ctx.ShouldBindBodyWith(&emailInfo, binding.JSON)
		if err != nil {
			utils.Logger.Error(fmt.Sprintf("error while unmarshaling the request field for get token data validation, txid : %v", txid))
			utils.RespondWithError(ctx, http.StatusBadRequest, constants.InvalidBody)
			return
		}

		// Validate request body
		if emailInfo.Email == "" {
			utils.Logger.Error(fmt.Sprintf("request does not have email field, txid : %v", txid))
			err := fmt.Errorf("invalid request received")
			utils.RespondWithError(ctx, http.StatusBadRequest, err.Error())
			return
		}

		_, parseErr := mail.ParseAddress(emailInfo.Email)
		if parseErr != nil {
			utils.Logger.Error(fmt.Sprintf("email received is incorrect, txid : %v", txid))
			err := fmt.Errorf("invalid email found, err : %v", parseErr)
			utils.RespondWithError(ctx, http.StatusBadRequest, err.Error())
			return
		}

		ctx.Next()
	}
}
