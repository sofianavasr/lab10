package llm

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
)

type WeatherResult struct {
	Temperature float64
	WeatherCode int
	WindSpeed   float64
}

type WeatherClient interface {
	GetWeather(ctx context.Context, city string) (WeatherResult, error)
}

type WeatherTool struct {
	client WeatherClient
}

func NewWeatherTool(httpClient *http.Client) *WeatherTool {
	return &WeatherTool{client: &OpenMeteoClient{httpClient: httpClient}}
}

func (t *WeatherTool) Name() string { return "get_weather" }

func (t *WeatherTool) Description() string {
	return `Returns the current weather condition for a given city name.
Input: the city name as a plain string (e.g. "Buenos Aires").
Output: a short description including the weather condition (one of: hot, cold, rainy, snowy, windy, humid) and the temperature in Celsius.
Use the returned condition as the "weather" field when calling search_clothing.`
}

func (t *WeatherTool) Call(ctx context.Context, input string) (string, error) {
	city := strings.TrimSpace(input)
	if city == "" {
		return "error: city is required", nil
	}

	result, err := t.client.GetWeather(ctx, city)
	if err != nil {
		return fmt.Sprintf("error: %v", err), nil
	}

	condition := weatherConditionFromResult(result)
	return fmt.Sprintf("weather: %s, temperature: %.1f°C", condition, result.Temperature), nil
}

// weatherConditionFromResult maps raw weather data to one of the six clothing-database weather values.
// Priority order: snowy > rainy > windy > cold > hot > humid.
func weatherConditionFromResult(r WeatherResult) string {
	if isSnowy(r.WeatherCode) {
		return "snowy"
	}
	if isRainy(r.WeatherCode) {
		return "rainy"
	}
	if r.WindSpeed > 40 {
		return "windy"
	}
	if r.Temperature < 10 {
		return "cold"
	}
	if r.Temperature > 27 {
		return "hot"
	}
	return "humid"
}

// isRainy returns true for WMO codes that indicate rain or thunderstorms.
func isRainy(code int) bool {
	return (code >= 51 && code <= 67) ||
		(code >= 80 && code <= 82) ||
		(code >= 95 && code <= 99)
}

// isSnowy returns true for WMO codes that indicate snow.
func isSnowy(code int) bool {
	return (code >= 71 && code <= 77) ||
		code == 85 || code == 86
}

type OpenMeteoClient struct {
	httpClient *http.Client
}

type geocodingResponse struct {
	Results []struct {
		Latitude  float64 `json:"latitude"`
		Longitude float64 `json:"longitude"`
	} `json:"results"`
}

type forecastResponse struct {
	Current struct {
		Temperature2m float64 `json:"temperature_2m"`
		WeatherCode   int     `json:"weathercode"`
		WindSpeed10m  float64 `json:"wind_speed_10m"`
	} `json:"current"`
}

func (c *OpenMeteoClient) GetWeather(ctx context.Context, city string) (WeatherResult, error) {
	lat, lon, err := c.geocode(ctx, city)
	if err != nil {
		return WeatherResult{}, err
	}
	return c.fetchCurrentWeather(ctx, lat, lon)
}

func (c *OpenMeteoClient) geocode(ctx context.Context, city string) (lat, lon float64, err error) {
	q := url.Values{}
	q.Set("name", city)
	q.Set("count", "1")
	q.Set("language", "en")
	q.Set("format", "json")
	rawURL := "https://geocoding-api.open-meteo.com/v1/search?" + q.Encode()

	body, err := c.get(ctx, rawURL)
	if err != nil {
		return 0, 0, fmt.Errorf("geocoding request: %w", err)
	}

	var geo geocodingResponse
	if err := json.Unmarshal(body, &geo); err != nil {
		return 0, 0, fmt.Errorf("decode geocoding response: %w", err)
	}
	if len(geo.Results) == 0 {
		return 0, 0, fmt.Errorf("city not found: %s", city)
	}

	return geo.Results[0].Latitude, geo.Results[0].Longitude, nil
}

func (c *OpenMeteoClient) fetchCurrentWeather(ctx context.Context, lat, lon float64) (WeatherResult, error) {
	forecastURL := fmt.Sprintf(
		"https://api.open-meteo.com/v1/forecast?latitude=%f&longitude=%f&current=temperature_2m,weathercode,wind_speed_10m",
		lat, lon,
	)

	body, err := c.get(ctx, forecastURL)
	if err != nil {
		return WeatherResult{}, fmt.Errorf("forecast request: %w", err)
	}

	var f forecastResponse
	if err := json.Unmarshal(body, &f); err != nil {
		return WeatherResult{}, fmt.Errorf("decode forecast response: %w", err)
	}

	return WeatherResult{
		Temperature: f.Current.Temperature2m,
		WeatherCode: f.Current.WeatherCode,
		WindSpeed:   f.Current.WindSpeed10m,
	}, nil
}

func (c *OpenMeteoClient) get(ctx context.Context, url string) ([]byte, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, fmt.Errorf("build request: %w", err)
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("http get: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("unexpected status %d", resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("read response body: %w", err)
	}
	return body, nil
}
