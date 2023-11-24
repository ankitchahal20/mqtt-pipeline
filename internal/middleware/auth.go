package middleware

import (
	"fmt"
	"net/http"

	"github.com/dgrijalva/jwt-go"
	"github.com/gin-gonic/gin"
	"github.com/mqtt-pipeline/internal/constants"
	"github.com/mqtt-pipeline/internal/utils"
)


func Authorization() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		txid := ctx.Request.Header.Get(constants.TransactionID)
		tokenString := ctx.GetHeader("Authorization")
		if tokenString == "" {
			utils.Logger.Error(fmt.Sprintf("authorization header is empty, txid : %v", txid))
			ctx.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
			ctx.Abort()
			return
		}

		token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
			return constants.SecretKey, nil
		})

		if err != nil {
			switch err1 := err.(type) {
			case *jwt.ValidationError: // something was wrong during the validation
				vErr := err1
				switch vErr.Errors {
				case jwt.ValidationErrorExpired:
					utils.Logger.Error(fmt.Sprintf("token expired, txid : %v", txid))
					utils.RespondWithError(ctx, http.StatusUnauthorized, "token expired")
					return
				default:
					utils.Logger.Error(fmt.Sprintf("error while parsing token, txid : %v", txid))
					utils.RespondWithError(ctx, http.StatusInternalServerError, "error while parsing token")
					return
				}
			default: // something else went wrong
			utils.Logger.Error(fmt.Sprintf("error while parsing token, txid : %v", txid))
			utils.RespondWithError(ctx, http.StatusInternalServerError, "error while parsing token")
			return
			}
		}
		if !token.Valid {
			utils.Logger.Error(fmt.Sprintf("invalid token received, txid : %v", txid))
			utils.RespondWithError(ctx, http.StatusUnauthorized, "invalid token")
			return
		}

		utils.Logger.Info(fmt.Sprintf("received valid token, txid : %v", txid))
		ctx.Next()
	}
}