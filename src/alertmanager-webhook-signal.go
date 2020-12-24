package main

import (
	"fmt"
	"log"
	"net/http"
	"encoding/json"
	"github.com/gorilla/mux"
	"io/ioutil"
	"bytes"
	"os"
	"gopkg.in/yaml.v2"
)

type Config struct {
    Signal struct {
        Number string `yaml:"number"`
		Recipients []string `yaml:"recipients"`
		Send string `yaml:"send"`
		IgnoreLabels []string `yaml:"ignoreLabels"`
		IgnoreAnnotations []string `yaml:"ignoreAnnotations"`
		GeneratorURL bool `yaml:"generatorURL"`
	} `yaml:"signal"`
	Server struct {
		Port string `yaml:"port"`
	}
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
	TruncatedAlerts		int 					`json:"truncatedAlerts"`
	Status				string					`json:"status"`
	Receiver			string					`json:"receiver"`
	GroupLabels			map[string]interface{}	`json:"groupLabels"`
	CommonLabels		map[string]interface{}	`json:"commonLabels"`
	CommonAnnotations	map[string]interface{}	`json:"commonAnnotations"`
	ExternalURL			string					`json:"externalURL"`
	Alerts				[]Alert 				`json:"alerts"`
}

type Alert struct {
	Status			string						`json:"status"`
	Labels			map[string]interface{}		`json:"labels"`
	Annotations		map[string]interface{}		`json:"annotations"`
	StartsAt		string						`json:"startsAt"`
	EndsAt			string						`json:"endsAt"`
	GeneratorURL	string						`json:"generatorURL"`
}

type SignalMessage struct {
	Attachments		[]string	`json:"base64_attachments"`
	Message			string		`json:"message"`
	Number			string		`json:"number"`
	Recipients 		[]string	`json:"recipients"`
}

var cfg, _ = NewConfig("./config.yaml")

func main() {
	checkConfig()
	handleRequests()
}

func checkConfig() {
	if cfg.Server.Port == "" {
		log.Fatal("Server Port not set.")
	}
	if cfg.Signal.Number == "" {
		log.Fatal("Signal Number not set.")
	}
	if len(cfg.Signal.Recipients) == 0 {
		log.Fatal("Signal Recipients not set.")
	}
	if cfg.Signal.Send == "" {
		log.Fatal("Signal URL not set.")
	}
}

func handleRequests() {
	fmt.Println("Starting server. Listening on port", cfg.Server.Port)
    router := mux.NewRouter().StrictSlash(true)
	router.HandleFunc("/api/v1/alert", alertReceived).Methods("POST")
	router.HandleFunc("/-/reload", reloadConfig).Methods("GET")
    log.Fatal(http.ListenAndServe(":"+cfg.Server.Port, router))
}

func reloadConfig(w http.ResponseWriter, r *http.Request) {
	cfg, _ = NewConfig("./config.yaml")
	fmt.Println("Reloading Config.")
}

func alertReceived(w http.ResponseWriter, r *http.Request) {
	reqBody, _ := ioutil.ReadAll(r.Body)
	var Alert Alertmanager
	json.Unmarshal(reqBody, &Alert)
	mapSignal(Alert)
}

func mapSignal(a Alertmanager) {
	for _, element := range a.Alerts {
		_message := "Alert " + fmt.Sprint(element.Labels["alertname"]) + " is " + element.Status
		for k, v := range element.Annotations {
			if !stringInSlice(k, cfg.Signal.IgnoreAnnotations) {
				_message += fmt.Sprintf("\n%v: %v", k, v)
			}
		}
		for k, v := range element.Labels {
			if !stringInSlice(k, cfg.Signal.IgnoreLabels) {
				_message += fmt.Sprintf("\n%v: %v", k, v)
			}
		}
		if cfg.Signal.GeneratorURL {
			_message += fmt.Sprintf("\nuri: %v", element.GeneratorURL)
		}
		_signal := SignalMessage{Message: _message, Number: cfg.Signal.Number, Recipients: cfg.Signal.Recipients} 
		fmt.Println("sending alert:", element.Labels["alertname"])
		sendSignal(_signal)
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

func sendSignal(m SignalMessage) {
	payloadBuf := new(bytes.Buffer)
	json.NewEncoder(payloadBuf).Encode(m)
	req, _ := http.NewRequest("POST", cfg.Signal.Send, payloadBuf)
	client := &http.Client{}
	res, e := client.Do(req)
	if e != nil {
		fmt.Println(e)
	}
	defer res.Body.Close()
	fmt.Println("signal response:", res.Status)
}
