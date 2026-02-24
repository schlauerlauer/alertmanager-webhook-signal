package main

import (
	"alertmanager-webhook-signal/internal/alerts"
	"alertmanager-webhook-signal/internal/config"
	"alertmanager-webhook-signal/internal/util"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"text/template"
	"time"

	"github.com/lmittmann/tint"
	"github.com/schlauerlauer/go-middleware"
)

const appVersion = "1.1.1" // updated by bumpver

func main() {
	slog.SetDefault(slog.New(
		tint.NewHandler(os.Stdout, &tint.Options{
			Level:      slog.LevelInfo,
			TimeFormat: time.TimeOnly,
		}),
	))

	const defaultGrafanaTemplate = `{{ if eq .State "alerting" }}❗{{ else }}✅{{ end }} {{ .Title }}
{{ .RuleName }}
{{ .Message }}
{{ .RuleUrl }}`

	const defaultAlertmanageTemplate = `{{ if eq .Alert.Status "firing" }}❗{{ else }}✅{{ end }} Alert **{{ .Alertname }}** is {{ .Alert.Status }}

{{- if gt (len (.Alert.Labels)) 0 }}

Labels:
{{- range $key, $value := .Alert.Labels }}
  - {{ $key }}: {{ $value }}
{{- end }}
{{- end }}

{{- if gt (len (.Alert.Annotations)) 0 }}

Annotations:
{{- range $key, $value := .Alert.Annotations }}
  - {{ $key }}: {{ $value }}
{{- end }}
{{- end }}

{{- if .Config.GeneratorURL }}
{{ .Alert.GeneratorURL}}
{{ end -}}
`

	configPath := util.StringDefault(os.Getenv("CONFIG_PATH"), "/config.yaml")

	cfg, err := config.NewConfig(configPath)
	if err != nil {
		slog.Error("Error reading config", "err", err)
		os.Exit(1)
	}

	grafanaTemplate, err := template.New("grafana").Parse(util.StringDefault(cfg.Templates.Grafana, defaultGrafanaTemplate))
	if err != nil {
		slog.Error("error parsing grafana template", "err", err)
		os.Exit(1)
	}

	alertmanagerTemplate, err := template.New("alertmanager").Parse(util.StringDefault(cfg.Templates.Alertmanager, defaultAlertmanageTemplate))
	if err != nil {
		slog.Error("error parsing alertmanager template", "err", err)
		os.Exit(1)
	}

	alert := alerts.NewAlert(
		cfg,
		grafanaTemplate,
		alertmanagerTemplate,
	)

	mux := http.NewServeMux()

	mux.HandleFunc("GET /ping", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("pong"))
	})
	mux.HandleFunc("GET /version", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte(appVersion))
	})

	mux.HandleFunc("POST /alertmanager", alert.Alertmanager)
	mux.HandleFunc("POST /api/v3/alertmanager", alert.Alertmanager)

	mux.HandleFunc("POST /grafana", alert.Grafana)
	mux.HandleFunc("POST /api/v2/alert/grafana", alert.Grafana)
	mux.HandleFunc("POST /api/v3/grafana", alert.Grafana)

	listenInterface := util.StringDefault(cfg.Server.Interface, "0.0.0.0")
	listenPort := util.StringDefault(cfg.Server.Port, "10000")

	slog.Info("Server starting", "version", appVersion, "interface", listenInterface, "port", listenPort)
	if err := http.ListenAndServe(fmt.Sprint(listenInterface, ":", listenPort), middleware.Logging(mux)); err != nil {
		slog.Error("error starting server", "err", err)
		os.Exit(1)
	}
}
