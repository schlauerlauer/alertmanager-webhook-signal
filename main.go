package main

import (
	"alertmanager-webhook-signal/interfaces"
	"alertmanager-webhook-signal/interfaces/config"
	"errors"
	"log/slog"
	"os"

	"github.com/gin-gonic/gin"
)

const appVersion = "1.0.0" // updated by bumpver

func main() {
	configPath := os.Getenv("CONFIG_PATH")
	if configPath == "" {
		configPath = "./config.yaml"
	}

	cfg, err := config.NewConfig(configPath)
	if err != nil {
		slog.Error("Error reading config", "err", err)
		os.Exit(1)
	}

	alerts := interfaces.NewAlert(
		cfg,
	)

	gin.SetMode(gin.ReleaseMode)
	r := gin.Default()

	api := r.Group("/api")

	v3 := api.Group("v3")
	{
		v3.POST("alertmanager", alerts.ReceiveAlertmanager)
		v3.POST("grafana", alerts.ReceiveGrafana)
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

	slog.Info("Server starting", "version", appVersion, "port", cfg.Config.Server.Port)
	r.Run(":" + cfg.Config.Server.Port)
}
