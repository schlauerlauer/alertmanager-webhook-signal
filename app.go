package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/gin-gonic/gin"
	"gopkg.in/yaml.v2"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"time"
)

type Config struct {
	Signal struct {
		Number            string   `yaml:"number"`
		Recipients        []string `yaml:"recipients"`
		Send              string   `yaml:"send"`
		IgnoreLabels      []string `yaml:"ignoreLabels"`
		IgnoreAnnotations []string `yaml:"ignoreAnnotations"`
		GeneratorURL      bool     `yaml:"generatorURL"`
	} `yaml:"signal"`
	Server struct {
		Port    string        `yaml:"port"`
		Timeout time.Duration `yaml:"timeout"`
	}
	Recipients map[string]interface{} `yaml:"recipients"`
}

func NewConfig(configPath string) (*Config, error) {
	// Create config structure
	config := &Config{}

	// Open config file
	file, err := os.Open(configPath)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	// Init new YAML decode
	d := yaml.NewDecoder(file)

	// Start YAML decoding from file
	if err := d.Decode(&config); err != nil {
		return nil, err
	}
	return config, nil
}

type Alertmanager struct {
	Version				string					`json:"version"`
	GroupKey			string					`json:"groupKey"`
	TruncatedAlerts		int						`json:"truncatedAlerts"`
	Status				string					`json:"status"`
	Receiver			string					`json:"receiver"`
	GroupLabels			map[string]interface{}	`json:"groupLabels"`
	CommonLabels		map[string]interface{}	`json:"commonLabels"`
	CommonAnnotations	map[string]interface{}	`json:"commonAnnotations"`
	ExternalURL			string					`json:"externalURL"`
	Alerts				[]AMAlert				`json:"alerts"`
}

type AMAlert struct {
	Status				string                 	`json:"status"`
	Labels				map[string]interface{} 	`json:"labels"`
	Annotations			map[string]interface{} 	`json:"annotations"`
	StartsAt			string                 	`json:"startsAt"`
	EndsAt				string                 	`json:"endsAt"`
	GeneratorURL		string                 	`json:"generatorURL"`
}

type GrafanaAlert struct {
	DashboardId			int						`json:"dashboardId"`
	EvalMatches			[]GrafanaMatches		`json:"evalMatches"`
	ImageUrl			string					`json:"imageUrl"`
	Message				string					`json:"message"`
	OrgId				int						`json:"orgId"`
	PanelId				int						`json:"panelId"`
	RuleId				string					`json:"ruleId"`
	RuleName			string					`json:"ruleName"`
	RuleUrl				string					`json:"ruleUrl"`
	State				string					`json:"state"`
	Tags				map[string]interface{}	`json:"tags"`
	Title				string					`json:"title"`
}

type GrafanaMatches struct {
	Value				int						`json:"value"`
	Metric				string					`json:"metric"`
	Tags				map[string]interface{}	`json:"tags"`
}

type SignalMessage struct {
	Attachments []string `json:"base64_attachments"`
	Message     string   `json:"message"`
	Number      string   `json:"number"`
	Recipients  []string `json:"recipients"`
}

var cfg, _ = NewConfig("./config.yaml")

func checkConfig() {
	if cfg.Server.Port == "" {
		log.Fatal("Server port not set.")
	}
	if cfg.Signal.Number == "" {
		log.Fatal("Signal number not set.")
	}
	if len(cfg.Signal.Recipients) == 0 {
		log.Fatal("Signal default recipients not set.")
	}
	if cfg.Signal.Send == "" {
		log.Fatal("Signal URL not set.")
	}
}

func main() {
	checkConfig()
	gin.SetMode(gin.ReleaseMode)
	fmt.Println("Starting server. Listening on port:", cfg.Server.Port)
	r := gin.Default()
	r.GET("/-/reload", reload)
	r.POST("/api/v2/:provider", receive)
	r.POST("/api/v1/alert", deprecated)
	r.Run(":"+cfg.Server.Port)
}

func deprecated(c *gin.Context) {
	b, _ := ioutil.ReadAll(c.Request.Body)
	var alert Alertmanager
	json.Unmarshal(b, &alert)
	c.String(http.StatusOK, "This api version is deprecated. Please use /api/v2/alertmanager instead.")
	mapAM2Signal(alert, c)
}

func reload(c *gin.Context) {
	var err error
	cfg, err = NewConfig("./config.yaml")
	if err != nil {
		fmt.Println(err)
	}
}

func receive(c *gin.Context) {
	b, _ := ioutil.ReadAll(c.Request.Body)
	switch c.Param("provider") {
	case "alertmanager":
		var alert Alertmanager
		json.Unmarshal(b, &alert)
		mapAM2Signal(alert, c)
	case "grafana":
		var alert GrafanaAlert
		json.Unmarshal(b, &alert)
		mapGrafana2Signal(alert, c)
	default:
		c.String(http.StatusNotFound, "provider not available")
		return
	}
}

func mapGrafana2Signal(ga GrafanaAlert, c *gin.Context) {
	signal := SignalMessage{Message: ga.Message, Number: cfg.Signal.Number, Recipients: cfg.Signal.Recipients}
	sendSignal(signal, c)
}

func mapAM2Signal(a Alertmanager, c *gin.Context) {
	for _, element := range a.Alerts {
		recipients := cfg.Signal.Recipients
		recipientName := "default"
		message := "Alert " + fmt.Sprint(element.Labels["alertname"]) + " is " + element.Status
		for k, v := range element.Annotations {
			if !stringInSlice(k, cfg.Signal.IgnoreAnnotations) {
				message += fmt.Sprintf("\n%v: %v", k, v)
			}
			if k == "recipients" {
				for r := range cfg.Recipients {
					if v == r {
						var _newRec string = fmt.Sprintf("%v", cfg.Recipients[v.(string)])
						recipients = nil
						recipients = append(recipients, _newRec)
						recipientName = r
					}
				}
			}
		}
		for k, v := range element.Labels {
			if !stringInSlice(k, cfg.Signal.IgnoreLabels) {
				message += fmt.Sprintf("\n%v: %v", k, v)
			}
		}
		if cfg.Signal.GeneratorURL {
			message += fmt.Sprintf("\nuri: %v", element.GeneratorURL)
		}
		signal := SignalMessage{Message: message, Number: cfg.Signal.Number, Recipients: recipients}
		fmt.Println("Sending alert:", element.Labels["alertname"], "to recipient:", recipientName)
		sendSignal(signal, c)
	}
}

func stringInSlice(a string, list []string) bool {
	for _, b := range list {
		if b == a {
			return true
		}
	}
	return false
}

func sendSignal(m SignalMessage, c *gin.Context) {
	payloadBuf := new(bytes.Buffer)
	json.NewEncoder(payloadBuf).Encode(m)
	req, _ := http.NewRequest("POST", cfg.Signal.Send, payloadBuf)
	client := &http.Client{}
	res, e := client.Do(req)
	if e != nil {
		c.String(http.StatusInternalServerError, "could not reach signal api.")
		return
	}
	defer res.Body.Close()
	fmt.Println("signal response:", res.Status)
}
