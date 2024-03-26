package main

import (
	"alertmanager-webhook-signal/interfaces"
	"alertmanager-webhook-signal/interfaces/config"
	"fmt"
	"log/slog"
	"net/http"
	"os"
)

const appVersion = "1.0.1" // updated by bumpver

func main() {
	configPath := emptyStringDefault(os.Getenv("CONFIG_PATH"), ".config.yaml")

	cfg, err := config.NewConfig(configPath)
	if err != nil {
		slog.Error("Error reading config", "err", err)
		os.Exit(1)
	}

	alerts := interfaces.NewAlert(
		cfg,
	)

	mux := http.NewServeMux()

	mux.HandleFunc("GET /ping", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("pong"))
	})
	mux.HandleFunc("GET /version", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte(appVersion))
	})

	mux.HandleFunc("POST /api/v3/alertmanager", alerts.Alertmanager)
	mux.HandleFunc("POST /api/v3/grafana", alerts.Grafana)

	mux.HandleFunc("POST /api/v2/alert/alertmanager", alerts.AlertmanagerOld)
	mux.HandleFunc("POST /api/v2/alert/grafana", alerts.Grafana)

	listenInterface := emptyStringDefault(cfg.Config.Server.Interface, "0.0.0.0")
	listenPort := emptyStringDefault(cfg.Config.Server.Port, "10000")
	slog.Info("Server starting", "version", appVersion, "interface", listenInterface, "port", listenPort)
	if err := http.ListenAndServe(fmt.Sprint(listenInterface, ":", cfg.Config.Server.Port), mux); err != nil {
		slog.Error("error starting server", "err", err)
		os.Exit(1)
	}
}

func emptyStringDefault(str, def string) string {
	if str == "" {
		return def
	}
	return str
}
