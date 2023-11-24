package middleware

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/mqtt-pipeline/internal/constants"
	"github.com/mqtt-pipeline/internal/models"
	"github.com/mqtt-pipeline/internal/utils"
	"gotest.tools/assert"
)

func TestValidateGetTokenRequestInput(t *testing.T) {
	// init logging client
	utils.InitLogClient()

	// Case 1 : email missing
	requestFields := models.Email{}

	jsonValue, _ := json.Marshal(requestFields)

	w := httptest.NewRecorder()
	_, e := gin.CreateTestContext(w)
	req, _ := http.NewRequest(http.MethodPost, "/v1", bytes.NewBuffer(jsonValue))
	req.Header.Add(constants.ContentType, "application/json")
	e.Use(ValidateGetTokenEndointRequest())
	e.ServeHTTP(w, req)
	assert.Equal(t, http.StatusBadRequest, w.Code)

	// Case 2 : invalid email
	requestFields = models.Email{
		Email: "some-invalid-mail.com",
	}

	jsonValue, _ = json.Marshal(requestFields)

	w = httptest.NewRecorder()
	_, e = gin.CreateTestContext(w)
	req, _ = http.NewRequest(http.MethodPost, "/v1", bytes.NewBuffer(jsonValue))
	fmt.Println("Req : ", req)
	req.Header.Add(constants.ContentType, "application/json")
	e.Use(ValidateGetTokenEndointRequest())
	e.ServeHTTP(w, req)
	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestValidatePublishRequestInput(t *testing.T) {
	// init logging client
	utils.InitLogClient()

	// Case 1 : speed missing
	requestFields := models.SpeedData{}

	jsonValue, _ := json.Marshal(requestFields)

	w := httptest.NewRecorder()
	_, e := gin.CreateTestContext(w)
	req, _ := http.NewRequest(http.MethodPost, "/v1/publish", bytes.NewBuffer(jsonValue))
	fmt.Println("Req : ", req)
	req.Header.Add(constants.ContentType, "application/json")
	e.Use(ValidatePublishEndpointRequest())
	e.ServeHTTP(w, req)
	assert.Equal(t, http.StatusBadRequest, w.Code)

	// Case 2 : invalid speed 
	speed := -100
	requestFields = models.SpeedData{
		Speed: &speed,
	}

	jsonValue, _ = json.Marshal(requestFields)

	w = httptest.NewRecorder()
	_, e = gin.CreateTestContext(w)
	req, _ = http.NewRequest(http.MethodPost, "/v1/publish", bytes.NewBuffer(jsonValue))
	req.Header.Add(constants.ContentType, "application/json")
	e.Use(ValidatePublishEndpointRequest())
	e.ServeHTTP(w, req)
	assert.Equal(t, http.StatusBadRequest, w.Code)
}