package llm

import (
	"context"
	"errors"
	"strings"
	"testing"

	"github.com/tmc/langchaingo/llms"
)

// repoStub is a stub for ClothingQuerier used in ClothesTool tests.
type repoStub struct {
	item ClothingItem
	err  error
}

func (r *repoStub) OneByCategoryStyleWeather(_ context.Context, _, _, _ string) (ClothingItem, error) {
	return r.item, r.err
}

func TestClothesTool_Name(t *testing.T) {
	t.Parallel()
	tool := &ClothesTool{repo: &repoStub{}}
	if tool.Name() != "search_clothing" {
		t.Fatalf("expected name %q, got %q", "search_clothing", tool.Name())
	}
}

func TestClothesTool_Call(t *testing.T) {
	t.Parallel()

	foundItem := ClothingItem{
		ID:       1,
		Name:     "White T-Shirt",
		Price:    19.99,
		Color:    "white",
		Category: "tops",
		Style:    "casual",
		Weather:  "hot",
	}

	tests := []struct {
		name        string
		input       string
		repo        *repoStub
		wantContain string
		wantErr     bool
	}{
		{
			name:        "returns JSON item on successful lookup",
			input:       `{"category":"tops","style":"casual","weather":"hot"}`,
			repo:        &repoStub{item: foundItem},
			wantContain: `"name":"White T-Shirt"`,
		},
		{
			name:        "returns error string when item not found",
			input:       `{"category":"tops","style":"casual","weather":"hot"}`,
			repo:        &repoStub{err: ErrNotFound},
			wantContain: "error: no item found",
		},
		{
			name:        "returns error string for database error",
			input:       `{"category":"tops","style":"casual","weather":"hot"}`,
			repo:        &repoStub{err: errors.New("db connection lost")},
			wantContain: "error: database error",
		},
		{
			name:        "returns error string for malformed JSON",
			input:       `not-json`,
			repo:        &repoStub{},
			wantContain: "error: invalid JSON input",
		},
		{
			name:        "returns error string when category is missing",
			input:       `{"style":"casual","weather":"hot"}`,
			repo:        &repoStub{},
			wantContain: "error: category, style, and weather are all required",
		},
		{
			name:        "returns error string when style is missing",
			input:       `{"category":"tops","weather":"hot"}`,
			repo:        &repoStub{},
			wantContain: "error: category, style, and weather are all required",
		},
		{
			name:        "returns error string when weather is missing",
			input:       `{"category":"tops","style":"casual"}`,
			repo:        &repoStub{},
			wantContain: "error: category, style, and weather are all required",
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			tool := &ClothesTool{repo: tt.repo}
			got, err := tool.Call(context.Background(), tt.input)

			if tt.wantErr {
				if err == nil {
					t.Fatal("expected error, got nil")
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if !strings.Contains(got, tt.wantContain) {
				t.Fatalf("expected output to contain %q, got %q", tt.wantContain, got)
			}
		})
	}
}

func TestExtractJSONArray(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		in   string
		want string
	}{
		{
			name: "extracts bare JSON array",
			in:   `[{"id":1},{"id":2}]`,
			want: `[{"id":1},{"id":2}]`,
		},
		{
			name: "extracts array from surrounding text",
			in:   `Final Answer: [{"id":1},{"id":2}]`,
			want: `[{"id":1},{"id":2}]`,
		},
		{
			name: "returns raw when no array found",
			in:   "not-json",
			want: "not-json",
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			got := extractJSONArray(tt.in)
			if got != tt.want {
				t.Fatalf("expected %q, got %q", tt.want, got)
			}
		})
	}
}

// fakeModel simulates the LLM for agent tests.
type fakeModel struct {
	responses []string
	idx       int
	err       error
}

func (f *fakeModel) GenerateContent(_ context.Context, _ []llms.MessageContent, _ ...llms.CallOption) (*llms.ContentResponse, error) {
	if f.err != nil {
		return nil, f.err
	}
	if f.idx >= len(f.responses) {
		return nil, errors.New("unexpected call: no more responses")
	}
	resp := f.responses[f.idx]
	f.idx++
	return &llms.ContentResponse{
		Choices: []*llms.ContentChoice{{Content: resp}},
	}, nil
}

func (f *fakeModel) Call(_ context.Context, _ string, _ ...llms.CallOption) (string, error) {
	if f.err != nil {
		return "", f.err
	}
	if f.idx >= len(f.responses) {
		return "", errors.New("unexpected call: no more responses")
	}
	resp := f.responses[f.idx]
	f.idx++
	return resp, nil
}

func TestClothingAgent_Recommend(t *testing.T) {
	t.Parallel()

	validFinalAnswer := `Thought: I now know the final answer.
Final Answer: [{"id":1,"name":"T-Shirt","price":19.99,"color":"white","category":"tops","style":"casual","weather":"hot"},{"id":2,"name":"Jeans","price":49.99,"color":"blue","category":"bottoms","style":"casual","weather":"hot"},{"id":3,"name":"Sneakers","price":79.99,"color":"white","category":"shoes","style":"casual","weather":"hot"}]`

	tests := []struct {
		name      string
		model     *fakeModel
		repo      *repoStub
		weather   WeatherClient
		wantItems int
		wantErr   string
	}{
		{
			name:      "parses valid JSON array from Final Answer",
			model:     &fakeModel{responses: []string{validFinalAnswer}},
			repo:      &repoStub{},
			weather:   &weatherClientStub{},
			wantItems: 3,
		},
		{
			name:    "returns error when model fails",
			model:   &fakeModel{err: errors.New("llm unavailable")},
			repo:    &repoStub{},
			weather: &weatherClientStub{},
			wantErr: "agent run",
		},
		{
			name:    "returns error when Final Answer contains invalid JSON",
			model:   &fakeModel{responses: []string{"Thought: done.\nFinal Answer: not-a-json-array"}},
			repo:    &repoStub{},
			weather: &weatherClientStub{},
			wantErr: "parse agent result",
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			agent, err := newClothingAgentWithModel(tt.model, tt.repo, tt.weather)
			if err != nil {
				t.Fatalf("unexpected error building agent: %v", err)
			}

			got, err := agent.Recommend(context.Background(), "casual clothes for a hot day")

			if tt.wantErr != "" {
				if err == nil {
					t.Fatalf("expected error containing %q, got nil", tt.wantErr)
				}
				if !strings.Contains(err.Error(), tt.wantErr) {
					t.Fatalf("expected error %q to contain %q", err.Error(), tt.wantErr)
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if len(got) != tt.wantItems {
				t.Fatalf("expected %d items, got %d", tt.wantItems, len(got))
			}
		})
	}
}
