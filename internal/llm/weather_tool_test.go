package llm

import (
	"context"
	"errors"
	"strings"
	"testing"
)

type weatherClientStub struct {
	result WeatherResult
	err    error
}

func (s *weatherClientStub) GetWeather(_ context.Context, _ string) (WeatherResult, error) {
	return s.result, s.err
}

func TestWeatherTool_Name(t *testing.T) {
	t.Parallel()
	tool := &WeatherTool{client: &weatherClientStub{}}
	if tool.Name() != "get_weather" {
		t.Fatalf("expected name %q, got %q", "get_weather", tool.Name())
	}
}

func TestWeatherTool_Call(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name        string
		input       string
		stub        *weatherClientStub
		wantContain string
	}{
		{
			name:        "hot weather: high temp, clear sky",
			input:       "Buenos Aires",
			stub:        &weatherClientStub{result: WeatherResult{Temperature: 30, WeatherCode: 0, WindSpeed: 10}},
			wantContain: "hot",
		},
		{
			name:        "cold weather: low temp, clear sky",
			input:       "Ushuaia",
			stub:        &weatherClientStub{result: WeatherResult{Temperature: 5, WeatherCode: 0, WindSpeed: 5}},
			wantContain: "cold",
		},
		{
			name:        "rainy weather: wmo rain code",
			input:       "London",
			stub:        &weatherClientStub{result: WeatherResult{Temperature: 18, WeatherCode: 61, WindSpeed: 15}},
			wantContain: "rainy",
		},
		{
			name:        "snowy weather: wmo snow code",
			input:       "Oslo",
			stub:        &weatherClientStub{result: WeatherResult{Temperature: -2, WeatherCode: 71, WindSpeed: 10}},
			wantContain: "snowy",
		},
		{
			name:        "windy weather: high wind, no precipitation",
			input:       "Wellington",
			stub:        &weatherClientStub{result: WeatherResult{Temperature: 20, WeatherCode: 2, WindSpeed: 55}},
			wantContain: "windy",
		},
		{
			name:        "humid weather: mild temp, low wind, no precipitation",
			input:       "Singapore",
			stub:        &weatherClientStub{result: WeatherResult{Temperature: 22, WeatherCode: 2, WindSpeed: 10}},
			wantContain: "humid",
		},
		{
			name:        "rainy overrides high wind",
			input:       "SomeCity",
			stub:        &weatherClientStub{result: WeatherResult{Temperature: 20, WeatherCode: 80, WindSpeed: 60}},
			wantContain: "rainy",
		},
		{
			name:        "snowy overrides cold and wind",
			input:       "SomeCity",
			stub:        &weatherClientStub{result: WeatherResult{Temperature: -5, WeatherCode: 85, WindSpeed: 55}},
			wantContain: "snowy",
		},
		{
			name:        "empty city returns error string",
			input:       "",
			stub:        &weatherClientStub{},
			wantContain: "error: city is required",
		},
		{
			name:        "whitespace-only city returns error string",
			input:       "   ",
			stub:        &weatherClientStub{},
			wantContain: "error: city is required",
		},
		{
			name:        "client error returns error string",
			input:       "Atlantis",
			stub:        &weatherClientStub{err: errors.New("city not found")},
			wantContain: "error:",
		},
		{
			name:        "thunderstorm code maps to rainy",
			input:       "Miami",
			stub:        &weatherClientStub{result: WeatherResult{Temperature: 28, WeatherCode: 95, WindSpeed: 20}},
			wantContain: "rainy",
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			tool := &WeatherTool{client: tt.stub}
			got, err := tool.Call(context.Background(), tt.input)

			if err != nil {
				t.Fatalf("unexpected error from Call: %v", err)
			}
			if !strings.Contains(got, tt.wantContain) {
				t.Fatalf("expected output to contain %q, got %q", tt.wantContain, got)
			}
		})
	}
}

func TestWeatherConditionFromResult(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		r    WeatherResult
		want string
	}{
		{"hot clear", WeatherResult{Temperature: 28, WeatherCode: 0, WindSpeed: 5}, "hot"},
		{"cold clear", WeatherResult{Temperature: 9, WeatherCode: 1, WindSpeed: 5}, "cold"},
		{"rainy code 61", WeatherResult{Temperature: 15, WeatherCode: 61, WindSpeed: 10}, "rainy"},
		{"rainy code 82", WeatherResult{Temperature: 20, WeatherCode: 82, WindSpeed: 10}, "rainy"},
		{"rainy code 99", WeatherResult{Temperature: 25, WeatherCode: 99, WindSpeed: 10}, "rainy"},
		{"snowy code 71", WeatherResult{Temperature: 0, WeatherCode: 71, WindSpeed: 5}, "snowy"},
		{"snowy code 86", WeatherResult{Temperature: -3, WeatherCode: 86, WindSpeed: 5}, "snowy"},
		{"windy no precip", WeatherResult{Temperature: 18, WeatherCode: 3, WindSpeed: 41}, "windy"},
		{"humid default", WeatherResult{Temperature: 22, WeatherCode: 3, WindSpeed: 15}, "humid"},
		{"boundary 27C is hot", WeatherResult{Temperature: 27.1, WeatherCode: 0, WindSpeed: 5}, "hot"},
		{"boundary 10C is cold", WeatherResult{Temperature: 9.9, WeatherCode: 0, WindSpeed: 5}, "cold"},
		{"boundary 10C exactly is humid", WeatherResult{Temperature: 10, WeatherCode: 0, WindSpeed: 5}, "humid"},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			got := weatherConditionFromResult(tt.r)
			if got != tt.want {
				t.Fatalf("expected %q, got %q", tt.want, got)
			}
		})
	}
}
