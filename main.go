package main

import (
	"alertmanager-webhook-signal/interfaces"
	"alertmanager-webhook-signal/interfaces/config"
	"errors"
	"fmt"
	"log"
	"net/http"
	"os"

	"github.com/gin-gonic/gin"
)

func main() {
	log.SetOutput(os.Stdout)

	configPath := os.Getenv("CONFIG_PATH")
	if configPath == "" {
		configPath = "./config.yaml"
	}

	cfg, err := config.NewConfig(configPath)
	if err != nil {
		log.Fatal(err)
	}

	alerts := interfaces.NewAlert(
		cfg,
	)

	gin.SetMode(gin.ReleaseMode)
	r := gin.Default()

	r.GET("/-/reload", func(c *gin.Context) {
		newConfig, err := config.NewConfig("./config.yaml")
		if err != nil {
			c.AbortWithError(http.StatusInternalServerError, err)
		}
		cfg = newConfig
	})

	api := r.Group("/api")

	api.POST(":version/:provider", alerts.Receive)
	api.POST("v1/alert", func(c *gin.Context) {
		c.AbortWithError(299, errors.New("this api version is deprecated. Please use \"/api/v3/alertmanager\" instead"))
	})

	fmt.Println("Starting server. Listening on port:", cfg.Config.Server.Port)
	r.Run(":"+cfg.Config.Server.Port)
}
