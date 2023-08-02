package config

import (
	"encoding/json"
	"log"
	"os"

	"gopkg.in/yaml.v3"
)

type ConfigInterface interface {}

type ConfigService struct {
	Config	*ConfigData
}

type ConfigData struct {
	Signal				SignalConfig			`yaml:"signal"`
	Server 				ServerConfig			`yaml:"server"`
	AMConfig			AlertmanagerConfig		`yaml:"alertmanager"`
	Recipients			map[string]interface{}	`yaml:"recipients"`
}

type SignalConfig struct {
	Number				string					`yaml:"number"`
	Recipients			[]string				`yaml:"recipients"`
	Send				string					`yaml:"send"`
}

type ServerConfig struct {
	Port				string					`yaml:"port"`
	Debug				bool					`yaml:"debug"`
}

type AlertmanagerConfig struct {
	IgnoreLabels		[]string				`yaml:"ignoreLabels"`
	IgnoreAnnotations	[]string				`yaml:"ignoreAnnotations"`
	GeneratorURL		bool					`yaml:"generatorURL"`
	MatchLabel			string					`yaml:"matchLabel"`
}

var _ ConfigInterface = &ConfigData{}

func NewConfig(configPath string) (*ConfigService, error) {
	var cfg *ConfigData

	file, err := os.ReadFile(configPath)
	if err != nil {
		log.Println("error opening config file", configPath, err)
		return nil, err
	}

	err = yaml.Unmarshal(file, &cfg)
	if err != nil {
		log.Println("error parsing yaml, trying json", err)
		err = json.Unmarshal([]byte(file), &cfg)
		if err != nil {
			log.Println("could not parse config yaml or json")
			return nil, err
		}
	}

	return &ConfigService{
		Config: cfg,
	}, nil
}
