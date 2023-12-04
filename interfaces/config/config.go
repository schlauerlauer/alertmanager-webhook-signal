package config

import (
	"encoding/json"
	"log/slog"
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
		slog.Error("Error opening config file", "config_path", configPath, "err", err)
		return nil, err
	}

	err = yaml.Unmarshal(file, &cfg)
	if err != nil {
		slog.Error("Error parsing yaml, trying json", "err", err)
		err = json.Unmarshal([]byte(file), &cfg)
		if err != nil {
			slog.Error("Could not parse config yaml or json", "err", err)
			return nil, err
		}
	}

	return &ConfigService{
		Config: cfg,
	}, nil
}
