package main

import (
	"fmt"
	"net/http"
	"os"
	"strings"

	"github.com/gin-gonic/gin"
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
	router := gin.Default()

	router.GET("/get_config", func(c *gin.Context) {
		uid := c.DefaultQuery("uid", "4321")
		channel := c.DefaultQuery("channel", "go-smoke")
		c.JSON(http.StatusOK, gin.H{
			"code": 0,
			"data": gin.H{
				"app_id":       "fake-app-id",
				"token":        "fake-token",
				"uid":          uid,
				"channel_name": channel,
				"agent_uid":    "9999",
			},
			"msg": "success",
		})
	})

	startAgentHandler := func(c *gin.Context) {
		var request startAgentRequest
		if err := c.ShouldBindJSON(&request); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"detail": "Invalid request body"})
			return
		}
		if strings.TrimSpace(request.ChannelName) == "" || request.RTCUid <= 0 || request.UserUID <= 0 {
			c.JSON(http.StatusBadRequest, gin.H{"detail": "channel_name, agent_uid, and user_uid are required"})
			return
		}

		c.JSON(http.StatusOK, gin.H{
			"code": 0,
			"data": gin.H{
				"agent_id":     fmt.Sprintf("fake-agent-%d", request.RTCUid),
				"channel_name": request.ChannelName,
				"status":       "started",
			},
			"msg": "success",
		})
	}
	router.POST("/startAgent", startAgentHandler)

	stopAgentHandler := func(c *gin.Context) {
		var request stopAgentRequest
		if err := c.ShouldBindJSON(&request); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"detail": "Invalid request body"})
			return
		}
		if strings.TrimSpace(request.AgentID) == "" {
			c.JSON(http.StatusBadRequest, gin.H{"detail": "agent_id is required"})
			return
		}

		c.JSON(http.StatusOK, gin.H{"code": 0, "msg": "success"})
	}
	router.POST("/stopAgent", stopAgentHandler)

	port := os.Getenv("PORT")
	if strings.TrimSpace(port) == "" {
		port = "8000"
	}

	_ = router.Run(":" + port)
}
