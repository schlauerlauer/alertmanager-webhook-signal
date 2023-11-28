package interfaces

import (
	"alertmanager-webhook-signal/domain/dto"
	"alertmanager-webhook-signal/interfaces/config"
	"bytes"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"net/http"

	"github.com/gin-gonic/gin"
)

type Alert struct {
	config *config.ConfigData
}

func NewAlert(config *config.ConfigService) *Alert {
	return &Alert{
		config: config.Config,
	}
}

const (
	ProviderGrafana      string = "grafana"
	ProviderAlertmanager string = "alertmanager"
)

func (al *Alert) ReceiveAlertmanager(c *gin.Context) {
	buff, _ := io.ReadAll(c.Request.Body)

	var alert dto.Alertmanager
	err := json.Unmarshal(buff, &alert)
	if err != nil {
		c.AbortWithError(http.StatusBadRequest, errors.New("could not unmarshal json"))
		slog.Error("Error unmarshalling json", "err", err)
		return
	}

	al.mapAM2Signal(&alert, c)
	return
}

func (al *Alert) ReceiveGrafana(c *gin.Context) {
	buff, _ := io.ReadAll(c.Request.Body)

	var alert dto.GrafanaAlert
	err := json.Unmarshal(buff, &alert)
	if err != nil {
		c.AbortWithError(http.StatusBadRequest, errors.New("could not unmarshal json"))
		slog.Error("Error unmarshalling json", "err", err)
		return
	}

	al.mapGrafana2Signal(alert, c)
	return
}

func (al *Alert) ReceiveV2(c *gin.Context) {
	provider := c.Param("provider")
	buff, _ := io.ReadAll(c.Request.Body)

	switch provider {
	case ProviderAlertmanager:
		var alert dto.Alertmanager
		json.Unmarshal(buff, &alert)
		al.mapAM2SignalDeprecated(alert, c)
		return
	case ProviderGrafana:
		var alert dto.GrafanaAlert
		json.Unmarshal(buff, &alert)
		al.mapGrafana2Signal(alert, c)
		return
	default:
		c.AbortWithError(http.StatusNotFound, errors.New("provider not available"))
		return
	}
}

func (al *Alert) mapReceiver(receiver string) string {
	for r := range al.config.Recipients {
		if r == receiver {
			return fmt.Sprintf("%v", al.config.Recipients[receiver])
		}
	}
	return ""
}

func (al *Alert) sendSignal(m dto.SignalMessage, c *gin.Context) {
	payloadBuf := new(bytes.Buffer)
	json.NewEncoder(payloadBuf).Encode(m)
	if al.config.Server.Debug {
		fmt.Println(payloadBuf)
	}
	req, _ := http.NewRequest("POST", al.config.Signal.Send, payloadBuf)
	client := &http.Client{}
	res, e := client.Do(req)
	if e != nil {
		c.String(http.StatusInternalServerError, "could not reach signal api.")
		return
	}
	defer res.Body.Close()
	fmt.Println("signal response:", res.Status)
}

func getImage(url string, c *gin.Context) string {
	resp, e := http.Get(url)
	if e != nil {
		c.String(http.StatusInternalServerError, "could not download grafana image.")
		return ""
	}
	defer resp.Body.Close()
	b, e := io.ReadAll(resp.Body)
	if e != nil {
		c.String(http.StatusInternalServerError, "could not download grafana image.")
		return ""
	}
	return base64.StdEncoding.EncodeToString(b)
}

// this is using annotations instead of labels for the recipients. Please use the newer mapAM2Signal, which uses labels
func (al *Alert) mapAM2SignalDeprecated(a dto.Alertmanager, c *gin.Context) {
	for _, element := range a.Alerts {
		recipients := al.config.Signal.Recipients
		message := fmt.Sprint("Alert ", element.Labels["alertname"], " is ", element.Status)
		for k, v := range element.Annotations {
			if !stringInSlice(k, al.config.AMConfig.IgnoreAnnotations) {
				message += fmt.Sprintf("\n%v: %v", k, v)
			}
			if k == "recipients" {
				newReceiver := al.mapReceiver(v.(string))
				fmt.Println(newReceiver)
				if newReceiver != "" {
					recipients = nil
					recipients = append(recipients, newReceiver)
				}
			}
		}
		for k, v := range element.Labels {
			if !stringInSlice(k, al.config.AMConfig.IgnoreLabels) {
				message += fmt.Sprintf("\n%v: %v", k, v)
			}
		}
		if al.config.AMConfig.GeneratorURL {
			message += fmt.Sprintf("\nuri: %v", element.GeneratorURL)
		}
		signal := dto.SignalMessage{
			Message:     message,
			Number:      al.config.Signal.Number,
			Recipients:  recipients,
			Attachments: []string{},
		}
		al.sendSignal(signal, c)
	}
}

func (al *Alert) mapAM2Signal(a *dto.Alertmanager, c *gin.Context) {
	for _, element := range a.Alerts {
		recipients := al.config.Signal.Recipients
		message := fmt.Sprint("Alert ", element.Labels["alertname"], " is ", element.Status)
		for k, v := range element.Annotations {
			if !stringInSlice(k, al.config.AMConfig.IgnoreAnnotations) {
				message += fmt.Sprintf("\n%v: %v", k, v)
			}
		}
		for k, v := range element.Labels {
			if !stringInSlice(k, al.config.AMConfig.IgnoreLabels) {
				message += fmt.Sprintf("\n%v: %v", k, v)
			}
			if k == "recipients" {
				newReceiver := al.mapReceiver(v.(string))
				fmt.Println(newReceiver)
				if newReceiver != "" {
					recipients = nil
					recipients = append(recipients, newReceiver)
				}
			}
		}
		if al.config.AMConfig.GeneratorURL {
			message += fmt.Sprintf("\nuri: %v", element.GeneratorURL)
		}
		signal := dto.SignalMessage{
			Message:     message,
			Number:      al.config.Signal.Number,
			Recipients:  recipients,
			Attachments: []string{},
		}
		al.sendSignal(signal, c)
	}
}

func (al *Alert) mapGrafana2Signal(ga dto.GrafanaAlert, c *gin.Context) {
	var encoded string
	if ga.ImageUrl != "" {
		encoded = getImage(ga.ImageUrl, c)
	}
	message := fmt.Sprintf("%s\n%s\n%s\n%s",
		ga.Title,
		ga.RuleName,
		ga.Message,
		ga.RuleUrl,
	)
	signal := dto.SignalMessage{
		Message:     message,
		Number:      al.config.Signal.Number,
		Recipients:  al.config.Signal.Recipients,
		Attachments: []string{encoded},
	}
	al.sendSignal(signal, c)
}
