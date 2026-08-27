package main

import (
	"log"
	"math/rand"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"time"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
)

type startAgentRequest struct {
	ChannelName string `json:"channelName"`
	RTCUid      int    `json:"rtcUid"`
	UserUID     int    `json:"userUid"`
}

type stopAgentRequest struct {
	AgentID string `json:"agentId"`
}

func main() {
	rand.Seed(time.Now().UnixNano())
	loadEnvFiles()

	service, err := newAgentService()
	if err != nil {
		log.Printf("warning: failed to initialize SDK: %v", err)
		log.Printf("service will fail if endpoints are called without proper configuration")
	}

	router := newRouter(service)

	port := firstNonEmpty(os.Getenv("PORT"), "8000")
	if err := router.Run(":" + port); err != nil {
		log.Fatalf("start gin server: %v", err)
	}
}

func newRouter(service *agentService) *gin.Engine {
	router := gin.Default()
	router.Use(cors.New(cors.Config{
		AllowAllOrigins:  true,
		AllowMethods:     []string{"GET", "POST", "OPTIONS"},
		AllowHeaders:     []string{"Origin", "Content-Type", "Accept", "Authorization"},
		AllowCredentials: true,
	}))

	router.GET("/get_config", func(c *gin.Context) {
		if service == nil {
			c.JSON(http.StatusInternalServerError, gin.H{"detail": "Service not properly configured. Please check environment variables."})
			return
		}

		uid, err := parseOptionalInt(c.Query("uid"))
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"detail": "uid must be an integer"})
			return
		}

		config, err := service.generateConfig(c.Query("channel"), uid)
		if err != nil {
			status, detail := toHTTPError(err)
			c.JSON(status, gin.H{"detail": detail})
			return
		}

		c.JSON(http.StatusOK, gin.H{
			"code": 0,
			"data": config,
			"msg":  "success",
		})
	})

	startAgentHandler := func(c *gin.Context) {
		if service == nil {
			c.JSON(http.StatusInternalServerError, gin.H{"detail": "Service not properly configured. Please check environment variables."})
			return
		}

		var request startAgentRequest
		if err := c.ShouldBindJSON(&request); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"detail": "Invalid request body"})
			return
		}

		result, err := service.start(request.ChannelName, request.RTCUid, request.UserUID)
		if err != nil {
			status, detail := toHTTPError(err)
			c.JSON(status, gin.H{"detail": detail})
			return
		}

		c.JSON(http.StatusOK, gin.H{
			"code": 0,
			"data": result,
			"msg":  "success",
		})
	}
	router.POST("/startAgent", startAgentHandler)

	stopAgentHandler := func(c *gin.Context) {
		if service == nil {
			c.JSON(http.StatusInternalServerError, gin.H{"detail": "Service not properly configured. Please check environment variables."})
			return
		}

		var request stopAgentRequest
		if err := c.ShouldBindJSON(&request); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"detail": "Invalid request body"})
			return
		}

		if err := service.stop(request.AgentID); err != nil {
			status, detail := toHTTPError(err)
			c.JSON(status, gin.H{"detail": detail})
			return
		}

		c.JSON(http.StatusOK, gin.H{
			"code": 0,
			"msg":  "success",
		})
	}
	router.POST("/stopAgent", stopAgentHandler)
	return router
}

func loadEnvFiles() {
	baseDir, err := os.Getwd()
	if err != nil {
		return
	}

	_ = godotenv.Load(filepath.Join(baseDir, ".env.local"))
	_ = godotenv.Load(filepath.Join(baseDir, ".env"))
}

func parseOptionalInt(value string) (int, error) {
	if value == "" {
		return 0, nil
	}
	return strconv.Atoi(value)
}

func toHTTPError(err error) (int, string) {
	switch {
	case err == nil:
		return http.StatusOK, ""
	default:
		message := err.Error()
		if isValidationError(message) {
			return http.StatusBadRequest, message
		}
		return http.StatusInternalServerError, message
	}
}

func isValidationError(message string) bool {
	switch message {
	case "channel_name is required and cannot be empty",
		"agent_uid is required and cannot be empty",
		"user_uid is required and cannot be empty",
		"agent_id is required and cannot be empty":
		return true
	default:
		return false
	}
}
