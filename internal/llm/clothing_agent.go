package llm

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strings"

	"github.com/tmc/langchaingo/agents"
	"github.com/tmc/langchaingo/chains"
	"github.com/tmc/langchaingo/llms"
	"github.com/tmc/langchaingo/llms/openai"
	lctools "github.com/tmc/langchaingo/tools"
)

var ErrNotFound = errors.New("not found")

type ClothingItem struct {
	ID       int64   `json:"id"`
	Name     string  `json:"name"`
	Price    float64 `json:"price"`
	Color    string  `json:"color"`
	Category string  `json:"category"`
	Style    string  `json:"style"`
	Weather  string  `json:"weather"`
}

type Recommender interface {
	Recommend(ctx context.Context, prompt string) ([]ClothingItem, error)
}

type ClothingQuerier interface {
	OneByCategoryStyleWeather(ctx context.Context, category, style, weather string) (ClothingItem, error)
}

type ClothesTool struct {
	repo ClothingQuerier
}

func (t *ClothesTool) Name() string { return "search_clothing" }

func (t *ClothesTool) Description() string {
	return `Searches the clothing database for an item matching a specific category, style, and weather.
Input must be a JSON object with three required fields:
- "category": one of "tops", "bottoms", "shoes"
- "style": one of "casual", "formal", "business_casual", "streetwear", "athleisure"
- "weather": one of "cold", "hot", "rainy", "snowy", "windy", "humid"
Returns a JSON object of the matching clothing item, or an error message if not found.
If not found, try a different style or weather value.`
}

type toolInput struct {
	Category string `json:"category"`
	Style    string `json:"style"`
	Weather  string `json:"weather"`
}

func (t *ClothesTool) Call(ctx context.Context, input string) (string, error) {
	var in toolInput
	if err := json.Unmarshal([]byte(input), &in); err != nil {
		return fmt.Sprintf("error: invalid JSON input: %v", err), nil
	}
	if in.Category == "" || in.Style == "" || in.Weather == "" {
		return "error: category, style, and weather are all required", nil
	}

	item, err := t.repo.OneByCategoryStyleWeather(ctx, in.Category, in.Style, in.Weather)
	if err != nil {
		if errors.Is(err, ErrNotFound) {
			return fmt.Sprintf("error: no item found for category=%s style=%s weather=%s; try different values", in.Category, in.Style, in.Weather), nil
		}
		return fmt.Sprintf("error: database error: %v", err), nil
	}

	out, err := json.Marshal(item)
	if err != nil {
		return fmt.Sprintf("error: marshal item: %v", err), nil
	}
	return string(out), nil
}

const agentSystemPrompt = `You are a clothing recommendation assistant.
Given a user's request, find one clothing item for each of three categories: tops, bottoms, and shoes.
If the user mentions a city or location, use the get_weather tool first to determine the current weather condition, then use that condition when searching for clothing.
Use the search_clothing tool for each category.
Always infer the most appropriate style and weather from the user's request.
If no style or weather is mentioned, pick reasonable defaults (e.g. style="casual", weather="hot").
If a combination is not found, try a different style or weather value.
When you have found all three items, return your Final Answer as a JSON array containing the three item objects.`

type ClothingAgent struct {
	executor *agents.Executor
}

func NewClothingAgent(apiKey, model string, repo ClothingQuerier, httpClient *http.Client) (*ClothingAgent, error) {
	client, err := openai.New(
		openai.WithToken(apiKey),
		openai.WithModel(model),
		openai.WithBaseURL("https://openrouter.ai/api/v1"),
	)
	if err != nil {
		return nil, fmt.Errorf("create openrouter model: %w", err)
	}
	weatherClient := &OpenMeteoClient{httpClient: httpClient}
	return newClothingAgentWithModel(client, repo, weatherClient)
}

func newClothingAgentWithModel(model llms.Model, repo ClothingQuerier, weather WeatherClient) (*ClothingAgent, error) {
	agentTools := []lctools.Tool{
		&ClothesTool{repo: repo},
		&WeatherTool{client: weather},
	}
	executor, err := agents.Initialize(
		model,
		agentTools,
		agents.ZeroShotReactDescription,
		agents.WithMaxIterations(12),
	)
	if err != nil {
		return nil, fmt.Errorf("initialize agent: %w", err)
	}
	return &ClothingAgent{executor: executor}, nil
}

func (a *ClothingAgent) Recommend(ctx context.Context, prompt string) ([]ClothingItem, error) {
	fullPrompt := agentSystemPrompt + "\n\nUser request: " + prompt
	result, err := chains.Run(ctx, a.executor, fullPrompt)
	if err != nil {
		return nil, fmt.Errorf("agent run: %w", err)
	}

	raw := extractJSONArray(result)
	var items []ClothingItem
	if err := json.Unmarshal([]byte(raw), &items); err != nil {
		return nil, fmt.Errorf("parse agent result: %w", err)
	}
	return items, nil
}

func extractJSONArray(raw string) string {
	start := strings.Index(raw, "[")
	end := strings.LastIndex(raw, "]")
	if start >= 0 && end > start {
		return raw[start : end+1]
	}
	return raw
}
