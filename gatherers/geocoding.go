package gatherers

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"os"
)

type OpenWeatherMapResponse struct {
	Lat  float64 `json:"lat"`
	Lon  float64 `json:"lon"`
	Name string  `json:"name"`
}

var AllowedCountries = []string{"FR", "CZ"}
var OpenWeatherAPIKey = os.Getenv("OPENWEATHER_API_KEY")
var OpenWeatherURL = "http://api.openweathermap.org/geo/1.0/zip?zip=%s,%s&appid=%s"

func isCountryAllowed(countryCode string) bool {
	for _, c := range AllowedCountries {
		if c == countryCode {
			return true
		}
	}
	return false
}

func GetCityPosition(cityName string) (GeoPosition, error) {
	url := fmt.Sprintf("http://api.openweathermap.org/geo/1.0/direct?q=%s&limit=1&appid=%s", cityName, OpenWeatherAPIKey)
	resp, err := http.Get(url)
	if err != nil {
		fmt.Println("Error getting position")
		return GeoPosition{}, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		return GeoPosition{}, errors.New("error getting position, got status code " + resp.Status)
	}

	var openWeatherResponse []OpenWeatherMapResponse
	err = json.NewDecoder(resp.Body).Decode(&openWeatherResponse)

	if err != nil {
		fmt.Println("Error decoding position")
		return GeoPosition{}, err
	}

	return GeoPosition{lat: openWeatherResponse[0].Lat, lon: openWeatherResponse[0].Lon, name: openWeatherResponse[0].Name}, nil

}

func GetPosition(zipCode string, countryCode string) (GeoPosition, error) {
	if !isCountryAllowed(countryCode) {
		fmt.Println("Country not allowed")
		return GeoPosition{}, errors.New("country not allowed")
	}

	url := fmt.Sprintf(OpenWeatherURL, zipCode, countryCode, OpenWeatherAPIKey)
	resp, err := http.Get(url)
	if err != nil {
		fmt.Println("Error getting position")
		return GeoPosition{}, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		return GeoPosition{}, errors.New("error getting position, got status code " + resp.Status)
	}

	var openWeatherResponse OpenWeatherMapResponse
	err = json.NewDecoder(resp.Body).Decode(&openWeatherResponse)

	if err != nil {
		fmt.Println("Error decoding position")
		return GeoPosition{}, err
	}

	return GeoPosition{lat: openWeatherResponse.Lat, lon: openWeatherResponse.Lon, name: openWeatherResponse.Name}, nil
}
