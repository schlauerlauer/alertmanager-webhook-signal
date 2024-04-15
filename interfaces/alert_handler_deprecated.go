package interfaces

import (
	"alertmanager-webhook-signal/domain/dto"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"net/http"
)

func (al *Alert) AlertmanagerDeprecated(w http.ResponseWriter, req *http.Request) {
	buff, _ := io.ReadAll(req.Body)

	var alert dto.Alertmanager
	err := json.Unmarshal(buff, &alert)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		slog.Error("could not unmarshal json", "err", err)
		return
	}

	messages := al.mapAM2SignalDeprecated(alert)
	for _, message := range messages {
		if code, err := al.sendSignal(message); err != nil {
			slog.Warn("error sending signal message", "err", err, "statusCode", code)
		}
	}
}

// this is using annotations instead of labels for the recipients. Please use the newer mapAM2Signal, which uses labels
func (al *Alert) mapAM2SignalDeprecated(alert dto.Alertmanager) []dto.SignalMessage {
	messages := make([]dto.SignalMessage, len(alert.Alerts))

	for idx, element := range alert.Alerts {
		recipients := al.config.Signal.Recipients
		message := fmt.Sprint("Alert ", element.Labels["alertname"], " is ", element.Status)
		for key, val := range element.Annotations {
			if !stringInSlice(key, al.config.AMConfig.IgnoreAnnotations) {
				message += fmt.Sprintf("\n%v: %v", key, val)
			}
			if key == "recipients" {
				newReceiver, ok := al.config.Recipients[val.(string)]
				if ok {
					recipients = nil
					recipients = append(recipients, fmt.Sprintf("%v", newReceiver))
				}
			}
		}
		for key, val := range element.Labels {
			if !stringInSlice(key, al.config.AMConfig.IgnoreLabels) {
				message += fmt.Sprintf("\n%v: %v", key, val)
			}
		}
		if al.config.AMConfig.GeneratorURL {
			message += fmt.Sprintf("\nuri: %v", element.GeneratorURL)
		}
		messages[idx] = dto.SignalMessage{
			Message:     message,
			Number:      al.config.Signal.Number,
			Recipients:  recipients,
			Attachments: nil,
			TextMode:    "normal",
		}
	}

	return messages
}
