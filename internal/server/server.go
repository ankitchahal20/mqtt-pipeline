package server

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/mqtt-pipeline/internal/constants"
	"github.com/mqtt-pipeline/internal/middleware"
	"github.com/mqtt-pipeline/internal/service"
)

func registerGetTokenEndPoints(handler gin.IRoutes) {
	handler.POST(constants.ForwardSlash+strings.Join([]string{}, constants.ForwardSlash), middleware.ValidateGetTokenEndointRequest(), service.GenerateToken())
}

func registerPublishEndpointPoints(handler gin.IRoutes) {
	handler.POST(constants.ForwardSlash+strings.Join([]string{constants.Publish}, constants.ForwardSlash), middleware.Authorization(), middleware.ValidatePublishEndpointRequest(), service.Publish())
}

func registerSpeedDataEndPoints(handler gin.IRoutes) {
	handler.GET(constants.ForwardSlash, middleware.Authorization(),  service.GetSpeedData())
}

func Start() {
	plainHandler := gin.New()

	mqttPipelineHandler := plainHandler.Group(constants.ForwardSlash + constants.Version).Use(gin.Recovery()).
		Use(middleware.SetTransactionId())
	registerGetTokenEndPoints(mqttPipelineHandler)
	registerPublishEndpointPoints(mqttPipelineHandler)
	registerSpeedDataEndPoints(mqttPipelineHandler)

	srv := &http.Server{
		Handler:      plainHandler,
		Addr:         ":4000",
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 10 * time.Second,
	}

	// Start Server
	go func() {
		log.Println("Starting Server")
		if err := srv.ListenAndServe(); err != nil {
			log.Fatal(err)
		}
	}()

	waitForShutdown(srv)
}

func waitForShutdown(srv *http.Server) {

	/*
		if somewhere you are listening for output from a channel but in the meanwhile that channel not being given any input,
		the listening place will be blocked until it receives the output.
	*/

	interruptChan := make(chan os.Signal, 1)

	/*
		SIGINT means is Signal Interrupted, send when the user types the INTR character (e.g. Ctrl-C).
		SIGTERM signal is a generic signal used to cause program termination.
		The below statement is to send the interruot signal to the channel "interruptChan".
	*/

	signal.Notify(interruptChan, os.Interrupt, syscall.SIGINT, syscall.SIGTERM)

	// Block until we receive our signal.
	<-interruptChan

	// Create a deadline to wait for.
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*10)

	/*
		some zombie connections may still be there and use your memory,
		in order to solve this, the safest way is to setup a timeout threshold using the Context.
	*/

	defer cancel()

	/*
		The timer context and the cancel function (the cancel function releases its resources) will be returned,
		 and you can use that context to perform Shutdown(ctx) ,
		inside the Shutdown it check if the timer context Done channel is closed and will not run indefinitely.
	*/

	srv.Shutdown(ctx)

	log.Println("Shutting down")
	os.Exit(0)
}
