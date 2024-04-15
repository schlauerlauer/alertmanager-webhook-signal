package interfaces

import (
	"alertmanager-webhook-signal/domain/dto"
	"alertmanager-webhook-signal/interfaces/config"
	"bytes"
	"encoding/json"
	"errors"
	"io"
	"log/slog"
	"net/http"
	"text/template"
)

type Alert struct {
	config               *config.ConfigData
	grafanaTemplate      *template.Template
	alertmanagerTemplate *template.Template
}

type alertTemplateData struct {
	Alertname   string
	Alert       dto.AMAlert
	Config      config.AlertmanagerConfig
	Labels      map[string]interface{}
	Annotations map[string]interface{}
}

func NewAlert(
	config *config.ConfigService,
	grafanaTemplate *template.Template,
	alertmanagerTemplate *template.Template,
) *Alert {
	return &Alert{
		config:               config.Config,
		grafanaTemplate:      grafanaTemplate,
		alertmanagerTemplate: alertmanagerTemplate,
	}
}

func (al *Alert) Alertmanager(w http.ResponseWriter, req *http.Request) {
	buff, _ := io.ReadAll(req.Body)

	var alert dto.Alertmanager
	err := json.Unmarshal(buff, &alert)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		slog.Error("could not unmarshal json", "err", err)
		return
	}

	messages := al.alertToSignal(alert)
	for _, message := range messages {
		if code, err := al.sendSignal(message); err != nil {
			slog.Warn("error sending signal message", "err", err, "statusCode", code)
		}
	}
}

func (al *Alert) Grafana(w http.ResponseWriter, req *http.Request) {
	buff, _ := io.ReadAll(req.Body)

	var alert dto.GrafanaAlert
	err := json.Unmarshal(buff, &alert)
	if err != nil {
		slog.Error("could not unmarshal json", "err", err)
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	message := al.grafanaToSignal(alert)
	if code, err := al.sendSignal(message); err != nil {
		slog.Warn("error sending signal message", "err", err, "code", code)
		if code >= 100 {
			w.WriteHeader(code)
		}
	}
}

func (al *Alert) sendSignal(message dto.SignalMessage) (int, error) {
	payloadBuf := new(bytes.Buffer)
	json.NewEncoder(payloadBuf).Encode(message)
	if al.config.Server.Debug {
		slog.Debug("payload", "message", payloadBuf)
	}
	req, _ := http.NewRequest("POST", al.config.Signal.Send, payloadBuf)
	client := &http.Client{}
	res, err := client.Do(req)
	if err != nil {
		return 400, errors.New("could not reach signal api")
	}
	defer res.Body.Close()

	if res.StatusCode >= 400 {
		body, err := io.ReadAll(res.Body)
		if err == nil {
			slog.Warn("response", "body", string(body))
		}
		return res.StatusCode, errors.New("message was not successful")
	}

	return http.StatusOK, nil
}

func (al *Alert) grafanaToSignal(alert dto.GrafanaAlert) dto.SignalMessage {
	var message bytes.Buffer
	err := al.grafanaTemplate.Execute(&message, alert)
	if err != nil {
		slog.Error("could not execute grafana message template", "err", err)
	}

	signal := dto.SignalMessage{
		Message:    message.String(),
		Number:     al.config.Signal.Number,
		Recipients: al.config.Signal.Recipients,
		TextMode:   ternary(al.config.Signal.TextModeNormal, "normal", "styled"),
	}

	if alert.ImageUrl != "" {
		attachment, err := getImage(alert.ImageUrl)
		if err != nil {
			slog.Warn("could not attach image to signal message", "err", err)
		} else {
			signal.Attachments = &[]string{
				attachment,
			}
		}
	}

	return signal
}

func (al *Alert) alertToSignal(alert dto.Alertmanager) []dto.SignalMessage {
	messages := make([]dto.SignalMessage, len(alert.Alerts))
	for idx, alertElement := range alert.Alerts {
		customRecipients := []string{}
		alertName := alertElement.Labels["alertname"].(string)

		if recipients, ok := alertElement.Labels["recipients"]; ok {
			newReceiver, ok := al.config.Recipients[recipients.(string)]
			if ok {
				customRecipients = append(customRecipients, newReceiver)
			}
		}

		for _, annotation := range al.config.AMConfig.IgnoreAnnotations {
			delete(alertElement.Annotations, annotation)
		}
		for _, label := range al.config.AMConfig.IgnoreLabels {
			delete(alertElement.Labels, label)
		}

		var message bytes.Buffer
		err := al.alertmanagerTemplate.Execute(&message, alertTemplateData{
			Alertname:   alertName,
			Alert:       alertElement,
			Config:      al.config.AMConfig,
			Labels:      alertElement.Labels,
			Annotations: alertElement.Annotations,
		})
		if err != nil {
			slog.Error("could not execute alertmanager message template", "err", err)
		}

		newMesage := dto.SignalMessage{
			Message:     message.String(),
			Number:      al.config.Signal.Number,
			Recipients:  al.config.Signal.Recipients,
			Attachments: nil,
			TextMode:    ternary(al.config.Signal.TextModeNormal, "normal", "styled"),
		}
		if len(customRecipients) > 0 {
			newMesage.Recipients = customRecipients
		}

		messages[idx] = newMesage
	}

	return messages
}
