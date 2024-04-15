package interfaces

import (
	"encoding/base64"
	"errors"
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
