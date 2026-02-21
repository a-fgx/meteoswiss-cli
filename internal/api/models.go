package api

// PLZDetail is the response from the MeteoSwiss app API for a given Swiss postal code.
type PLZDetail struct {
	CurrentWeather CurrentWeather `json:"currentWeather"`
	Forecast       []DayForecast  `json:"forecast"`
	Warnings       []Warning      `json:"warnings"`
	Graph          *GraphData     `json:"graph,omitempty"`
}

// CurrentWeather holds the current observed conditions.
type CurrentWeather struct {
	Time        int64   `json:"time"`
	Icon        int     `json:"icon"`
	Temperature float64 `json:"temperature"`
}

// DayForecast holds a single day's forecast data.
type DayForecast struct {
	DayDate          string  `json:"dayDate"`
	IconDay          int     `json:"iconDay"`
	TemperatureMax   float64 `json:"temperatureMax"`
	TemperatureMin   float64 `json:"temperatureMin"`
	Precipitation    float64 `json:"precipitation"`
	PrecipitationMin float64 `json:"precipitationMin"`
	PrecipitationMax float64 `json:"precipitationMax"`
}

// GraphData holds precipitation data for the rain command.
// High-resolution (10-min) data starts at Start; low-resolution (1-hour)
// data starts at StartLowResolution. All timestamps are Unix milliseconds.
type GraphData struct {
	Start              int64     `json:"start"`
	StartLowResolution int64     `json:"startLowResolution"`
	Precipitation10m   []float64 `json:"precipitation10m"`
	Precipitation1h    []float64 `json:"precipitation1h"`
}

// Warning represents a MeteoSwiss weather warning.
type Warning struct {
	WarnType    int      `json:"warnType"`
	WarnLevel   int      `json:"warnLevel"`
	ValidFrom   string   `json:"validFrom"`
	ValidTo     string   `json:"validTo"`
	Regions     []string `json:"regions"`
	Headline    string   `json:"headline"`
	Body        string   `json:"body"`
}

// WarnType maps warning type IDs to human-readable names.
var WarnType = map[int]string{
	0:  "Wind",
	1:  "Thunderstorm",
	2:  "Rain",
	3:  "Snow",
	4:  "Slippery roads",
	5:  "Frost",
	6:  "Heat",
	7:  "Avalanche",
	8:  "Fire danger",
	9:  "Flooding",
	10: "UV",
}

// WarnLevel maps warning level IDs to human-readable severity.
var WarnLevel = map[int]string{
	1: "Minor",
	2: "Moderate",
	3: "Considerable",
	4: "High",
	5: "Very high",
}

// WeatherIcon maps MeteoSwiss icon codes (1–42) to a short text description
// and a Unicode emoji approximation for terminal display.
var WeatherIcon = map[int][2]string{
	1:  {"Sunny", "☀️"},
	2:  {"Mostly sunny", "🌤️"},
	3:  {"Partly cloudy", "⛅"},
	4:  {"Mostly cloudy", "🌥️"},
	5:  {"Overcast", "☁️"},
	6:  {"Fog", "🌫️"},
	7:  {"Light rain showers", "🌦️"},
	8:  {"Rain showers", "🌧️"},
	9:  {"Heavy rain showers", "🌧️"},
	10: {"Thunderstorm", "⛈️"},
	11: {"Light snowfall", "🌨️"},
	12: {"Snowfall", "❄️"},
	13: {"Heavy snowfall", "❄️"},
	14: {"Sleet", "🌨️"},
	15: {"Freezing rain", "🌧️"},
	16: {"Clear night", "🌙"},
	17: {"Mostly clear night", "🌙"},
	18: {"Partly cloudy night", "🌙"},
	19: {"Mostly cloudy night", "☁️"},
	20: {"Fog night", "🌫️"},
	21: {"Light rain showers night", "🌧️"},
	22: {"Rain showers night", "🌧️"},
	23: {"Heavy rain showers night", "🌧️"},
	24: {"Thunderstorm night", "⛈️"},
	25: {"Light snowfall night", "🌨️"},
	26: {"Snowfall night", "❄️"},
	27: {"Heavy snowfall night", "❄️"},
	28: {"Sleet night", "🌨️"},
	29: {"Freezing rain night", "🌧️"},
	30: {"Sunny intervals", "🌤️"},
	31: {"Mostly sunny intervals", "🌤️"},
	32: {"Light drizzle", "🌦️"},
	33: {"Drizzle", "🌧️"},
	34: {"Light rain", "🌦️"},
	35: {"Rain", "🌧️"},
	36: {"Heavy rain", "🌧️"},
	37: {"Hail", "⛈️"},
	38: {"Light snow", "🌨️"},
	39: {"Snow", "❄️"},
	40: {"Heavy snow", "❄️"},
	41: {"Thunderstorm with hail", "⛈️"},
	42: {"Blowing snow", "❄️"},
}

// IconDescription returns a short text label for a weather icon code.
func IconDescription(code int) string {
	if v, ok := WeatherIcon[code]; ok {
		return v[0]
	}
	return "Unknown"
}

// IconEmoji returns an emoji for a weather icon code.
func IconEmoji(code int) string {
	if v, ok := WeatherIcon[code]; ok {
		return v[1]
	}
	return "?"
}

// WindDirection converts degrees to a cardinal direction string.
func WindDirectionLabel(deg int) string {
	dirs := []string{"N", "NNE", "NE", "ENE", "E", "ESE", "SE", "SSE", "S", "SSW", "SW", "WSW", "W", "WNW", "NW", "NNW"}
	if deg < 0 {
		return "—"
	}
	idx := int((float64(deg)+11.25)/22.5) % 16
	return dirs[idx]
}
