package endpoints

import (
	"io"
	"log"
	"net/http"
	"os"
)

func contains(slice []string, str string) bool {
	for _, s := range slice {
		if s == str {
			return true
		}
	}
	return false
}

func EnsureEnvVars(vars []string) {
	for _, v := range vars {
		if os.Getenv(v) == "" {
			log.Fatalf("%s env var is required", v)
		}
	}
}

func DownloadFile(url string) (string, error) {
	resp, httpError := http.Get(url)
	if httpError != nil {
		return "", httpError
	}

	defer resp.Body.Close()

	filedir, fsErr := os.MkdirTemp("", "toopasbo-")
	if fsErr != nil {
		return "", fsErr
	}

	file, fileErr := os.CreateTemp(filedir, "dalle-*.png")

	if fileErr != nil {
		return "", fileErr
	}

	defer file.Close()

	filepath := file.Name()

	_, copyErr := io.Copy(file, resp.Body)
	if copyErr != nil {
		return "", copyErr
	}

	return filepath, nil

}
