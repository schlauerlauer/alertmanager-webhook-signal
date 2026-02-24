package config

import (
	"log/slog"
	"os"

	"github.com/goccy/go-yaml"
)

type ConfigData struct {
	Signal     SignalConfig       `yaml:"signal"`
	Server     ServerConfig       `yaml:"server"`
	AMConfig   AlertmanagerConfig `yaml:"alertmanager"`
	Recipients map[string]string  `yaml:"recipients"`
	Templates  TemplateConfig     `yaml:"templates"`
}

type SignalConfig struct {
	Number         string   `yaml:"number"`
	Recipients     []string `yaml:"recipients"`
	Send           string   `yaml:"send"`
	TextModeNormal bool     `yaml:"textmodeNormal"`
}

type ServerConfig struct {
	Interface string `yaml:"interface"`
	Port      string `yaml:"port"`
	Debug     bool   `yaml:"debug"`
}

type AlertmanagerConfig struct {
	IgnoreLabels      []string `yaml:"ignoreLabels"`
	IgnoreAnnotations []string `yaml:"ignoreAnnotations"`
	GeneratorURL      bool     `yaml:"generatorURL"`
}

type TemplateConfig struct {
	Grafana      string `yaml:"grafana"`
	Alertmanager string `yaml:"alertmanager"`
}

func NewConfig(configPath string) (*ConfigData, error) {
	var cfg ConfigData

	file, err := os.ReadFile(configPath)
	if err != nil {
		slog.Error("Error opening config file", "config_path", configPath, "err", err)
		return nil, err
	}

	err = yaml.Unmarshal(file, &cfg)
	if err != nil {
		slog.Error("Error parsing yaml, trying json", "err", err)
		return nil, err
	}

	return &cfg, nil
}
