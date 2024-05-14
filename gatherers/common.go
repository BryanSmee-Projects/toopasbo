package gatherers

type Weather struct {
	Location           string
	CurrentTemperature int
	MinTemperature     int
	MaxTemperature     int
	Summary            string
	WindSpeed          int
	Description        string
}

type GeoPosition struct {
	lat  float64
	lon  float64
	name string
}

func GenerateDebugWeather() Weather {
	return Weather{
		Location:           "Paris",
		CurrentTemperature: 20,
		MinTemperature:     15,
		MaxTemperature:     25,
		Summary:            "Sunny",
		WindSpeed:          5,
		Description:        "Sunny",
	}
}
