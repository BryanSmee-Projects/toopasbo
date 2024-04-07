package gatherers

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"os"
)

type OpenWeatherResponse struct {
	Current struct {
		Temperature float64 `json:"temp"`
		FeelsLike   float64 `json:"feels_like"`
		WindSpeed   float64 `json:"wind_speed"`
		Weather     []struct {
			Description string `json:"description"`
		} `json:"weather"`
	} `json:"current"`
	Daily []struct {
		Summary     string `json:"summary"`
		Temperature struct {
			Day float64 `json:"day"`
			Min float64 `json:"min"`
			Max float64 `json:"max"`
		} `json:"temp"`
		Weather []struct {
			Description string `json:"description"`
		} `json:"weather"`
		WindSpeed float64 `json:"wind_speed"`
	} `json:"daily"`
}

var apiKey string = os.Getenv("OPENWEATHER_API_KEY")
var defaultUrl string = "https://api.openweathermap.org/data/3.0/onecall?lat=%s&lon=%s&units=metric&exclude=minutely,hourly&appid=%s"

func buildUrl(lat, lon string) string {
	return fmt.Sprintf(defaultUrl, lat, lon, apiKey)
}

func GetWeather(position GeoPosition) (Weather, error) {
	latStr := fmt.Sprintf("%f", position.lat)
	lonStr := fmt.Sprintf("%f", position.lon)
	url := buildUrl(latStr, lonStr)
	resp, err := http.Get(url)
	if err != nil {
		return Weather{}, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		return Weather{}, errors.New("error getting weather, got status code " + resp.Status)
	}

	var openWeatherResponse OpenWeatherResponse
	err = json.NewDecoder(resp.Body).Decode(&openWeatherResponse)

	if err != nil {
		return Weather{}, err
	}

	return Weather{
		CurrentTemperature: int(openWeatherResponse.Current.Temperature),
		MinTemperature:     int(openWeatherResponse.Daily[0].Temperature.Min),
		MaxTemperature:     int(openWeatherResponse.Daily[0].Temperature.Max),
		Summary:            openWeatherResponse.Daily[0].Summary,
		WindSpeed:          int(openWeatherResponse.Current.WindSpeed),
		Description:        openWeatherResponse.Current.Weather[0].Description,
	}, nil
}

func GetWeatherForWeek(position GeoPosition) ([]Weather, error) {
	latStr := fmt.Sprintf("%f", position.lat)
	lonStr := fmt.Sprintf("%f", position.lon)
	url := buildUrl(latStr, lonStr)
	resp, err := http.Get(url)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		return nil, errors.New("error getting weather, got status code " + resp.Status)
	}

	var openWeatherResponse OpenWeatherResponse
	err = json.NewDecoder(resp.Body).Decode(&openWeatherResponse)

	if err != nil {
		return nil, err
	}

	var weather []Weather
	for _, day := range openWeatherResponse.Daily {
		weather = append(weather, Weather{
			CurrentTemperature: int(day.Temperature.Day),
			MinTemperature:     int(day.Temperature.Min),
			MaxTemperature:     int(day.Temperature.Max),
			Summary:            day.Summary,
			WindSpeed:          int(day.WindSpeed),
			Description:        day.Weather[0].Description,
		})
	}

	if len(weather) > 7 {
		weather = weather[:7]
	}

	return weather, nil
}
