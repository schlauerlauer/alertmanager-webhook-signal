package main

import (
	"alertmanager-webhook-signal/interfaces"
	"alertmanager-webhook-signal/interfaces/config"
	"errors"
	"fmt"
	"log"
	"os"

	"github.com/gin-gonic/gin"
)

const appVersion = "3.0.0" // updated by bumpver

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

	api := r.Group("/api")

	v3 := api.Group("v3")
	{
		v3.POST(":provider", alerts.ReceiveV3)
	}

	v2 := api.Group("v2")
	{
		v2.POST(":provider", alerts.ReceiveV2)
	}

	v1 := api.Group("v1")
	{
		v1.POST("alert", func(c *gin.Context) {
			c.AbortWithError(299, errors.New("this api version is deprecated. Please use \"/api/v3/alertmanager\" instead"))
		})
	}

	log.Println(fmt.Sprintf("Server (v%s) started. Listening on port %s", appVersion, cfg.Config.Server.Port))
	r.Run(":" + cfg.Config.Server.Port)
}
