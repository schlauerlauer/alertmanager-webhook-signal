package interfaces

import (
	"alertmanager-webhook-signal/domain/dto"
	"encoding/base64"
	"errors"
	"fmt"
	"io"
	"net/http"
)

func stringInSlice(compare string, list []string) bool {
	for _, element := range list {
		if element == compare {
			return true
		}
	}
	return false
}

func getImage(url string) (string, error) {
	resp, err := http.Get(url)
	if err != nil {
		return "", errors.New("could not download grafana image")
	}
	defer resp.Body.Close()
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", errors.New("could not download grafana image")
	}
	return base64.StdEncoding.EncodeToString(body), nil
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
