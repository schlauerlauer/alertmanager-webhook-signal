package util

import (
	"encoding/base64"
	"errors"
	"io"
	"net/http"
)

func StringInSlice(compare string, list []string) bool {
	for _, element := range list {
		if element == compare {
			return true
		}
	}
	return false
}

func GetImage(url string) (string, error) {
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

func Ternary(condition bool, truthy string, falsy string) string {
	if condition {
		return truthy
	}
	return falsy
}

func StringDefault(str, def string) string {
	if str == "" {
		return def
	}
	return str
}
